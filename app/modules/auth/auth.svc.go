package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	entitiesdto "pichost.io/app/modules/entities/dto"
	"pichost.io/app/modules/entities/ent"
	entitiesinf "pichost.io/app/modules/entities/inf"
	mailerinf "pichost.io/app/modules/mailer/inf"
	"pichost.io/app/utils/hashing"
	"pichost.io/internal/config"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
)

type Service struct {
	tracer    trace.Tracer
	user      entitiesinf.UserEntity
	auth      entitiesinf.AuthEntity
	quotaEnt  entitiesinf.UserQuotaEntity
	planEnt   entitiesinf.PlanSettingEntity
	imageEnt  entitiesinf.ImageEntity
	mailer    mailerinf.Mailer
	activator EntitlementActivator
	conf      *config.Config[Config]
}

type Options struct {
	*config.Config[Config]
	tracer   trace.Tracer
	user     entitiesinf.UserEntity
	auth     entitiesinf.AuthEntity
	quotaEnt entitiesinf.UserQuotaEntity
	planEnt  entitiesinf.PlanSettingEntity
}

func newService(opt *Options) *Service {
	return &Service{
		tracer:   opt.tracer,
		user:     opt.user,
		auth:     opt.auth,
		quotaEnt: opt.quotaEnt,
		planEnt:  opt.planEnt,
		conf:     opt.Config,
	}
}

type RegisterRequestService struct {
	Email    string
	Password string
	Username string
}

type LoginRequestService struct {
	Email    string
	Password string
}

type AuthResultService struct {
	AccessToken  string
	AccessExpiry int
	RefreshToken string
	User         *ent.UserEntity
}

type accessClaims struct {
	Sub string `json:"sub"`
	Iss string `json:"iss"`
	Typ string `json:"typ"`
	Exp int64  `json:"exp"`
	Iat int64  `json:"iat"`
}

func (s *Service) Register(ctx context.Context, req RegisterRequestService, userAgent string, ip string) (*AuthResultService, error) {
	existing, err := s.user.GetUserByEmail(ctx, req.Email)
	if err == nil && existing != nil {
		return nil, ErrUserEmailAlreadyExists
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	hash, err := hashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	plan := "Free"
	created, err := s.user.CreateUser(ctx, entitiesdto.CreateUser{
		Email:    &req.Email,
		Password: &hash,
		Username: &req.Username,
		Plan:     plan,
		IsGuest:  false,
	})
	if err != nil {
		return nil, err
	}

	// Initialise quota row for the new user (best-effort — do not block auth).
	_, _ = s.quotaEnt.UpsertUserQuota(ctx, created.ID)

	// Send email verification asynchronously
	go func() {
		_ = s.SendEmailVerification(context.Background(), created)
	}()

	return s.issueAuth(ctx, created, userAgent, ip)
}

func (s *Service) Login(ctx context.Context, req LoginRequestService, userAgent string, ip string) (*AuthResultService, error) {
	user, err := s.user.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAuthInvalidCredentials
		}
		return nil, err
	}

	if user.Password == nil || !verifyPassword(*user.Password, req.Password) {
		return nil, ErrAuthInvalidCredentials
	}

	if !user.IsActive {
		return nil, ErrAuthUnauthorized
	}

	return s.issueAuth(ctx, user, userAgent, ip)
}

func (s *Service) issueAuth(ctx context.Context, user *ent.UserEntity, userAgent string, ip string) (*AuthResultService, error) {
	accessToken, err := s.signAccessToken(user.ID)
	if err != nil {
		return nil, err
	}

	refreshToken, refreshHash, err := newRefreshToken()
	if err != nil {
		return nil, err
	}

	ua := nullableString(userAgent)
	ipAddr := nullableString(ip)
	_, err = s.auth.CreateAuthSession(ctx, entitiesdto.CreateAuthSession{
		UserID:           user.ID,
		RefreshTokenHash: refreshHash,
		UserAgent:        ua,
		IPAddress:        ipAddr,
		ExpiresAt:        time.Now().Add(time.Duration(s.conf.Val.RefreshTokenTTLSeconds) * time.Second),
	})
	if err != nil {
		return nil, err
	}

	_ = s.user.RecordUserLogin(ctx, user.ID, ipAddr)

	return &AuthResultService{
		AccessToken:  accessToken,
		AccessExpiry: s.conf.Val.AccessTokenTTLSeconds,
		RefreshToken: refreshToken,
		User:         user,
	}, nil
}

func (s *Service) Refresh(ctx context.Context, rawRefreshToken string, userAgent string, ip string) (*AuthResultService, error) {
	if strings.TrimSpace(rawRefreshToken) == "" {
		return nil, ErrAuthInvalidRefreshToken
	}

	refreshHash := hashRefreshToken(rawRefreshToken)
	session, err := s.auth.GetAuthSessionByRefreshTokenHash(ctx, refreshHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAuthSessionNotFound
		}
		return nil, err
	}

	user, err := s.user.GetUserByID(ctx, session.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	if !user.IsActive {
		return nil, ErrAuthUnauthorized
	}

	newRefreshToken, newRefreshHash, err := newRefreshToken()
	if err != nil {
		return nil, err
	}

	_, err = s.auth.RotateAuthSession(ctx, session.ID, entitiesdto.RotateAuthSession{
		RefreshTokenHash: newRefreshHash,
		ExpiresAt:        time.Now().Add(time.Duration(s.conf.Val.RefreshTokenTTLSeconds) * time.Second),
	})
	if err != nil {
		return nil, err
	}

	accessToken, err := s.signAccessToken(user.ID)
	if err != nil {
		return nil, err
	}

	_ = userAgent
	_ = ip

	return &AuthResultService{
		AccessToken:  accessToken,
		AccessExpiry: s.conf.Val.AccessTokenTTLSeconds,
		RefreshToken: newRefreshToken,
		User:         user,
	}, nil
}

func (s *Service) LogoutByRefreshToken(ctx context.Context, rawRefreshToken string) error {
	if strings.TrimSpace(rawRefreshToken) == "" {
		return nil
	}

	refreshHash := hashRefreshToken(rawRefreshToken)
	session, err := s.auth.GetAuthSessionByRefreshTokenHash(ctx, refreshHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}

	return s.auth.RevokeAuthSession(ctx, session.ID)
}

func (s *Service) Me(ctx context.Context, userID uuid.UUID) (*ent.UserEntity, error) {
	user, err := s.user.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	// Lazy plan expiry: if the paid plan has expired, downgrade to Free automatically.
	if user.Plan != ent.PlanTypeFree && user.PlanExpiresAt != nil && time.Now().After(*user.PlanExpiresAt) {
		user, err = s.user.DowngradeUserPlanToFree(ctx, userID)
		if err != nil {
			return nil, err
		}
	}

	return user, nil
}

func (s *Service) ParseAccessToken(token string) (uuid.UUID, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return uuid.Nil, ErrAuthUnauthorized
	}

	signedData := parts[0] + "." + parts[1]
	expectedSig := s.signPayload(signedData)
	if !hmac.Equal([]byte(expectedSig), []byte(parts[2])) {
		return uuid.Nil, ErrAuthUnauthorized
	}

	payloadRaw, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return uuid.Nil, ErrAuthUnauthorized
	}

	var claims accessClaims
	if err := json.Unmarshal(payloadRaw, &claims); err != nil {
		return uuid.Nil, ErrAuthUnauthorized
	}

	if claims.Typ != "access" || claims.Iss != s.conf.Val.JWTIssuer || claims.Exp <= time.Now().Unix() {
		return uuid.Nil, ErrAuthUnauthorized
	}

	userID, err := uuid.Parse(claims.Sub)
	if err != nil {
		return uuid.Nil, ErrAuthUnauthorized
	}

	return userID, nil
}

func (s *Service) signAccessToken(userID uuid.UUID) (string, error) {
	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}
	claims := accessClaims{
		Sub: userID.String(),
		Iss: s.conf.Val.JWTIssuer,
		Typ: "access",
		Iat: time.Now().Unix(),
		Exp: time.Now().Add(time.Duration(s.conf.Val.AccessTokenTTLSeconds) * time.Second).Unix(),
	}

	headerBytes, err := json.Marshal(header)
	if err != nil {
		return "", err
	}
	payloadBytes, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	encodedHeader := base64.RawURLEncoding.EncodeToString(headerBytes)
	encodedPayload := base64.RawURLEncoding.EncodeToString(payloadBytes)
	signedData := encodedHeader + "." + encodedPayload
	sig := s.signPayload(signedData)

	return signedData + "." + sig, nil
}

func (s *Service) signPayload(payload string) string {
	h := hmac.New(sha256.New, []byte(s.conf.Val.JWTSecret))
	_, _ = h.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}

func newRefreshToken() (raw string, hash string, err error) {
	b := make([]byte, 48)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}
	raw = base64.RawURLEncoding.EncodeToString(b)
	hash = hashRefreshToken(raw)
	return raw, hash, nil
}

func hashRefreshToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return fmt.Sprintf("%x", sum[:])
}

func hashPassword(password string) (string, error) {
	return hashing.HashPasswordArgon2(password, hashing.DefaultArgon2Params())
}

func verifyPassword(passwordHash string, password string) bool {
	if strings.Contains(passwordHash, ".") {
		return hashing.CheckPasswordHashArgon2(passwordHash, password, hashing.DefaultArgon2Params())
	}
	return hashing.CheckPasswordHash([]byte(passwordHash), []byte(password))
}

func nullableString(value string) *string {
	v := strings.TrimSpace(value)
	if v == "" {
		return nil
	}
	return &v
}

func (s *Service) ForgotPassword(ctx context.Context, email string) error {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" {
		return nil
	}

	user, err := s.user.GetUserByEmail(ctx, email)
	if err != nil || user == nil {
		// Respond success always to prevent user enumeration
		return nil
	}

	rawToken, tokenHash, err := newRefreshToken()
	if err != nil {
		return err
	}

	expiresAt := time.Now().Add(1 * time.Hour)
	_, err = s.user.CreatePasswordResetToken(ctx, user.ID, tokenHash, expiresAt)
	if err != nil {
		return err
	}

	frontendURL := "http://localhost:3000"
	if s.conf != nil && s.conf.Val.FrontendURL != "" {
		frontendURL = s.conf.Val.FrontendURL
	}

	resetURL := fmt.Sprintf("%s/auth/reset-password?token=%s", frontendURL, rawToken)
	if s.mailer != nil {
		_ = s.mailer.SendPasswordReset(ctx, email, resetURL, true)
	}

	return nil
}

func (s *Service) ResetPassword(ctx context.Context, rawToken string, newPassword string) error {
	if strings.TrimSpace(rawToken) == "" || len(newPassword) < 8 {
		return ErrInvalidOrExpiredToken
	}

	tokenHash := hashRefreshToken(rawToken)
	token, err := s.user.GetPasswordResetTokenByHash(ctx, tokenHash)
	if err != nil || token == nil {
		return ErrInvalidOrExpiredToken
	}

	if token.UsedAt != nil || time.Now().After(token.ExpiresAt) {
		return ErrInvalidOrExpiredToken
	}

	hash, err := hashPassword(newPassword)
	if err != nil {
		return err
	}

	_, err = s.user.UpdateUserPassword(ctx, token.UserID, entitiesdto.UpdateUserPassword{
		NewPassword: hash,
	})
	if err != nil {
		return err
	}

	_ = s.user.MarkPasswordResetTokenUsed(ctx, token.ID)

	// Revoke all active sessions for this user
	_ = s.auth.RevokeAuthSessionsByUserID(ctx, token.UserID)

	return nil
}

func (s *Service) SendEmailVerification(ctx context.Context, user *ent.UserEntity) error {
	if user == nil || user.Email == nil || user.EmailVerifiedAt != nil {
		return nil
	}

	rawToken, tokenHash, err := newRefreshToken()
	if err != nil {
		return err
	}

	expiresAt := time.Now().Add(24 * time.Hour)
	_, err = s.user.CreateEmailVerificationToken(ctx, user.ID, tokenHash, expiresAt)
	if err != nil {
		return err
	}

	frontendURL := "http://localhost:3000"
	if s.conf != nil && s.conf.Val.FrontendURL != "" {
		frontendURL = s.conf.Val.FrontendURL
	}

	verifyURL := fmt.Sprintf("%s/auth/verify-email?token=%s", frontendURL, rawToken)
	if s.mailer != nil {
		_ = s.mailer.SendEmailVerification(ctx, *user.Email, verifyURL, true)
	}

	return nil
}

func (s *Service) VerifyEmail(ctx context.Context, rawToken string) error {
	if strings.TrimSpace(rawToken) == "" {
		return ErrInvalidOrExpiredToken
	}

	tokenHash := hashRefreshToken(rawToken)
	token, err := s.user.GetEmailVerificationTokenByHash(ctx, tokenHash)
	if err != nil || token == nil {
		return ErrInvalidOrExpiredToken
	}

	if time.Now().After(token.ExpiresAt) {
		return ErrInvalidOrExpiredToken
	}

	_, err = s.user.SetUserEmailVerified(ctx, token.UserID)
	if err != nil {
		return err
	}

	// Activate any plan entitlement that was approved while the email was
	// still unverified. If this fails, do NOT delete the verification token:
	// replaying the same verification link safely retries (SetUserEmailVerified
	// is idempotent and ActivatePendingEntitlements only ever activates rows
	// still missing activated_at), so no user-visible state is lost.
	if s.activator != nil {
		if _, err := s.activator.ActivatePendingEntitlements(ctx, token.UserID); err != nil {
			return err
		}
	}

	_ = s.user.DeleteEmailVerificationToken(ctx, token.ID)

	return nil
}

func (s *Service) ResendVerificationEmail(ctx context.Context, userID uuid.UUID) error {
	user, err := s.user.GetUserByID(ctx, userID)
	if err != nil || user == nil {
		return ErrUserNotFound
	}

	if user.EmailVerifiedAt != nil {
		return ErrEmailAlreadyVerified
	}

	return s.SendEmailVerification(ctx, user)
}
