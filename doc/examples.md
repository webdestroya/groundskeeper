# Example Migrations

## Basic
```sql
-- +goose up
SELECT 'up SQL query';

-- +goose down
SELECT 'down SQL query';
```

## Complex Queries
Statements are split on semicolons (`;`). If you have many queries, or are defining a function (don't), then you will need to group those statements together:

```sql
-- +goose up
CREATE TABLE users (
    id int NOT NULL PRIMARY KEY,
    username text,
    name text,
    surname text
);

-- +goose statementbegin
INSERT INTO "users" ("id", "username") VALUES (1, 'gallant_almeida7');
INSERT INTO "users" ("id", "username") VALUES (2, 'brave_spence8');
.
.
INSERT INTO "users" ("id", "username") VALUES (99999, 'jovial_chaum1');
INSERT INTO "users" ("id", "username") VALUES (100000, 'goofy_ptolemy0');
-- +goose statementend

-- +goose down
DROP TABLE users;
```

## No Transaction
This is the same as Rails' `disable_ddl_transaction!`

```sql
-- +goose no transaction

-- +goose up
CREATE INDEX CONCURRENTLY ON users (user_id);

-- +goose down
DROP INDEX users_user_id_idx;
```

---

For more information, check out the [goose annotation documentation](https://pressly.github.io/goose/documentation/annotations/).