package gooserunner

import (
	"context"
	"log/slog" //nolint:depguard // no choice

	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/lock"
	"github.com/spf13/afero"
	"github.com/webdestroya/groundskeeper/internal/runners"
	"github.com/webdestroya/x/logger"
)

type GooseOperation uint8

const (
	GooseStatus GooseOperation = iota
	GooseUpAll
	GooseUpByOne
	GooseDown
)

const TableName = `groundskeeper_migrations`

type Runner struct {
	Operation    GooseOperation
	MigrationsFS afero.Fs
}

func Run(ctx context.Context, operation GooseOperation, migrationFS afero.Fs) error {
	return New(operation, migrationFS).Run(ctx)
}

func New(operation GooseOperation, migrationFs afero.Fs) runners.Runner {
	return &Runner{
		Operation:    operation,
		MigrationsFS: migrationFs,
	}
}

func (r Runner) Run(ctx context.Context) error {
	log := logger.Ctx(ctx)

	log.Debug().Uint8("op", uint8(r.Operation)).Msg("using goose provider")

	sqlDB, dialect, err := connect(ctx)
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	if err := sqlDB.PingContext(ctx); err != nil {
		log.Error().Err(err).Str("dialect", string(dialect)).Msg("unable to ping db")
		return err
	}

	log.Info().Str("dialect", string(dialect)).Msg("connected to DB")

	fs := afero.NewIOFS(r.MigrationsFS)

	opts := []goose.ProviderOption{
		goose.WithTableName(TableName),
		goose.WithAllowOutofOrder(true),
	}

	if log.Debug().Enabled() {
		opts = append(opts, goose.WithSlog(slog.New(logger.NewSlogHandler(*log))))
		opts = append(opts, goose.WithVerbose(true))
	}

	if dialect == goose.DialectPostgres {
		locker, err := lock.NewPostgresSessionLocker()
		if err != nil {
			return err
		}
		opts = append(opts, goose.WithSessionLocker(locker))
	}

	provider, err := goose.NewProvider(dialect, sqlDB, fs, opts...)
	if err != nil {
		return err
	}

	var (
		result  *goose.MigrationResult
		results []*goose.MigrationResult
	)

	switch r.Operation {
	case GooseUpAll:
		results, err = provider.Up(ctx)

	case GooseUpByOne:
		result, err = provider.UpByOne(ctx)
		results = []*goose.MigrationResult{result}

	case GooseDown:
		result, err = provider.Down(ctx)
		results = []*goose.MigrationResult{result}

	default:
		results, err := provider.Status(ctx)
		if err != nil {
			return err
		}

		for _, result := range results {
			evt := log.Info().Str("state", string(result.State)).Str("name", result.Source.Path)

			if !result.AppliedAt.IsZero() {
				evt = evt.Time("applied_at", result.AppliedAt)
			}

			evt.Send()
		}
		return nil
	}

	for _, res := range results {
		log.Err(res.Error).Msg(res.String())
	}

	if err != nil {
		return err
	}

	return nil
}
