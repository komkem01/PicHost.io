package auth

import "errors"

var ErrUserNotFound = errors.New("user not found")
var ErrUserEmailAlreadyExists = errors.New("user email already exists")
var ErrAuthInvalidCredentials = errors.New("email or password is incorrect")
var ErrAuthUnauthorized = errors.New("unauthorized")
var ErrAuthSessionNotFound = errors.New("auth session not found")
var ErrAuthInvalidRefreshToken = errors.New("invalid refresh token")
var ErrGoogleOAuthNotConfigured = errors.New("google oauth is not configured")
var ErrGoogleOAuthInvalidState = errors.New("invalid google oauth state")
var ErrGoogleOAuthInvalidCode = errors.New("invalid google oauth code")
