-- +goose up
CREATE TABLE IF NOT EXISTS "books" (
  "id" INTEGER PRIMARY KEY,
  "author_id" INTEGER,
  "title" TEXT
);

-- +goose down
DROP TABLE IF EXISTS "books";