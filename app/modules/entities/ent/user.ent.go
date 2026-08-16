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

	ID              uuid.UUID  `bun:"id,pk,type:uuid,default:gen_random_uuid()" json:"id"`
	Email           *string    `bun:"email,unique" json:"email,omitempty"`
	Password        *string    `bun:"password" json:"-"`
	Username        *string    `bun:"username" json:"username,omitempty"`
	Plan            PlanType   `bun:"plan,type:plan_type,notnull,default:'Free'" json:"plan"`
	PlanExpiresAt   *time.Time `bun:"plan_expires_at" json:"plan_expires_at,omitempty"`
	PlanCancelledAt *time.Time `bun:"plan_cancelled_at" json:"plan_cancelled_at,omitempty"`
	EmailVerifiedAt *time.Time `bun:"email_verified_at" json:"email_verified_at,omitempty"`
	IsActive        bool       `bun:"is_active,notnull" json:"is_active"`
	IsGuest         bool       `bun:"is_guest,notnull" json:"is_guest"`
	IsAdmin         bool       `bun:"is_admin,notnull,default:false" json:"is_admin"`
	LastLoginAt     *time.Time `bun:"last_login_at" json:"last_login_at,omitempty"`
	LoginCount      int        `bun:"login_count,notnull,default:0" json:"login_count"`
	LastLoginIP     *string    `bun:"last_login_ip" json:"last_login_ip,omitempty"`
	CreatedAt       time.Time  `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt       time.Time  `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
}
