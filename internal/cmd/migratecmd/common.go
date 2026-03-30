package migratecmd

import (
	"errors"
	"fmt"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/webdestroya/groundskeeper/internal/config"
	"github.com/webdestroya/groundskeeper/internal/migrationfs"
	"github.com/webdestroya/groundskeeper/internal/runners/gooserunner"
	"github.com/webdestroya/groundskeeper/internal/runners/pullmigrationsrunner"
	"github.com/webdestroya/groundskeeper/internal/utils/cmdutil"
	"github.com/webdestroya/x/logger"
)

type runner struct {
	fs        afero.Fs
	operation gooserunner.GooseOperation

	from *cmdutil.RepoPathRef

	cfgFile    string
	projectCfg *config.ProjectConfig
}

func (r *runner) setFlags(cmd *cobra.Command) {
	cmdutil.RepoPathRefCmd(cmd, &r.from, false)
	cmdutil.SetupDatabaseFlag(cmd)
	config.AddProjectConfigFlag(cmd, &r.cfgFile)
}

func (r *runner) PreRunE(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	if dburl, _ := cmd.Flags().GetString("dburl"); dburl != "" {
		config.Set(config.KeyDatabaseUrl, dburl)
	}

	if ok, _ := afero.Exists(r.fs, r.cfgFile); ok {
		cfg, err := config.LoadProjectConfig(r.fs, r.cfgFile)
		if err != nil {
			return fmt.Errorf("failed to load project config: %w", err)
		}
		r.projectCfg = cfg
	}

	// Nonsensical operation
	if r.from != nil && r.projectCfg != nil {
		return errors.New("Using --from within a project is not needed")
	}

	if r.from == nil && r.projectCfg == nil {
		return errors.New("You either need to initialize the project, or provide a --from flag")
	}

	return nil
}

func (r *runner) RunE(cmd *cobra.Command, args []string) error {

	ctx := cmd.Context()

	log := logger.Ctx(ctx)

	if r.from != nil {
		r.projectCfg = config.DefaultProjectConfig()
	}

	migFS, err := migrationfs.FromConfig(r.fs, r.projectCfg)
	if err != nil {
		return err
	}

	hasMig, err := migrationfs.HasMigrations(migFS)
	if err != nil {
		return err
	}

	if !hasMig {
		if r.from != nil {
			log.Info().Msg("no migrations found locally, pulling migrations")

			err = pullmigrationsrunner.Run(ctx, migFS, r.from, r.fs)
			if err != nil {
				return err
			}

		} else {
			return errors.New("No migrations found")
		}
	} else {
		finfo, _ := migFS.Stat(".")
		log.Info().Str("path", finfo.Name()).Msg("migrations found locally")
	}

	// check again incase we pulled
	hasMig, err = migrationfs.HasMigrations(migFS)
	if err != nil {
		return err
	}
	if !hasMig {
		log.Error().Msg("No migrations found. Wrong path?")
		return nil
	}

	return gooserunner.Run(ctx, r.operation, migFS)
}
