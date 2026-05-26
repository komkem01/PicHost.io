package ent

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type PlanType string

const (
	PlanTypeFree       PlanType = "Free"
	PlanTypeBasic      PlanType = "Basic"
	PlanTypePro        PlanType = "Pro"
	PlanTypeEnterprise PlanType = "Enterprise"
)

type UserEntity struct {
	bun.BaseModel `bun:"table:users,alias:u"`

	ID              uuid.UUID  `bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	Email           *string    `bun:"email,unique"`
	Password        *string    `bun:"password"`
	Username        *string    `bun:"username"`
	Plan            PlanType   `bun:"plan,type:plan_type,notnull,default:'Free'"`
	PlanExpiresAt   *time.Time `bun:"plan_expires_at"`
	PlanCancelledAt *time.Time `bun:"plan_cancelled_at"`
	IsActive        bool       `bun:"is_active,notnull"`
	IsGuest         bool       `bun:"is_guest,notnull"`
	IsAdmin         bool       `bun:"is_admin,notnull,default:false"`
	CreatedAt       time.Time  `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt       time.Time  `bun:"updated_at,notnull,default:current_timestamp"`
}
