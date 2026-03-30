package gooserunner

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/webdestroya/groundskeeper/internal/config"
	"github.com/webdestroya/groundskeeper/internal/utils/secretresolver"
	"github.com/webdestroya/x/logger"
	_ "modernc.org/sqlite"
)

func resolveDatabaseUrl(ctx context.Context) (string, error) {
	databaseUrl := config.DatabaseURL()

	if databaseUrl == "" {
		return "", errors.New("database url is missing")
	}

	return secretresolver.Resolve(ctx, databaseUrl)
}

func connect(ctx context.Context) (*sql.DB, goose.Dialect, error) {
	databaseUrl, err := resolveDatabaseUrl(ctx)
	if err != nil {
		return nil, goose.DialectCustom, err
	}

	if uri, err := url.Parse(databaseUrl); err == nil {
		logger.Ctx(ctx).Info().Str("database_url", uri.Redacted()).Send()
	}

	logger.Ctx(ctx).Info().Msg("connecting to database")

	scheme, dsn, ok := strings.Cut(databaseUrl, `://`)
	if !ok {
		return nil, goose.DialectCustom, fmt.Errorf("invalid database url provided")
	}

	switch {

	case strings.HasPrefix(scheme, `sqlite`):
		sqlDB, err := sql.Open("sqlite", dsn)
		return sqlDB, goose.DialectSQLite3, err

	case scheme == `mysql` || scheme == `maria`:
		sqlDB, err := sql.Open("mysql", dsn)
		return sqlDB, goose.DialectSQLite3, err

	case strings.HasPrefix(scheme, `postgres`):
		sqlDB, err := connectPG(databaseUrl)
		return sqlDB, goose.DialectPostgres, err

	default:
		return nil, goose.DialectCustom, fmt.Errorf("unsupported database: %s", scheme)
	}

}

func connectPG(databaseUrl string) (*sql.DB, error) {
	cfg, err := pgx.ParseConfig(databaseUrl)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}

	if len(cfg.RuntimeParams) == 0 {
		cfg.RuntimeParams = make(map[string]string)
	}
	cfg.RuntimeParams[`application_name`] = config.ServiceName

	sqlDB := stdlib.OpenDB(*cfg)

	return sqlDB, nil
}

// // Deprecated: Use [connectPG] instead
// func connectPGPool(ctx context.Context, databaseUrl string) (*sql.DB, error) {

// 	// log := logger.Ctx(ctx)

// 	poolConfig, err := pgxpool.ParseConfig(databaseUrl)
// 	if err != nil {
// 		return nil, fmt.Errorf("parse database url: %w", err)
// 	}
// 	if len(poolConfig.ConnConfig.RuntimeParams) == 0 {
// 		poolConfig.ConnConfig.RuntimeParams = make(map[string]string)
// 	}
// 	poolConfig.ConnConfig.RuntimeParams[`application_name`] = config.ServiceName
// 	poolConfig.MaxConns = 4

// 	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
// 	if err != nil {
// 		return nil, fmt.Errorf("create pool: %w", err)
// 	}

// 	if err := pool.Ping(ctx); err != nil {
// 		pool.Close()
// 		return nil, fmt.Errorf("ping database: %w", err)
// 	}

// 	sqlDB := stdlib.OpenDBFromPool(pool)

// 	return sqlDB, nil
// }
