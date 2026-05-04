SET statement_timeout = 0;

--bun:split

DO $$
BEGIN
	CREATE TYPE plan_type AS ENUM ('Free', 'Basic', 'Pro', 'Enterprise');
EXCEPTION
	WHEN duplicate_object THEN NULL;
END$$;

--bun:split

CREATE TABLE IF NOT EXISTS "users" (
	"id" uuid NOT NULL DEFAULT gen_random_uuid(),
	"email" varchar(255) UNIQUE,
	"password" text,
	"username" varchar(255),
	"plan" plan_type NOT NULL DEFAULT 'Free',
	"is_active" boolean NOT NULL DEFAULT true,
	"is_guest" boolean NOT NULL DEFAULT true,
	"created_at" timestamp NOT NULL DEFAULT current_timestamp,
	"updated_at" timestamp NOT NULL DEFAULT current_timestamp,
	PRIMARY KEY ("id")
);

CREATE OR REPLACE FUNCTION set_users_updated_at()
RETURNS trigger AS $$
BEGIN
	NEW.updated_at = current_timestamp;
	RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_users_set_updated_at ON "users";
CREATE TRIGGER trg_users_set_updated_at
BEFORE UPDATE ON "users"
FOR EACH ROW
EXECUTE FUNCTION set_users_updated_at();

COMMENT ON TABLE "users" IS 'ตารางผู้ใช้งาน';
COMMENT ON COLUMN "users"."id" IS 'รหัสผู้ใช้งาน';
COMMENT ON COLUMN "users"."email" IS 'อีเมล (Guest จะเป็น null)';
COMMENT ON COLUMN "users"."password" IS 'รหัสผ่าน (Guest จะเป็น null)';
COMMENT ON COLUMN "users"."username" IS 'ชื่อผู้ใช้งาน (Guest อาจเป็น null หรือชื่อสุ่ม)';
COMMENT ON COLUMN "users"."plan" IS 'แผนการใช้งาน';
COMMENT ON COLUMN "users"."is_active" IS 'สถานะบัญชี';
COMMENT ON COLUMN "users"."is_guest" IS 'ถ้าสร้างใหม่โดยไม่ได้สมัคร ให้เป็น true';
COMMENT ON COLUMN "users"."created_at" IS 'วันที่สร้างข้อมูล';
COMMENT ON COLUMN "users"."updated_at" IS 'วันที่แก้ไขล่าสุด';
