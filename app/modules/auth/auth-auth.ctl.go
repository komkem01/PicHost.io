package auth

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"pichost.io/app/utils/base"
	"pichost.io/config/i18n"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RegisterRequestController struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Username string `json:"username" binding:"required"`
}

type LoginRequestController struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type UserResponseController struct {
	ID       string  `json:"id"`
	Email    *string `json:"email"`
	Username *string `json:"username"`
	Plan     string  `json:"plan"`
	IsActive bool    `json:"is_active"`
	IsGuest  bool    `json:"is_guest"`
}

type AuthResponseController struct {
	AccessToken string                 `json:"access_token"`
	TokenType   string                 `json:"token_type"`
	ExpiresIn   int                    `json:"expires_in"`
	User        UserResponseController `json:"user"`
}

func toUserResponseController(userID uuid.UUID, email *string, username *string, plan string, isActive bool, isGuest bool) UserResponseController {
	return UserResponseController{
		ID:       userID.String(),
		Email:    email,
		Username: username,
		Plan:     plan,
		IsActive: isActive,
		IsGuest:  isGuest,
	}
}

func (c *Controller) Register(ctx *gin.Context) {
	var req RegisterRequestController
	if err := ctx.ShouldBindJSON(&req); err != nil {
		base.BadRequest(ctx, i18n.InvalidRequestForm, nil)
		return
	}

	res, err := c.svc.Register(ctx.Request.Context(), RegisterRequestService{
		Email:    req.Email,
		Password: req.Password,
		Username: req.Username,
	}, ctx.GetHeader("User-Agent"), ctx.ClientIP())
	if err != nil {
		c.recordAudit("auth.register", "failure", nil, strPtr("user"), nil,
			ctx.ClientIP(), ctx.GetHeader("User-Agent"),
			map[string]any{"email": req.Email}, strPtr(err.Error()))
		if errors.Is(err, ErrUserEmailAlreadyExists) {
			base.BadRequest(ctx, i18n.BadRequest, gin.H{"error": err.Error()})
			return
		}
		base.InternalServerError(ctx, i18n.InternalError, gin.H{"error": err.Error()})
		return
	}

	c.recordAudit("auth.register", "success", uuidPtr(res.User.ID), strPtr("user"), uuidPtr(res.User.ID),
		ctx.ClientIP(), ctx.GetHeader("User-Agent"),
		map[string]any{"email": req.Email, "username": req.Username, "plan": string(res.User.Plan)}, nil)

	c.setRefreshCookie(ctx, res.RefreshToken)
	base.Success(ctx, AuthResponseController{
		AccessToken: res.AccessToken,
		TokenType:   "Bearer",
		ExpiresIn:   res.AccessExpiry,
		User:        toUserResponseController(res.User.ID, res.User.Email, res.User.Username, string(res.User.Plan), res.User.IsActive, res.User.IsGuest),
	})
}

func (c *Controller) Login(ctx *gin.Context) {
	var req LoginRequestController
	if err := ctx.ShouldBindJSON(&req); err != nil {
		base.BadRequest(ctx, i18n.InvalidRequestForm, nil)
		return
	}

	res, err := c.svc.Login(ctx.Request.Context(), LoginRequestService{
		Email:    req.Email,
		Password: req.Password,
	}, ctx.GetHeader("User-Agent"), ctx.ClientIP())
	if err != nil {
		c.recordAudit("auth.login", "failure", nil, strPtr("user"), nil,
			ctx.ClientIP(), ctx.GetHeader("User-Agent"),
			map[string]any{"email": req.Email}, strPtr(err.Error()))
		if errors.Is(err, ErrAuthInvalidCredentials) || errors.Is(err, ErrAuthUnauthorized) {
			base.Unauthorized(ctx, i18n.Unauthorized, gin.H{"error": err.Error()})
			return
		}
		base.InternalServerError(ctx, i18n.InternalError, gin.H{"error": err.Error()})
		return
	}

	c.recordAudit("auth.login", "success", uuidPtr(res.User.ID), strPtr("user"), uuidPtr(res.User.ID),
		ctx.ClientIP(), ctx.GetHeader("User-Agent"),
		map[string]any{"email": req.Email, "plan": string(res.User.Plan)}, nil)

	c.setRefreshCookie(ctx, res.RefreshToken)
	base.Success(ctx, AuthResponseController{
		AccessToken: res.AccessToken,
		TokenType:   "Bearer",
		ExpiresIn:   res.AccessExpiry,
		User:        toUserResponseController(res.User.ID, res.User.Email, res.User.Username, string(res.User.Plan), res.User.IsActive, res.User.IsGuest),
	})
}

func (c *Controller) Refresh(ctx *gin.Context) {
	refreshToken, err := ctx.Cookie(c.svc.conf.Val.RefreshCookieName)
	if err != nil {
		base.Unauthorized(ctx, i18n.Unauthorized, gin.H{"error": ErrAuthInvalidRefreshToken.Error()})
		return
	}

	res, err := c.svc.Refresh(ctx.Request.Context(), refreshToken, ctx.GetHeader("User-Agent"), ctx.ClientIP())
	if err != nil {
		c.recordAudit("auth.token_refresh", "failure", nil, strPtr("session"), nil,
			ctx.ClientIP(), ctx.GetHeader("User-Agent"), nil, strPtr(err.Error()))
		if errors.Is(err, ErrAuthSessionNotFound) || errors.Is(err, ErrAuthInvalidRefreshToken) || errors.Is(err, ErrAuthUnauthorized) {
			base.Unauthorized(ctx, i18n.Unauthorized, gin.H{"error": err.Error()})
			return
		}
		base.InternalServerError(ctx, i18n.InternalError, gin.H{"error": err.Error()})
		return
	}

	c.recordAudit("auth.token_refresh", "success", uuidPtr(res.User.ID), strPtr("session"), nil,
		ctx.ClientIP(), ctx.GetHeader("User-Agent"), nil, nil)

	c.setRefreshCookie(ctx, res.RefreshToken)
	base.Success(ctx, AuthResponseController{
		AccessToken: res.AccessToken,
		TokenType:   "Bearer",
		ExpiresIn:   res.AccessExpiry,
		User:        toUserResponseController(res.User.ID, res.User.Email, res.User.Username, string(res.User.Plan), res.User.IsActive, res.User.IsGuest),
	})
}

func (c *Controller) Logout(ctx *gin.Context) {
	refreshToken, _ := ctx.Cookie(c.svc.conf.Val.RefreshCookieName)
	if err := c.svc.LogoutByRefreshToken(ctx.Request.Context(), refreshToken); err != nil {
		base.InternalServerError(ctx, i18n.InternalError, gin.H{"error": err.Error()})
		return
	}

	c.recordAudit("auth.logout", "success", nil, strPtr("session"), nil,
		ctx.ClientIP(), ctx.GetHeader("User-Agent"), nil, nil)

	c.clearRefreshCookie(ctx)
	_ = base.RawJSON(ctx, http.StatusNoContent, gin.H{})
}

func (c *Controller) Me(ctx *gin.Context) {
	userIDValue, exists := ctx.Get("auth_user_id")
	if !exists {
		base.Unauthorized(ctx, i18n.Unauthorized, nil)
		return
	}

	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		base.Unauthorized(ctx, i18n.Unauthorized, nil)
		return
	}

	user, err := c.svc.Me(ctx.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			_ = base.JSON(ctx, http.StatusNotFound, i18n.UserNotFound, nil, nil)
			return
		}
		base.InternalServerError(ctx, i18n.InternalError, gin.H{"error": err.Error()})
		return
	}

	base.Success(ctx, toUserResponseController(user.ID, user.Email, user.Username, string(user.Plan), user.IsActive, user.IsGuest), i18n.UserFetched)
}

func (c *Controller) GoogleLogin(ctx *gin.Context) {
	googleAuthURL, stateNonce, err := c.svc.BuildGoogleAuthURL()
	if err != nil {
		if errors.Is(err, ErrGoogleOAuthNotConfigured) {
			base.BadRequest(ctx, i18n.BadRequest, gin.H{"error": err.Error()})
			return
		}
		base.InternalServerError(ctx, i18n.InternalError, gin.H{"error": err.Error()})
		return
	}

	ctx.SetSameSite(http.SameSiteLaxMode)
	ctx.SetCookie(
		GoogleStateCookieName,
		stateNonce,
		c.svc.conf.Val.GoogleStateTTLSeconds,
		"/",
		c.svc.conf.Val.RefreshCookieDomain,
		c.svc.conf.Val.RefreshCookieSecure,
		true,
	)

	ctx.Redirect(http.StatusFound, googleAuthURL)
}

func (c *Controller) GoogleCallback(ctx *gin.Context) {
	stateNonce, err := ctx.Cookie(GoogleStateCookieName)
	if err != nil {
		base.BadRequest(ctx, i18n.BadRequest, gin.H{"error": ErrGoogleOAuthInvalidState.Error()})
		return
	}

	res, err := c.svc.GoogleCallback(
		ctx.Request.Context(),
		ctx.Query("code"),
		ctx.Query("state"),
		stateNonce,
		ctx.GetHeader("User-Agent"),
		ctx.ClientIP(),
	)
	if err != nil {
		c.recordAudit("auth.google_login", "failure", nil, strPtr("user"), nil,
			ctx.ClientIP(), ctx.GetHeader("User-Agent"), nil, strPtr(err.Error()))
		if errors.Is(err, ErrGoogleOAuthNotConfigured) ||
			errors.Is(err, ErrGoogleOAuthInvalidState) ||
			errors.Is(err, ErrGoogleOAuthInvalidCode) {
			base.BadRequest(ctx, i18n.BadRequest, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, ErrAuthUnauthorized) {
			base.Unauthorized(ctx, i18n.Unauthorized, gin.H{"error": err.Error()})
			return
		}
		base.InternalServerError(ctx, i18n.InternalError, gin.H{"error": err.Error()})
		return
	}

	c.recordAudit("auth.google_login", "success", uuidPtr(res.User.ID), strPtr("user"), uuidPtr(res.User.ID),
		ctx.ClientIP(), ctx.GetHeader("User-Agent"),
		map[string]any{"plan": string(res.User.Plan)}, nil)

	ctx.SetSameSite(http.SameSiteLaxMode)
	ctx.SetCookie(
		GoogleStateCookieName,
		"",
		-1,
		"/",
		c.svc.conf.Val.RefreshCookieDomain,
		c.svc.conf.Val.RefreshCookieSecure,
		true,
	)

	c.setRefreshCookie(ctx, res.RefreshToken)

	// Redirect to frontend callback page; token is in the URL fragment
	// so it never reaches server logs or referrer headers.
	callbackURL := fmt.Sprintf(
		"%s/auth/callback#access_token=%s",
		c.svc.conf.Val.FrontendURL,
		url.QueryEscape(res.AccessToken),
	)
	ctx.Redirect(http.StatusFound, callbackURL)
}

func (c *Controller) setRefreshCookie(ctx *gin.Context, refreshToken string) {
	ctx.SetSameSite(http.SameSiteLaxMode)
	ctx.SetCookie(
		c.svc.conf.Val.RefreshCookieName,
		refreshToken,
		c.svc.conf.Val.RefreshTokenTTLSeconds,
		"/",
		c.svc.conf.Val.RefreshCookieDomain,
		c.svc.conf.Val.RefreshCookieSecure,
		true,
	)
}

func (c *Controller) clearRefreshCookie(ctx *gin.Context) {
	ctx.SetSameSite(http.SameSiteLaxMode)
	ctx.SetCookie(
		c.svc.conf.Val.RefreshCookieName,
		"",
		-1,
		"/",
		c.svc.conf.Val.RefreshCookieDomain,
		c.svc.conf.Val.RefreshCookieSecure,
		true,
	)
}
