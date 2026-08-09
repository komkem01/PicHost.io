SET statement_timeout = 0;

--bun:split

DROP TABLE IF EXISTS email_verification_tokens;

--bun:split

ALTER TABLE "users" DROP COLUMN IF EXISTS "email_verified_at";
