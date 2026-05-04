package entitiesdto

type CreateUser struct {
	Email    *string `json:"email"`
	Password *string `json:"password"`
	Username *string `json:"username"`
}

type UpdateUser struct {
	Email    *string `json:"email"`
	Username *string `json:"username"`
	IsActive *bool   `json:"is_active"`
}

type UpdateUserPlan struct {
	Plan string `json:"plan"`
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
