SET statement_timeout = 0;

--bun:split

ALTER TABLE "images" ADD COLUMN IF NOT EXISTS "view_count" bigint NOT NULL DEFAULT 0;

COMMENT ON COLUMN "images"."view_count" IS 'จำนวนครั้งที่รูปภาพถูกเปิดดู';
