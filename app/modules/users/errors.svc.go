package users

import "errors"

var ErrUserNotFound = errors.New("user not found")
var ErrUserEmailAlreadyExists = errors.New("user email already exists")
var ErrUserInvalidPassword = errors.New("invalid user password")
var ErrUserInvalidID = errors.New("invalid user id")
var ErrUserInvalidEmail = errors.New("invalid user email")
var ErrUserInvalidUsername = errors.New("invalid user username")
var ErrUserInvalidPlan = errors.New("invalid user plan")
var ErrUserInvalidIsActive = errors.New("invalid user is_active")
var ErrUserInvalidIsGuest = errors.New("invalid user is_guest")
var ErrAuthInvalidCredentials = errors.New("invalid credentials")
var ErrAuthUnauthorized = errors.New("unauthorized")
var ErrAuthSessionNotFound = errors.New("auth session not found")
var ErrAuthInvalidRefreshToken = errors.New("invalid refresh token")
