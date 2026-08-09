package storage

import (
	"context"
	"database/sql"
	"testing"
	"time"

	entitiesdto "pichost.io/app/modules/entities/dto"
	"pichost.io/app/modules/entities/ent"
	"pichost.io/internal/config"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
)

type mockStorageEnt struct {
	storages map[uuid.UUID]*ent.StorageEntity
}

func newMockStorageEnt() *mockStorageEnt {
	return &mockStorageEnt{storages: make(map[uuid.UUID]*ent.StorageEntity)}
}

func (m *mockStorageEnt) CreateStorage(ctx context.Context, s entitiesdto.CreateStorage) (*ent.StorageEntity, error) {
	id := uuid.New()
	st := &ent.StorageEntity{
		ID:        id,
		ShortCode: s.ShortCode,
		URL:       s.URL,
		FileSize:  s.FileSize,
		MIMEType:  s.MIMEType,
		CreatedAt: time.Now(),
	}
	m.storages[id] = st
	return st, nil
}
func (m *mockStorageEnt) GetStorageByID(ctx context.Context, id uuid.UUID) (*ent.StorageEntity, error) {
	s, ok := m.storages[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	return s, nil
}
func (m *mockStorageEnt) GetStorageByShortCode(ctx context.Context, shortCode string) (*ent.StorageEntity, error) {
	for _, s := range m.storages {
		if s.ShortCode == shortCode {
			return s, nil
		}
	}
	return nil, sql.ErrNoRows
}
func (m *mockStorageEnt) GetListStorage(ctx context.Context) ([]*ent.StorageEntity, error) {
	return nil, nil
}
func (m *mockStorageEnt) GetStorageByURL(ctx context.Context, url string) (*ent.StorageEntity, error) {
	return nil, nil
}
func (m *mockStorageEnt) GetStorageByEmail(ctx context.Context, email string) (*ent.StorageEntity, error) {
	return nil, nil
}
func (m *mockStorageEnt) UpdateStorage(ctx context.Context, id uuid.UUID, storage entitiesdto.UpdateStorage) (*ent.StorageEntity, error) {
	return nil, nil
}
func (m *mockStorageEnt) DeleteStorage(ctx context.Context, id uuid.UUID) error {
	delete(m.storages, id)
	return nil
}

type mockImageEnt struct {
	images map[uuid.UUID]*ent.ImageEntity
}

func newMockImageEnt() *mockMockImageEnt {
	return &mockMockImageEnt{images: make(map[uuid.UUID]*ent.ImageEntity)}
}

func (m *mockMockImageEnt) CreateImage(ctx context.Context, in entitiesdto.CreateImage) (*ent.ImageEntity, error) {
	id := uuid.New()
	img := &ent.ImageEntity{
		ID:        id,
		IsPrivate: in.IsPrivate != nil && *in.IsPrivate,
		CreatedAt: time.Now(),
	}
	if in.UserID != nil {
		uid, _ := uuid.Parse(*in.UserID)
		img.UserID = &uid
	}
	if in.StorageID != nil {
		sid, _ := uuid.Parse(*in.StorageID)
		img.StorageID = sid
	}
	m.images[id] = img
	return img, nil
}
func (m *mockMockImageEnt) GetImageByID(ctx context.Context, id uuid.UUID) (*ent.ImageEntity, error) {
	img, ok := m.images[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	return img, nil
}
func (m *mockMockImageEnt) GetImageByStorageID(ctx context.Context, storageID uuid.UUID) (*ent.ImageEntity, error) {
	for _, img := range m.images {
		if img.StorageID == storageID {
			return img, nil
		}
	}
	return nil, sql.ErrNoRows
}
func (m *mockMockImageEnt) GetImagesByUserID(ctx context.Context, userID uuid.UUID) ([]*ent.ImageEntity, error) {
	var list []*ent.ImageEntity
	for _, img := range m.images {
		if img.UserID != nil && *img.UserID == userID {
			list = append(list, img)
		}
	}
	return list, nil
}
func (m *mockMockImageEnt) UpdateImage(ctx context.Context, id uuid.UUID, image entitiesdto.UpdateImage) (*ent.ImageEntity, error) {
	return nil, nil
}
func (m *mockMockImageEnt) DeleteImage(ctx context.Context, id uuid.UUID) error {
	delete(m.images, id)
	return nil
}
func (m *mockMockImageEnt) ListExpiredImages(ctx context.Context, before time.Time) ([]*ent.ImageEntity, error) {
	return nil, nil
}
func (m *mockMockImageEnt) ListAllImages(ctx context.Context, limit int, offset int) ([]*ent.ImageEntity, int, error) {
	return nil, 0, nil
}
func (m *mockMockImageEnt) GetGuestStats(ctx context.Context) (int, int64, error) {
	return 0, 0, nil
}
func (m *mockMockImageEnt) GetUniqueGuestIPCount(ctx context.Context, since time.Time) (int, error) {
	return 0, nil
}

type mockMockImageEnt = mockImageEnt

func TestGetFileMetadata_OwnershipAndPrivacy(t *testing.T) {
	storeEnt := newMockStorageEnt()
	imageEnt := newMockImageEnt()

	svc := newService(&Options{
		Config:   &config.Config[Config]{Val: &Config{}},
		tracer:   otel.Tracer("test"),
		store:    storeEnt,
		imageEnt: imageEnt,
	})

	ctx := context.Background()

	urlStr := "http://localhost:8080/file/abc12345.png"
	mimeStr := "image/png"
	st, _ := storeEnt.CreateStorage(ctx, entitiesdto.CreateStorage{
		ShortCode: "abc12345",
		URL:       &urlStr,
		MIMEType:  &mimeStr,
		FileSize:  1024,
	})

	ownerID := uuid.New()
	otherUserID := uuid.New()

	ownerStr := ownerID.String()
	storageStr := st.ID.String()
	isPrivate := true

	// Create image associated with storage as PRIVATE owned by ownerID
	_, _ = imageEnt.CreateImage(ctx, entitiesdto.CreateImage{
		UserID:    &ownerStr,
		StorageID: &storageStr,
		IsPrivate: &isPrivate,
	})

	// Case A: Owner requests private file -> Success
	res, err := svc.GetFile(ctx, st.ID, &ownerID)
	if err != nil {
		t.Fatalf("expected owner to access private file, got error: %v", err)
	}
	if res.ID != st.ID {
		t.Errorf("expected storage ID %s, got %s", st.ID, res.ID)
	}

	// Case B: Non-owner requests private file -> Unauthorized
	_, err = svc.GetFile(ctx, st.ID, &otherUserID)
	if err == nil {
		t.Error("expected error when non-owner accesses private file, got nil")
	}
}

func TestDeleteFile_OwnershipCheck(t *testing.T) {
	storeEnt := newMockStorageEnt()
	imageEnt := newMockImageEnt()

	svc := newService(&Options{
		Config:   &config.Config[Config]{Val: &Config{}},
		tracer:   otel.Tracer("test"),
		store:    storeEnt,
		imageEnt: imageEnt,
	})

	ctx := context.Background()

	delUrl := "http://localhost:8080/file/del12345.png"
	st, _ := storeEnt.CreateStorage(ctx, entitiesdto.CreateStorage{
		ShortCode: "del12345",
		URL:       &delUrl,
	})

	ownerID := uuid.New()
	otherUserID := uuid.New()

	ownerStr := ownerID.String()
	storageStr := st.ID.String()

	_, _ = imageEnt.CreateImage(ctx, entitiesdto.CreateImage{
		UserID:    &ownerStr,
		StorageID: &storageStr,
	})

	// Case A: Non-owner attempts to delete file -> Fails with error
	err := svc.DeleteFile(ctx, st.ID, otherUserID)
	if err == nil {
		t.Error("expected error when non-owner attempts to delete file, got nil")
	}

	// Case B: Owner deletes file -> Succeeds
	err = svc.DeleteFile(ctx, st.ID, ownerID)
	if err != nil {
		t.Errorf("expected owner to delete file successfully, got: %v", err)
	}
}
