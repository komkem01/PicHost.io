SET statement_timeout = 0;

--bun:split

ALTER TABLE "images" ALTER COLUMN "user_id" DROP NOT NULL;

--bun:split

COMMENT ON COLUMN "images"."user_id" IS 'ผูกกับ users.id ถ้าเป็นสมาชิก หรือ NULL ถ้าเป็น Guest ที่ไม่มีบัญชี';
