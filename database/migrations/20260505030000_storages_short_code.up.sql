SET statement_timeout = 0;

--bun:split

ALTER TABLE "storages"
ADD COLUMN IF NOT EXISTS "short_code" varchar(8);

--bun:split

UPDATE "storages"
SET "short_code" = substring(replace(id::text, '-', ''), 1, 8)
WHERE "short_code" IS NULL OR "short_code" = '';

--bun:split

CREATE UNIQUE INDEX IF NOT EXISTS "idx_storages_short_code" ON "storages" ("short_code");

--bun:split

ALTER TABLE "storages"
ALTER COLUMN "short_code" SET NOT NULL;

--bun:split

COMMENT ON COLUMN "storages"."short_code" IS 'โค้ดสั้นสำหรับลิงก์สาธารณะของไฟล์ (6-8 ตัวอักษร)';
