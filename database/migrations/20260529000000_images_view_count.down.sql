SET statement_timeout = 0;

--bun:split

ALTER TABLE "images" DROP COLUMN IF EXISTS "view_count";
