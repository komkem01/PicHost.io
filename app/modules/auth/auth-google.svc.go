package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	entitiesdto "pichost.io/app/modules/entities/dto"
	"pichost.io/app/modules/entities/ent"
)

const GoogleStateCookieName = "google_oauth_state_nonce"

type googleStateClaims struct {
	Nonce string `json:"nonce"`
	Exp   int64  `json:"exp"`
}

type googleTokenResponse struct {
	AccessToken string `json:"access_token"`
}

type googleUserInfoResponse struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

func (s *Service) BuildGoogleAuthURL() (string, string, error) {
	if !s.googleOAuthConfigured() {
		return "", "", ErrGoogleOAuthNotConfigured
	}

	nonce, err := randomOpaque(24)
	if err != nil {
		return "", "", err
	}

	state, err := s.signGoogleState(nonce)
	if err != nil {
		return "", "", err
	}

	q := url.Values{}
	q.Set("client_id", s.conf.Val.GoogleClientID)
	q.Set("redirect_uri", s.conf.Val.GoogleRedirectURL)
	q.Set("response_type", "code")
	q.Set("scope", "openid email profile")
	q.Set("state", state)
	q.Set("access_type", "online")
	q.Set("include_granted_scopes", "true")
	q.Set("prompt", "select_account")

	return "https://accounts.google.com/o/oauth2/v2/auth?" + q.Encode(), nonce, nil
}

func (s *Service) GoogleCallback(
	ctx context.Context,
	code string,
	state string,
	stateNonceFromCookie string,
	userAgent string,
	ip string,
) (*AuthResultService, error) {
	if !s.googleOAuthConfigured() {
		return nil, ErrGoogleOAuthNotConfigured
	}
	if strings.TrimSpace(code) == "" {
		return nil, ErrGoogleOAuthInvalidCode
	}
	if err := s.verifyGoogleState(state, stateNonceFromCookie); err != nil {
		return nil, err
	}

	googleAccessToken, err := s.exchangeGoogleCode(ctx, code)
	if err != nil {
		return nil, err
	}

	userinfo, err := s.fetchGoogleUserInfo(ctx, googleAccessToken)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(userinfo.Sub) == "" {
		return nil, ErrAuthUnauthorized
	}

	user, err := s.findOrLinkGoogleIdentity(ctx, userinfo)
	if err != nil {
		return nil, err
	}
	if !user.IsActive {
		return nil, ErrAuthUnauthorized
	}

	return s.issueAuth(ctx, user, userAgent, ip)
}

func (s *Service) findOrLinkGoogleIdentity(ctx context.Context, userinfo *googleUserInfoResponse) (*ent.UserEntity, error) {
	account, err := s.auth.GetOAuthAccountByProviderUserID(ctx, "google", userinfo.Sub)
	if err == nil {
		return s.user.GetUserByID(ctx, account.UserID)
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	email := strings.TrimSpace(strings.ToLower(userinfo.Email))
	if email == "" {
		return nil, ErrAuthUnauthorized
	}

	user, err := s.user.GetUserByEmail(ctx, email)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}

		username := googleDefaultUsername(userinfo)
		created, createErr := s.user.CreateUser(ctx, entitiesdto.CreateUser{
			Email:    &email,
			Password: nil,
			Username: &username,
			Plan:     "Free",
			IsGuest:  false,
		})
		if createErr != nil {
			return nil, createErr
		}

		// Initialise quota row for the new Google user (best-effort).
		_, _ = s.quotaEnt.UpsertUserQuota(ctx, created.ID)

		user = created
	}

	_, err = s.auth.CreateOAuthAccount(ctx, entitiesdto.CreateOAuthAccount{
		UserID:         user.ID,
		Provider:       "google",
		ProviderUserID: userinfo.Sub,
		Email:          &email,
	})
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Service) fetchGoogleUserInfo(ctx context.Context, accessToken string) (*googleUserInfoResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://openidconnect.googleapis.com/v1/userinfo", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("google userinfo failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	var out googleUserInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *Service) exchangeGoogleCode(ctx context.Context, code string) (string, error) {
	form := url.Values{}
	form.Set("code", code)
	form.Set("client_id", s.conf.Val.GoogleClientID)
	form.Set("client_secret", s.conf.Val.GoogleClientSecret)
	form.Set("redirect_uri", s.conf.Val.GoogleRedirectURL)
	form.Set("grant_type", "authorization_code")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://oauth2.googleapis.com/token", strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return "", fmt.Errorf("google token exchange failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	var out googleTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	if strings.TrimSpace(out.AccessToken) == "" {
		return "", ErrAuthUnauthorized
	}

	return out.AccessToken, nil
}

func (s *Service) signGoogleState(nonce string) (string, error) {
	claims := googleStateClaims{
		Nonce: nonce,
		Exp:   time.Now().Add(time.Duration(s.conf.Val.GoogleStateTTLSeconds) * time.Second).Unix(),
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	payloadB64 := base64.RawURLEncoding.EncodeToString(payload)
	sig := s.signPayload("google_state." + payloadB64)
	return payloadB64 + "." + sig, nil
}

func (s *Service) verifyGoogleState(state string, stateNonceFromCookie string) error {
	parts := strings.Split(state, ".")
	if len(parts) != 2 {
		return ErrGoogleOAuthInvalidState
	}

	expectedSig := s.signPayload("google_state." + parts[0])
	if !secureStringEquals(expectedSig, parts[1]) {
		return ErrGoogleOAuthInvalidState
	}

	payloadRaw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return ErrGoogleOAuthInvalidState
	}

	var claims googleStateClaims
	if err := json.Unmarshal(payloadRaw, &claims); err != nil {
		return ErrGoogleOAuthInvalidState
	}

	if claims.Exp <= time.Now().Unix() {
		return ErrGoogleOAuthInvalidState
	}
	if strings.TrimSpace(claims.Nonce) == "" || claims.Nonce != stateNonceFromCookie {
		return ErrGoogleOAuthInvalidState
	}

	return nil
}

func (s *Service) googleOAuthConfigured() bool {
	return strings.TrimSpace(s.conf.Val.GoogleClientID) != "" &&
		strings.TrimSpace(s.conf.Val.GoogleClientSecret) != "" &&
		strings.TrimSpace(s.conf.Val.GoogleRedirectURL) != ""
}

func randomOpaque(size int) (string, error) {
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func googleDefaultUsername(userinfo *googleUserInfoResponse) string {
	candidate := strings.TrimSpace(strings.ToLower(userinfo.Name))
	candidate = strings.ReplaceAll(candidate, " ", "_")
	if candidate == "" {
		if idx := strings.Index(userinfo.Email, "@"); idx > 0 {
			candidate = userinfo.Email[:idx]
		}
	}
	if candidate == "" {
		candidate = "google_user"
	}
	sum := sha256.Sum256([]byte(userinfo.Sub + userinfo.Email))
	suffix := base64.RawURLEncoding.EncodeToString(sum[:])
	if len(suffix) > 8 {
		suffix = suffix[:8]
	}
	return candidate + "_" + suffix
}

func secureStringEquals(left string, right string) bool {
	if len(left) != len(right) {
		return false
	}
	var diff byte
	for i := 0; i < len(left); i++ {
		diff |= left[i] ^ right[i]
	}
	return diff == 0
}
