package entities

import (
	"context"
	"strings"
	"time"

	entitiesdto "pichost.io/app/modules/entities/dto"
	"pichost.io/app/modules/entities/ent"
	entitiesinf "pichost.io/app/modules/entities/inf"
)

var _ entitiesinf.LegalDocumentEntity = (*Service)(nil)

const defaultTermsContent = `By accessing or using PicHost.io, you agree to be bound by these Terms of Service.
การเข้าใช้บริการ PicHost.io ถือว่าท่านได้ยอมรับข้อตกลงและเงื่อนไขการให้บริการนี้แล้ว

1. Acceptance of Terms / การยอมรับข้อตกลง
By accessing or using PicHost.io, you agree to be bound by these Terms of Service.
การเข้าใช้บริการ PicHost.io ถือว่าท่านได้ยอมรับข้อตกลงและเงื่อนไขการให้บริการนี้แล้ว

2. Acceptable Use & Content Policy / นโยบายการใช้งานและเนื้อหา
Illegal content, malware, phishing, and copyrighted content violations are strictly prohibited.
ห้ามใช้อัปโหลดไฟล์ที่ผิดกฎหมาย สื่อลามกอนาจาร ไวรัส/มัลแวร์ หรือไฟล์ละเมิดลิขสิทธิ์

3. Account Termination / การยกเลิกบัญชี
We reserve the right to suspend or terminate accounts that violate our terms of service without prior notice.`

const defaultPrivacyContent = `We collect account email, username, uploaded image files, and security access logs strictly for operating PicHost.io.
เราจัดเก็บอีเมล ชื่อผู้ใช้ ภาพที่ถูกอัปโหลด และประวัติบันทึกความปลอดภัยเฉพาะที่จำเป็นสำหรับการให้บริการเท่านั้น

1. Information We Collect / ข้อมูลที่เราจัดเก็บ
We collect account email, username, uploaded image files, and security access logs strictly for operating PicHost.io.
เราจัดเก็บอีเมล ชื่อผู้ใช้ ภาพที่ถูกอัปโหลด และประวัติบันทึกความปลอดภัยเฉพาะที่จำเป็นสำหรับการให้บริการเท่านั้น

2. Data Protection & PDPA / การคุ้มครองข้อมูลส่วนบุคคล
We protect your data according to Thailand PDPA standards. Your personal data is never sold to third parties.`

func (s *Service) EnsureTableAndDefaults(ctx context.Context) {
	createTableSQL := `CREATE TABLE IF NOT EXISTS legal_documents (
		key VARCHAR(64) PRIMARY KEY,
		title VARCHAR(255) NOT NULL,
		content TEXT NOT NULL,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);`
	_, _ = s.db.ExecContext(ctx, createTableSQL)

	// Ensure 'terms' row exists and has content
	var terms ent.LegalDocumentEntity
	err := s.db.NewSelect().Model(&terms).Where("key = 'terms'").Scan(ctx)
	if err != nil || strings.TrimSpace(terms.Content) == "" {
		now := time.Now()
		_, _ = s.db.NewInsert().
			Model(&ent.LegalDocumentEntity{
				Key:       "terms",
				Title:     "Terms of Service / ข้อตกลงและเงื่อนไขการให้บริการ",
				Content:   defaultTermsContent,
				UpdatedAt: now,
				CreatedAt: now,
			}).
			On("CONFLICT (key) DO UPDATE").
			Set("title = EXCLUDED.title").
			Set("content = EXCLUDED.content").
			Exec(ctx)
	}

	// Ensure 'privacy' row exists and has content
	var privacy ent.LegalDocumentEntity
	err = s.db.NewSelect().Model(&privacy).Where("key = 'privacy'").Scan(ctx)
	if err != nil || strings.TrimSpace(privacy.Content) == "" {
		now := time.Now()
		_, _ = s.db.NewInsert().
			Model(&ent.LegalDocumentEntity{
				Key:       "privacy",
				Title:     "Privacy Policy / นโยบายความเป็นส่วนตัว",
				Content:   defaultPrivacyContent,
				UpdatedAt: now,
				CreatedAt: now,
			}).
			On("CONFLICT (key) DO UPDATE").
			Set("title = EXCLUDED.title").
			Set("content = EXCLUDED.content").
			Exec(ctx)
	}
}

func (s *Service) ListLegalDocuments(ctx context.Context) ([]*ent.LegalDocumentEntity, error) {
	s.EnsureTableAndDefaults(ctx)
	var rows []*ent.LegalDocumentEntity
	err := s.db.NewSelect().
		Model(&rows).
		Order("key ASC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *Service) GetLegalDocumentByKey(ctx context.Context, key string) (*ent.LegalDocumentEntity, error) {
	s.EnsureTableAndDefaults(ctx)
	var row ent.LegalDocumentEntity
	err := s.db.NewSelect().
		Model(&row).
		Where("key = ?", strings.TrimSpace(key)).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (s *Service) UpsertLegalDocument(ctx context.Context, input entitiesdto.UpsertLegalDocument) (*ent.LegalDocumentEntity, error) {
	s.EnsureTableAndDefaults(ctx)
	key := strings.TrimSpace(input.Key)
	now := time.Now()
	row := &ent.LegalDocumentEntity{
		Key:       key,
		Title:     strings.TrimSpace(input.Title),
		Content:   input.Content,
		UpdatedAt: now,
		CreatedAt: now,
	}

	_, err := s.db.NewInsert().
		Model(row).
		On("CONFLICT (key) DO UPDATE").
		Set("title = EXCLUDED.title").
		Set("content = EXCLUDED.content").
		Set("updated_at = EXCLUDED.updated_at").
		Exec(ctx)
	if err != nil {
		return nil, err
	}

	return s.GetLegalDocumentByKey(ctx, key)
}

func (s *Service) DeleteLegalDocumentByKey(ctx context.Context, key string) error {
	s.EnsureTableAndDefaults(ctx)
	_, err := s.db.NewDelete().
		Model((*ent.LegalDocumentEntity)(nil)).
		Where("key = ?", strings.TrimSpace(key)).
		Exec(ctx)
	return err
}

