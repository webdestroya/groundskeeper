-- +goose up
CREATE TABLE IF NOT EXISTS "ratings" (
  "id" INTEGER PRIMARY KEY,
  "book_id" INTEGER,
  "user_id" INTEGER,
  "stars" INTEGER
);

-- +goose down
DROP TABLE IF EXISTS "ratings";