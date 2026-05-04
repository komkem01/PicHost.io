SET statement_timeout = 0;

--bun:split

DROP TABLE IF EXISTS "oauth_accounts" CASCADE;

--bun:split

DROP FUNCTION IF EXISTS set_oauth_accounts_updated_at();

--bun:split

DROP TABLE IF EXISTS "auth_sessions" CASCADE;

--bun:split

DROP FUNCTION IF EXISTS set_auth_sessions_updated_at();
