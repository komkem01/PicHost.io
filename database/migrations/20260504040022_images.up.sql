SET statement_timeout = 0;

--bun:split

CREATE TABLE IF NOT EXISTS "images" (
	"id" uuid NOT NULL DEFAULT gen_random_uuid(),
	"user_id" uuid NOT NULL,
	"storage_id" uuid NOT NULL,
	"is_private" boolean NOT NULL DEFAULT false,
	"created_at" timestamp NOT NULL DEFAULT current_timestamp,
	PRIMARY KEY ("id"),
	CONSTRAINT "fk_images_user_id" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE,
	CONSTRAINT "fk_images_storage_id" FOREIGN KEY ("storage_id") REFERENCES "storages" ("id") ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS "idx_images_user_id" ON "images" ("user_id");
CREATE INDEX IF NOT EXISTS "idx_images_storage_id" ON "images" ("storage_id");
CREATE INDEX IF NOT EXISTS "idx_images_created_at" ON "images" ("created_at");

COMMENT ON TABLE "images" IS 'ตารางรูปภาพ';
COMMENT ON COLUMN "images"."id" IS 'รหัสรูปภาพ';
COMMENT ON COLUMN "images"."user_id" IS 'ผูกกับ users.id เสมอ ไม่ว่าจะเป็นสมาชิกหรือ Guest';
COMMENT ON COLUMN "images"."storage_id" IS 'อ้างอิงไปยัง storages.id';
COMMENT ON COLUMN "images"."is_private" IS 'สถานะการมองเห็นรูปภาพ';
COMMENT ON COLUMN "images"."created_at" IS 'วันที่สร้างข้อมูล';
