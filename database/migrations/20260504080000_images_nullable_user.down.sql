SET statement_timeout = 0;

--bun:split

ALTER TABLE "images" ALTER COLUMN "user_id" SET NOT NULL;

--bun:split

COMMENT ON COLUMN "images"."user_id" IS 'ผูกกับ users.id เสมอ ไม่ว่าจะเป็นสมาชิกหรือ Guest';
