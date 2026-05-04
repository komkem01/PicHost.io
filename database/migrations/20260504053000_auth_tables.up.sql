SET statement_timeout = 0;

--bun:split

CREATE TABLE IF NOT EXISTS "auth_sessions" (
	"id" uuid NOT NULL DEFAULT gen_random_uuid(),
	"user_id" uuid NOT NULL,
	"refresh_token_hash" text NOT NULL,
	"user_agent" text,
	"ip_address" varchar(64),
	"expires_at" timestamp NOT NULL,
	"revoked_at" timestamp,
	"created_at" timestamp NOT NULL DEFAULT current_timestamp,
	"updated_at" timestamp NOT NULL DEFAULT current_timestamp,
	PRIMARY KEY ("id"),
	CONSTRAINT "fk_auth_sessions_user_id" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS "idx_auth_sessions_user_id" ON "auth_sessions" ("user_id");
CREATE INDEX IF NOT EXISTS "idx_auth_sessions_refresh_token_hash" ON "auth_sessions" ("refresh_token_hash");
CREATE INDEX IF NOT EXISTS "idx_auth_sessions_expires_at" ON "auth_sessions" ("expires_at");
CREATE INDEX IF NOT EXISTS "idx_auth_sessions_revoked_at" ON "auth_sessions" ("revoked_at");

CREATE OR REPLACE FUNCTION set_auth_sessions_updated_at()
RETURNS trigger AS $$
BEGIN
	NEW.updated_at = current_timestamp;
	RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_auth_sessions_set_updated_at ON "auth_sessions";
CREATE TRIGGER trg_auth_sessions_set_updated_at
BEFORE UPDATE ON "auth_sessions"
FOR EACH ROW
EXECUTE FUNCTION set_auth_sessions_updated_at();

COMMENT ON TABLE "auth_sessions" IS 'ตารางเซสชันสำหรับ refresh token';
COMMENT ON COLUMN "auth_sessions"."id" IS 'รหัสเซสชัน';
COMMENT ON COLUMN "auth_sessions"."user_id" IS 'อ้างอิงไปยัง users.id';
COMMENT ON COLUMN "auth_sessions"."refresh_token_hash" IS 'ค่าแฮชของ refresh token';
COMMENT ON COLUMN "auth_sessions"."user_agent" IS 'ข้อมูล User-Agent ของ client';
COMMENT ON COLUMN "auth_sessions"."ip_address" IS 'IP ของ client ขณะออก token';
COMMENT ON COLUMN "auth_sessions"."expires_at" IS 'วันหมดอายุของ refresh token';
COMMENT ON COLUMN "auth_sessions"."revoked_at" IS 'เวลาที่ revoke token';
COMMENT ON COLUMN "auth_sessions"."created_at" IS 'วันที่สร้างข้อมูล';
COMMENT ON COLUMN "auth_sessions"."updated_at" IS 'วันที่แก้ไขล่าสุด';

--bun:split

CREATE TABLE IF NOT EXISTS "oauth_accounts" (
	"id" uuid NOT NULL DEFAULT gen_random_uuid(),
	"user_id" uuid NOT NULL,
	"provider" varchar(50) NOT NULL,
	"provider_user_id" varchar(255) NOT NULL,
	"email" varchar(255),
	"created_at" timestamp NOT NULL DEFAULT current_timestamp,
	"updated_at" timestamp NOT NULL DEFAULT current_timestamp,
	PRIMARY KEY ("id"),
	CONSTRAINT "fk_oauth_accounts_user_id" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS "uq_oauth_accounts_provider_user_id" ON "oauth_accounts" ("provider", "provider_user_id");
CREATE INDEX IF NOT EXISTS "idx_oauth_accounts_user_id" ON "oauth_accounts" ("user_id");

CREATE OR REPLACE FUNCTION set_oauth_accounts_updated_at()
RETURNS trigger AS $$
BEGIN
	NEW.updated_at = current_timestamp;
	RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_oauth_accounts_set_updated_at ON "oauth_accounts";
CREATE TRIGGER trg_oauth_accounts_set_updated_at
BEFORE UPDATE ON "oauth_accounts"
FOR EACH ROW
EXECUTE FUNCTION set_oauth_accounts_updated_at();

COMMENT ON TABLE "oauth_accounts" IS 'ตารางเชื่อมบัญชี OAuth';
COMMENT ON COLUMN "oauth_accounts"."id" IS 'รหัสความเชื่อมโยงบัญชี OAuth';
COMMENT ON COLUMN "oauth_accounts"."user_id" IS 'อ้างอิงไปยัง users.id';
COMMENT ON COLUMN "oauth_accounts"."provider" IS 'ผู้ให้บริการ OAuth เช่น google';
COMMENT ON COLUMN "oauth_accounts"."provider_user_id" IS 'รหัสผู้ใช้จากผู้ให้บริการ';
COMMENT ON COLUMN "oauth_accounts"."email" IS 'อีเมลที่ได้จากผู้ให้บริการ';
COMMENT ON COLUMN "oauth_accounts"."created_at" IS 'วันที่สร้างข้อมูล';
COMMENT ON COLUMN "oauth_accounts"."updated_at" IS 'วันที่แก้ไขล่าสุด';
