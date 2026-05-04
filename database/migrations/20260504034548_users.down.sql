SET statement_timeout = 0;

--bun:split

DROP TABLE IF EXISTS "users" CASCADE;

--bun:split

DROP FUNCTION IF EXISTS set_users_updated_at();

--bun:split

DROP TYPE IF EXISTS plan_type;
