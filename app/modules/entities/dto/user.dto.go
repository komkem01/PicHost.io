package entitiesdto

type CreateUser struct {
	Email    *string `json:"email"`
	Password *string `json:"password"`
	Username *string `json:"username"`
	Plan     string  `json:"plan"`
	IsGuest  bool    `json:"is_guest"`
}

type UpdateUser struct {
	Email    *string `json:"email"`
	Username *string `json:"username"`
	IsActive *bool   `json:"is_active"`
	Plan     *string `json:"plan"`
	IsGuest  *bool   `json:"is_guest"`
	// ClearEmailVerification, when true, resets email_verified_at to NULL in the
	// same UPDATE statement that changes Email. Callers changing a user's email
	// to a new address must set this so the DB is never left, even momentarily,
	// with the new (unverified) address alongside a stale, still-set
	// email_verified_at from the previous address.
	ClearEmailVerification bool `json:"-"`
}

type UpdateUserPlan struct {
	Plan *string `json:"plan"`
}

type UpdateUserProfile struct {
	Email    *string `json:"email"`
	Username *string `json:"username"`
}

type UpdateUserPassword struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

type UserResponse struct {
	ID        string  `json:"id"`
	Email     *string `json:"email"`
	Username  *string `json:"username"`
	Plan      string  `json:"plan"`
	IsActive  bool    `json:"is_active"`
	IsGuest   bool    `json:"is_guest"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}
