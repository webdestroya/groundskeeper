package gooserunner

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/webdestroya/groundskeeper/internal/awsclients/ssmclient"
	"github.com/webdestroya/groundskeeper/internal/config"
	"github.com/webdestroya/x/logger"
	_ "modernc.org/sqlite"
)

func resolveDatabaseUrl(ctx context.Context) (string, error) {
	databaseUrl := config.DatabaseURL()

	log := logger.Ctx(ctx)

	if databaseUrl == "" {
		return "", errors.New("database url is missing")
	}

	// assume SSM
	if strings.HasPrefix(databaseUrl, `/`) || strings.HasPrefix(databaseUrl, `ssm:`) {
		databaseUrl = strings.TrimPrefix(databaseUrl, `ssm:`)
		log.Info().Str("name", databaseUrl).Msg("Looking up SSM Parameter")

		if client, err := ssmclient.New(ctx); err == nil {
			resp, err := client.GetParameter(ctx, &ssm.GetParameterInput{
				Name:           &databaseUrl,
				WithDecryption: new(true),
			})
			if err != nil {
				return "", fmt.Errorf(`unable to resolve SSM parameter: %w`, err)
			}

			databaseUrl = *resp.Parameter.Value
		} else {

			return "", fmt.Errorf(`unable to resolve AWS config: %w`, err)
		}
	}

	return databaseUrl, nil
}

func connect(ctx context.Context) (*sql.DB, goose.Dialect, error) {
	databaseUrl, err := resolveDatabaseUrl(ctx)
	if err != nil {
		return nil, goose.DialectCustom, err
	}

	uri, err := url.Parse(databaseUrl)
	if err != nil {
		logger.Ctx(ctx).Error().Err(err).Str("uri", uri.Redacted()).Msg("invalid database url")
		return nil, goose.DialectCustom, err
	}

	logger.Ctx(ctx).Info().Str("uri", uri.Redacted()).Msg("connecting to database")

	switch {

	case databaseUrl == ":memory:" || strings.HasPrefix(databaseUrl, `sqlite`) || strings.HasPrefix(databaseUrl, `file:`):
		sqlDB, err := sql.Open("sqlite", databaseUrl)
		return sqlDB, goose.DialectSQLite3, err

	case strings.HasPrefix(databaseUrl, `mysql`) || strings.HasPrefix(databaseUrl, `maria`):
		// stdlib.OpenDB()
		return nil, goose.DialectMySQL, errors.New("MYSQL SUPPORT NOT ENABLED YET")

	default:
		sqlDB, err := connectPG(ctx, databaseUrl)
		return sqlDB, goose.DialectPostgres, err
	}

}

func connectPG(_ context.Context, databaseUrl string) (*sql.DB, error) {

	// log := logger.Ctx(ctx)

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
