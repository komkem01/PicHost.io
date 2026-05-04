SET statement_timeout = 0;

--bun:split

CREATE TABLE IF NOT EXISTS "storages" (
	"id" uuid NOT NULL DEFAULT gen_random_uuid(),
	"provider" varchar(100) NOT NULL DEFAULT 'Railway',
	"path" text,
	"url" text,
	"file_size" bigint NOT NULL,
	"mime_type" varchar(255),
	"created_at" timestamp NOT NULL DEFAULT current_timestamp,
	PRIMARY KEY ("id")
);

CREATE INDEX IF NOT EXISTS "idx_storages_provider" ON "storages" ("provider");
CREATE INDEX IF NOT EXISTS "idx_storages_created_at" ON "storages" ("created_at");

COMMENT ON TABLE "storages" IS 'ตารางพื้นที่จัดเก็บ';
COMMENT ON COLUMN "storages"."id" IS 'รหัสพื้นที่จัดเก็บ';
COMMENT ON COLUMN "storages"."provider" IS 'ผู้ให้บริการ';
COMMENT ON COLUMN "storages"."path" IS 'พาธที่จัดเก็บไฟล์';
COMMENT ON COLUMN "storages"."url" IS 'ลิงก์เข้าถึงไฟล์';
COMMENT ON COLUMN "storages"."file_size" IS 'ขนาดไฟล์ (Byte)';
COMMENT ON COLUMN "storages"."mime_type" IS 'ประเภทไฟล์ (MIME Type)';
COMMENT ON COLUMN "storages"."created_at" IS 'วันที่สร้างข้อมูล';
