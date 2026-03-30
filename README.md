# Groundskeeper

Got a database in your serverless app? Using a language without a built in migration system? This is the tool for you.

Migrations are tracked in a table called `groundskeeper_migrations` that will be added to your database.

## Installation

```shell
brew install webdestroya/tap/groundskeeper
```

## Writing your migrations
Pick a folder in your repository for your migrations. It will contain a list of `.sql` files for each migration. Do not nest files within this directory. Only `*.sql` files will be processed, all others are ignored.

Files should be named: `YYYYMMDDHHMMSS_name_of_migration.sql`

```
db/migrations
├── 20260309040736_create_authors.sql
├── 20260309041055_create_books.sql
├── 20260309041100_create_ratings.sql
├── 20260314203040_add_dob_to_authors.sql
├── 20260314232010_add_isbn_to_books.sql
└── 20260314235159_add_active_to_books.sql
```

View [example migrations](./doc/examples.md).

## Available Commands

### `init`
Initializes a project to use groundskeeper. Sets a config file and the path for migrations.

```shell
groundskeeper init
```

### `create` (or `new`)
Creates a new migration script in your project

```shell
groundskeeper create create_users_table
groundskeeper new create users table
groundskeeper new CreateUsersTable
```

### `migrate`
Applies all pending migrations

### `status`
Shows which migrations have been applied, and which ones need to be applied.

### `rollback`
Rollback the latest migration.

_Note: Migrations are rolled back in the order they were applied, not the order they are sorted by._

### `up`
Will apply the next available migration, and then stops.

### `pull`
Used only in a remote console session. Lets you pull migrations to the container, and you can run status/rollback/migrate manually.

If migrations have been pulled, then the above commands do not require the `from` flag.

## Running migrations
In your ECS pipeline, you should fire off the following task:

```shell
groundskeeper migrate \
  --from github.com/OrgName/RepoName/db/migrations@v1.2.3 \
  --dburl ssm:/ecsdeployer/secrets/yourapp/DATABASE_URL
```

## Specifying Parameters

### Database URLs (`--dburl`)
This can take multiple forms:

* a SecretsManager secret ARN. (You can specify a JSON key to have that value returned)
* a SSM Parameter ARN.
* a string starting with `ssm:` will be resolved as an SSM parameter
* a string starting with `/` will be assumed to be an SSM parameter
* all other values will be returned as-is

These are all valid `--dburl` values:
```
arn:aws:secretsmanager:region:account_id:secret:SecretName:json-key:version-stage:version-id
arn:aws:secretsmanager:region:account_id:secret:SecretName:json-key:version-stage
arn:aws:secretsmanager:region:account_id:secret:SecretName:json-key
arn:aws:secretsmanager:region:account_id:secret:SecretName
arn:aws:ssm:region:account_id:parameter/path/to/param
ssm:/path/to/param
/path/to/param
postgres://user:pass@localhost:5432/dbname
```

### Location Specifier (`--from`)

The `--from` flag takes a location specifier that has the following format:

```
[REPOSITORY]/[PATH TO MIGRATIONS]@[COMMITTISH]
```

Examples:

```
github.com/OrgName/Repo/db/migrations@v1.2.3
OrgName/Repo/db/migrations@v1.2.3
```

Both of these reference the `OrgName/Repo` repository. The migrations will be in the folder `db/migrations/*.sql` and the `v1.2.3` tag should be pulled.

## References
* [Goose: Annotations](https://pressly.github.io/goose/documentation/annotations/)