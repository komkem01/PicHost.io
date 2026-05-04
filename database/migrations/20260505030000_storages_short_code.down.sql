SET statement_timeout = 0;

--bun:split

DROP INDEX IF EXISTS "idx_storages_short_code";

--bun:split

ALTER TABLE "storages"
DROP COLUMN IF EXISTS "short_code";
