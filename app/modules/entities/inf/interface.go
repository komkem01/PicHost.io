package entitiesinf

import (
	"context"
	"time"

	entitiesdto "pichost.io/app/modules/entities/dto"
	"pichost.io/app/modules/entities/ent"

	"github.com/google/uuid"
)

// ObjectEntity defines the interface for object entity operations such as create, retrieve, update, and soft delete.
type ExampleEntity interface {
	CreateExample(ctx context.Context, userID uuid.UUID) (*ent.Example, error)
	GetExampleByID(ctx context.Context, id uuid.UUID) (*ent.Example, error)
	UpdateExampleByID(ctx context.Context, id uuid.UUID, status ent.ExampleStatus) (*ent.Example, error)
	SoftDeleteExampleByID(ctx context.Context, id uuid.UUID) error
	ListExamplesByStatus(ctx context.Context, status ent.ExampleStatus) ([]*ent.Example, error)
}
type ExampleTwoEntity interface {
	CreateExampleTwo(ctx context.Context, userID uuid.UUID) (*ent.Example, error)
}

type UserEntity interface {
	CreateUser(ctx context.Context, user entitiesdto.CreateUser) (*ent.UserEntity, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*ent.UserEntity, error)
	GetListUser(ctx context.Context) ([]*ent.UserEntity, error)
	GetUserByEmail(ctx context.Context, email string) (*ent.UserEntity, error)
	UpdateUser(ctx context.Context, id uuid.UUID, user entitiesdto.UpdateUser) (*ent.UserEntity, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
}

type StorageEntity interface {
	CreateStorage(ctx context.Context, storage entitiesdto.CreateStorage) (*ent.StorageEntity, error)
	GetStorageByID(ctx context.Context, id uuid.UUID) (*ent.StorageEntity, error)
	GetStorageByShortCode(ctx context.Context, shortCode string) (*ent.StorageEntity, error)
	GetListStorage(ctx context.Context) ([]*ent.StorageEntity, error)
	GetStorageByURL(ctx context.Context, url string) (*ent.StorageEntity, error)
	GetStorageByEmail(ctx context.Context, email string) (*ent.StorageEntity, error)
	UpdateStorage(ctx context.Context, id uuid.UUID, storage entitiesdto.UpdateStorage) (*ent.StorageEntity, error)
	DeleteStorage(ctx context.Context, id uuid.UUID) error
}

type ImageEntity interface {
	CreateImage(ctx context.Context, image entitiesdto.CreateImage) (*ent.ImageEntity, error)
	GetImageByID(ctx context.Context, id uuid.UUID) (*ent.ImageEntity, error)
	UpdateImage(ctx context.Context, id uuid.UUID, image entitiesdto.UpdateImage) (*ent.ImageEntity, error)
	DeleteImage(ctx context.Context, id uuid.UUID) error
	ListExpiredImages(ctx context.Context, before time.Time) ([]*ent.ImageEntity, error)
}

type AuthEntity interface {
	CreateAuthSession(ctx context.Context, auth entitiesdto.CreateAuthSession) (*ent.AuthSessionEntity, error)
	GetAuthSessionByRefreshTokenHash(ctx context.Context, refreshTokenHash string) (*ent.AuthSessionEntity, error)
	RotateAuthSession(ctx context.Context, id uuid.UUID, auth entitiesdto.RotateAuthSession) (*ent.AuthSessionEntity, error)
	RevokeAuthSession(ctx context.Context, id uuid.UUID) error
	RevokeAuthSessionsByUserID(ctx context.Context, userID uuid.UUID) error
	CreateOAuthAccount(ctx context.Context, account entitiesdto.CreateOAuthAccount) (*ent.OAuthAccountEntity, error)
	GetOAuthAccountByProviderUserID(ctx context.Context, provider string, providerUserID string) (*ent.OAuthAccountEntity, error)
}

type UserQuotaEntity interface {
	GetUserQuota(ctx context.Context, userID uuid.UUID) (*ent.UserQuotaEntity, error)
	UpsertUserQuota(ctx context.Context, userID uuid.UUID) (*ent.UserQuotaEntity, error)
	AddToUserQuota(ctx context.Context, userID uuid.UUID, delta entitiesdto.AddToUserQuota) (*ent.UserQuotaEntity, error)
}
