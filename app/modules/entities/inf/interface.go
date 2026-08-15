package entitiesinf

import (
	"context"
	"time"

	entitiesdto "pichost.io/app/modules/entities/dto"
	"pichost.io/app/modules/entities/ent"

	"github.com/google/uuid"
)

type UserEntity interface {
	CreateUser(ctx context.Context, user entitiesdto.CreateUser) (*ent.UserEntity, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*ent.UserEntity, error)
	GetListUser(ctx context.Context) ([]*ent.UserEntity, error)
	GetUserByEmail(ctx context.Context, email string) (*ent.UserEntity, error)
	UpdateUser(ctx context.Context, id uuid.UUID, user entitiesdto.UpdateUser) (*ent.UserEntity, error)
	UpdateUserPlan(ctx context.Context, id uuid.UUID, plan entitiesdto.UpdateUserPlan) (*ent.UserEntity, error)
	SetUserPlanExpiry(ctx context.Context, id uuid.UUID, expiresAt *time.Time, clearCancellation bool) (*ent.UserEntity, error)
	CancelUserPlan(ctx context.Context, id uuid.UUID) (*ent.UserEntity, error)
	DowngradeUserPlanToFree(ctx context.Context, id uuid.UUID) (*ent.UserEntity, error)
	UpdateUserProfile(ctx context.Context, id uuid.UUID, profile entitiesdto.UpdateUserProfile) (*ent.UserEntity, error)
	SetUserActive(ctx context.Context, id uuid.UUID, isActive bool) error
	UpdateUserPassword(ctx context.Context, id uuid.UUID, password entitiesdto.UpdateUserPassword) (*ent.UserEntity, error)
	SetUserAdmin(ctx context.Context, id uuid.UUID, isAdmin bool) error
	SetUserEmailVerified(ctx context.Context, id uuid.UUID) (*ent.UserEntity, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error

	CreatePasswordResetToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) (*ent.PasswordResetTokenEntity, error)
	GetPasswordResetTokenByHash(ctx context.Context, tokenHash string) (*ent.PasswordResetTokenEntity, error)
	MarkPasswordResetTokenUsed(ctx context.Context, id uuid.UUID) error

	CreateEmailVerificationToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) (*ent.EmailVerificationTokenEntity, error)
	GetEmailVerificationTokenByHash(ctx context.Context, tokenHash string) (*ent.EmailVerificationTokenEntity, error)
	DeleteEmailVerificationToken(ctx context.Context, id uuid.UUID) error
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
	GetImageByStorageID(ctx context.Context, storageID uuid.UUID) (*ent.ImageEntity, error)
	GetImagesByUserID(ctx context.Context, userID uuid.UUID) ([]*ent.ImageEntity, error)
	UpdateImage(ctx context.Context, id uuid.UUID, image entitiesdto.UpdateImage) (*ent.ImageEntity, error)
	DeleteImage(ctx context.Context, id uuid.UUID) error
	ListExpiredImages(ctx context.Context, before time.Time) ([]*ent.ImageEntity, error)
	ListAllImages(ctx context.Context, limit int, offset int) ([]*ent.ImageEntity, int, error)
	GetGuestStats(ctx context.Context) (int, int64, error)
	GetUniqueGuestIPCount(ctx context.Context, since time.Time) (int, error)
	IncrementImageViewCount(ctx context.Context, id uuid.UUID) error
	GetTotalImageViewsByUserID(ctx context.Context, userID uuid.UUID) (int64, error)
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

type PlanSettingEntity interface {
	ListPlanSettings(ctx context.Context) ([]*ent.PlanSettingEntity, error)
	GetPlanSettingByKey(ctx context.Context, key string) (*ent.PlanSettingEntity, error)
	UpsertPlanSetting(ctx context.Context, setting entitiesdto.UpsertPlanSetting) (*ent.PlanSettingEntity, error)
	DeletePlanSettingByKey(ctx context.Context, key string) error
}

type PaymentTransactionEntity interface {
	CreatePaymentTransaction(ctx context.Context, in entitiesdto.CreatePaymentTransaction) (*ent.PaymentTransactionEntity, error)
	GetPaymentTransactionByID(ctx context.Context, id uuid.UUID) (*ent.PaymentTransactionEntity, error)
	GetPaymentTransactionByCheckoutReference(ctx context.Context, checkoutReference string) (*ent.PaymentTransactionEntity, error)
	ListPaymentTransactionsByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]*ent.PaymentTransactionEntity, error)
	UpdatePaymentTransactionStatus(ctx context.Context, id uuid.UUID, in entitiesdto.UpdatePaymentTransactionStatus) (*ent.PaymentTransactionEntity, error)
	UpdatePaymentSlipStorageID(ctx context.Context, id uuid.UUID, in entitiesdto.UpdatePaymentSlipStorageID) (*ent.PaymentTransactionEntity, error)
	ListPaymentTransactions(ctx context.Context, limit int, offset int) ([]*ent.PaymentTransactionEntity, int, error)
}

type AuditEntity interface {
	CreateAuditLog(ctx context.Context, log entitiesdto.CreateAuditLog) error
	ListAuditLogs(ctx context.Context, filter entitiesdto.ListAuditLogsFilter) ([]*ent.AuditLogEntity, int, error)
}

