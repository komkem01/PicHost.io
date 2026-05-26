SET statement_timeout = 0;

--bun:split

DROP INDEX IF EXISTS idx_users_plan_expires_at;

--bun:split

ALTER TABLE "users"
    DROP COLUMN IF EXISTS "plan_cancelled_at",
    DROP COLUMN IF EXISTS "plan_expires_at";
