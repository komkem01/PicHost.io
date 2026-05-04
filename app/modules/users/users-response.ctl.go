package users

import (
	"time"

	"pichost.io/app/modules/entities/ent"
)

type UserResponseController struct {
	ID        string  `json:"id"`
	Email     *string `json:"email"`
	Username  *string `json:"username"`
	Plan      string  `json:"plan"`
	IsActive  bool    `json:"is_active"`
	IsGuest   bool    `json:"is_guest"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

func toUserResponseController(user *ent.UserEntity) UserResponseController {
	return UserResponseController{
		ID:        user.ID.String(),
		Email:     user.Email,
		Username:  user.Username,
		Plan:      string(user.Plan),
		IsActive:  user.IsActive,
		IsGuest:   user.IsGuest,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
	}
}
