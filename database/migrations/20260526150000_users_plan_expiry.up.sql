SET statement_timeout = 0;

--bun:split

ALTER TABLE "users"
    ADD COLUMN IF NOT EXISTS "plan_expires_at" timestamptz,
    ADD COLUMN IF NOT EXISTS "plan_cancelled_at" timestamptz;

--bun:split

CREATE INDEX IF NOT EXISTS idx_users_plan_expires_at ON "users" ("plan_expires_at")
    WHERE "plan_expires_at" IS NOT NULL;

--bun:split

COMMENT ON COLUMN "users"."plan_expires_at" IS 'วันที่แผนการใช้งานหมดอายุ (null = ไม่มีวันหมดอายุ สำหรับ Free plan หรือแผนที่ไม่มีการกำหนดอายุ)';
COMMENT ON COLUMN "users"."plan_cancelled_at" IS 'วันที่ผู้ใช้งานขอยกเลิกการต่ออายุ แผนยังใช้งานได้จนถึง plan_expires_at';
