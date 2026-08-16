package entities

import (
	"context"
	"fmt"
	"time"

	entitiesdto "pichost.io/app/modules/entities/dto"
	"pichost.io/app/modules/entities/ent"
	entitiesinf "pichost.io/app/modules/entities/inf"

	"github.com/google/uuid"
)

var _ entitiesinf.StorageEntity = (*Service)(nil)

func (s *Service) CreateStorage(ctx context.Context, storage entitiesdto.CreateStorage) (*ent.StorageEntity, error) {
	now := time.Now()
	data := &ent.StorageEntity{
		ShortCode: storage.ShortCode,
		Provider:  storage.Provider,
		Path:      storage.Path,
		URL:       storage.URL,
		FileSize:  storage.FileSize,
		MIMEType:  storage.MIMEType,
		CreatedAt: now,
	}

	_, err := s.db.NewInsert().
		Model(data).
		Exec(ctx)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (s *Service) GetStorageByID(ctx context.Context, id uuid.UUID) (*ent.StorageEntity, error) {
	var storage ent.StorageEntity
	err := s.db.NewSelect().
		Model(&storage).
		Where("id = ?", id).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &storage, nil
}

func (s *Service) GetStorageByShortCode(ctx context.Context, shortCode string) (*ent.StorageEntity, error) {
	var storage ent.StorageEntity
	err := s.db.NewSelect().
		Model(&storage).
		Where("short_code = ?", shortCode).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &storage, nil
}

func (s *Service) GetListStorage(ctx context.Context) ([]*ent.StorageEntity, error) {
	var storages []*ent.StorageEntity
	err := s.db.NewSelect().
		Model(&storages).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return storages, nil
}

func (s *Service) GetStorageByURL(ctx context.Context, url string) (*ent.StorageEntity, error) {
	var storage ent.StorageEntity
	err := s.db.NewSelect().
		Model(&storage).
		Where("url = ?", url).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &storage, nil
}

func (s *Service) GetStorageByEmail(ctx context.Context, email string) (*ent.StorageEntity, error) {
	var storage ent.StorageEntity
	err := s.db.NewSelect().
		Model(&storage).
		Where("email = ?", email).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &storage, nil
}

func (s *Service) UpdateStorage(ctx context.Context, id uuid.UUID, storage entitiesdto.UpdateStorage) (*ent.StorageEntity, error) {
	query := s.db.NewUpdate().
		Model((*ent.StorageEntity)(nil)).
		Table("storages").
		Where("id = ?", id)

	updated := 0
	if storage.ShortCode != nil {
		query = query.Set("short_code = ?", *storage.ShortCode)
		updated++
	}
	if storage.Provider != nil {
		query = query.Set("provider = ?", *storage.Provider)
		updated++
	}
	if storage.Path != nil {
		query = query.Set("path = ?", *storage.Path)
		updated++
	}
	if storage.URL != nil {
		query = query.Set("url = ?", *storage.URL)
		updated++
	}
	if storage.FileSize != nil {
		query = query.Set("file_size = ?", *storage.FileSize)
		updated++
	}
	if storage.MIMEType != nil {
		query = query.Set("mime_type = ?", *storage.MIMEType)
		updated++
	}

	if updated == 0 {
		return s.GetStorageByID(ctx, id)
	}

	result, err := query.Exec(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := result.RowsAffected()
	if err == nil && rows == 0 {
		return nil, fmt.Errorf("storage not found")
	}

	return s.GetStorageByID(ctx, id)
}

func (s *Service) DeleteStorage(ctx context.Context, id uuid.UUID) error {
	_, err := s.db.NewDelete().
		Model((*ent.StorageEntity)(nil)).
		Where("id = ?", id).
		Exec(ctx)
	return err
}

func (s *Service) GetStorageStats(ctx context.Context) (*entitiesdto.StorageStats, error) {
	stats := &entitiesdto.StorageStats{
		ProviderBreakdown: make(map[string]int64),
	}

	// Total files and bytes
	var totalRow struct {
		TotalFiles int64 `bun:"total_files"`
		TotalBytes int64 `bun:"total_bytes"`
	}
	_ = s.db.NewSelect().
		Model((*ent.StorageEntity)(nil)).
		ColumnExpr("COUNT(*) as total_files, COALESCE(SUM(file_size), 0) as total_bytes").
		Scan(ctx, &totalRow)
	stats.TotalFiles = totalRow.TotalFiles
	stats.TotalBytes = totalRow.TotalBytes

	// Orphan files and bytes
	var orphanRow struct {
		OrphanFiles int64 `bun:"orphan_files"`
		OrphanBytes int64 `bun:"orphan_bytes"`
	}
	_ = s.db.NewSelect().
		TableExpr("storages AS s").
		ColumnExpr("COUNT(s.id) as orphan_files, COALESCE(SUM(s.file_size), 0) as orphan_bytes").
		Join("LEFT JOIN images AS i ON s.id = i.storage_id").
		Join("LEFT JOIN payment_transactions AS pt ON s.id::text = pt.slip_storage_id").
		Where("i.id IS NULL AND pt.id IS NULL").
		Scan(ctx, &orphanRow)
	stats.OrphanFiles = orphanRow.OrphanFiles
	stats.OrphanBytes = orphanRow.OrphanBytes

	// Provider breakdown
	var providerRows []struct {
		Provider   string `bun:"provider"`
		TotalBytes int64  `bun:"total_bytes"`
	}
	_ = s.db.NewSelect().
		Model((*ent.StorageEntity)(nil)).
		ColumnExpr("provider, COALESCE(SUM(file_size), 0) as total_bytes").
		Group("provider").
		Scan(ctx, &providerRows)
	for _, row := range providerRows {
		stats.ProviderBreakdown[row.Provider] = row.TotalBytes
	}

	return stats, nil
}

func (s *Service) ListOrphanedStorage(ctx context.Context, limit int, offset int) ([]*ent.StorageEntity, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	var storages []*ent.StorageEntity
	count, err := s.db.NewSelect().
		TableExpr("storages AS s").
		ColumnExpr("s.*").
		Join("LEFT JOIN images AS i ON s.id = i.storage_id").
		Join("LEFT JOIN payment_transactions AS pt ON s.id::text = pt.slip_storage_id").
		Where("i.id IS NULL AND pt.id IS NULL").
		Order("s.created_at DESC").
		Limit(limit).
		Offset(offset).
		ScanAndCount(ctx, &storages)
	if err != nil {
		return nil, 0, err
	}
	return storages, count, nil
}

func (s *Service) CleanupOrphanedStorage(ctx context.Context, ids []uuid.UUID) (int64, error) {
	q := s.db.NewDelete().
		TableExpr("storages AS s").
		Where("s.id NOT IN (SELECT storage_id FROM images WHERE storage_id IS NOT NULL)").
		Where("s.id::text NOT IN (SELECT slip_storage_id FROM payment_transactions WHERE slip_storage_id IS NOT NULL)")

	if len(ids) > 0 {
		q = q.Where("s.id IN (?)", ids)
	}

	res, err := q.Exec(ctx)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

