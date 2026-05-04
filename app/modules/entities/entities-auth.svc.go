package entities

import (
	"context"
	"database/sql"
	"time"

	entitiesdto "pichost.io/app/modules/entities/dto"
	"pichost.io/app/modules/entities/ent"
	entitiesinf "pichost.io/app/modules/entities/inf"

	"github.com/google/uuid"
)

var _ entitiesinf.AuthEntity = (*Service)(nil)

func (s *Service) CreateAuthSession(ctx context.Context, auth entitiesdto.CreateAuthSession) (*ent.AuthSessionEntity, error) {
	now := time.Now()
	data := &ent.AuthSessionEntity{
		UserID:           auth.UserID,
		RefreshTokenHash: auth.RefreshTokenHash,
		UserAgent:        auth.UserAgent,
		IPAddress:        auth.IPAddress,
		ExpiresAt:        auth.ExpiresAt,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	_, err := s.db.NewInsert().
		Model(data).
		Exec(ctx)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (s *Service) GetAuthSessionByRefreshTokenHash(ctx context.Context, refreshTokenHash string) (*ent.AuthSessionEntity, error) {
	var session ent.AuthSessionEntity
	err := s.db.NewSelect().
		Model(&session).
		Where("refresh_token_hash = ?", refreshTokenHash).
		Where("revoked_at IS NULL").
		Where("expires_at > ?", time.Now()).
		Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}

	return &session, nil
}

func (s *Service) RotateAuthSession(ctx context.Context, id uuid.UUID, auth entitiesdto.RotateAuthSession) (*ent.AuthSessionEntity, error) {
	now := time.Now()
	data := &ent.AuthSessionEntity{}
	_, err := s.db.NewUpdate().
		Model(data).
		Set("refresh_token_hash = ?", auth.RefreshTokenHash).
		Set("expires_at = ?", auth.ExpiresAt).
		Set("revoked_at = NULL").
		Set("updated_at = ?", now).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return nil, err
	}

	err = s.db.NewSelect().
		Model(data).
		Where("id = ?", id).
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (s *Service) RevokeAuthSession(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	_, err := s.db.NewUpdate().
		Model((*ent.AuthSessionEntity)(nil)).
		Set("revoked_at = ?", now).
		Set("updated_at = ?", now).
		Where("id = ?", id).
		Where("revoked_at IS NULL").
		Exec(ctx)
	return err
}

func (s *Service) RevokeAuthSessionsByUserID(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	_, err := s.db.NewUpdate().
		Model((*ent.AuthSessionEntity)(nil)).
		Set("revoked_at = ?", now).
		Set("updated_at = ?", now).
		Where("user_id = ?", userID).
		Where("revoked_at IS NULL").
		Exec(ctx)
	return err
}

func (s *Service) CreateOAuthAccount(ctx context.Context, account entitiesdto.CreateOAuthAccount) (*ent.OAuthAccountEntity, error) {
	now := time.Now()
	data := &ent.OAuthAccountEntity{
		UserID:         account.UserID,
		Provider:       account.Provider,
		ProviderUserID: account.ProviderUserID,
		Email:          account.Email,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	_, err := s.db.NewInsert().
		Model(data).
		Exec(ctx)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (s *Service) GetOAuthAccountByProviderUserID(ctx context.Context, provider string, providerUserID string) (*ent.OAuthAccountEntity, error) {
	var account ent.OAuthAccountEntity
	err := s.db.NewSelect().
		Model(&account).
		Where("provider = ?", provider).
		Where("provider_user_id = ?", providerUserID).
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	return &account, nil
}
