-- user_quotas: ติดตามพื้นที่และจำนวนรูปภาพของแต่ละผู้ใช้
CREATE TABLE IF NOT EXISTS user_quotas (
    id                 UUID        NOT NULL DEFAULT gen_random_uuid(),
    user_id            UUID        NOT NULL,
    used_storage_bytes BIGINT      NOT NULL DEFAULT 0,
    image_count        INTEGER     NOT NULL DEFAULT 0,
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT user_quotas_pkey        PRIMARY KEY (id),
    CONSTRAINT user_quotas_user_id_key UNIQUE (user_id),
    CONSTRAINT user_quotas_user_id_fk  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT user_quotas_storage_nneg CHECK (used_storage_bytes >= 0),
    CONSTRAINT user_quotas_count_nneg  CHECK (image_count >= 0)
);

CREATE INDEX IF NOT EXISTS user_quotas_user_id_idx ON user_quotas (user_id);

-- trigger: อัปเดต updated_at อัตโนมัติ
CREATE OR REPLACE FUNCTION update_user_quotas_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER user_quotas_updated_at_trigger
    BEFORE UPDATE ON user_quotas
    FOR EACH ROW EXECUTE FUNCTION update_user_quotas_updated_at();

-- เพิ่มคอลัมน์ expires_at ในตาราง images สำหรับ guest (ลบอัตโนมัติหลัง 24 ชม.)
ALTER TABLE images ADD COLUMN IF NOT EXISTS expires_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS images_expires_at_idx ON images (expires_at) WHERE expires_at IS NOT NULL;

COMMENT ON TABLE  user_quotas                       IS 'ตารางติดตามการใช้งานพื้นที่จัดเก็บของผู้ใช้แต่ละคน';
COMMENT ON COLUMN user_quotas.id                    IS 'รหัสอ้างอิงของ quota record';
COMMENT ON COLUMN user_quotas.user_id               IS 'รหัสผู้ใช้ที่เชื่อมโยงกับ quota นี้';
COMMENT ON COLUMN user_quotas.used_storage_bytes     IS 'จำนวนไบต์ที่ใช้งานอยู่ในปัจจุบัน';
COMMENT ON COLUMN user_quotas.image_count            IS 'จำนวนรูปภาพที่อัปโหลดทั้งหมด';
COMMENT ON COLUMN user_quotas.updated_at             IS 'เวลาที่ข้อมูล quota ถูกอัปเดตล่าสุด';
COMMENT ON COLUMN images.expires_at                  IS 'เวลาหมดอายุของรูปภาพ ใช้สำหรับรูปของ guest ที่จะถูกลบหลัง 24 ชั่วโมง';
