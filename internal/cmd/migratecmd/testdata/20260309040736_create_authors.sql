-- +goose up
CREATE TABLE IF NOT EXISTS "authors" (
  "id" INTEGER PRIMARY KEY,
  "name" TEXT
);

-- +goose down
DROP TABLE IF EXISTS "authors";