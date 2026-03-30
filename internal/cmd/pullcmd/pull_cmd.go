package pullcmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/webdestroya/groundskeeper/internal/config"
	"github.com/webdestroya/groundskeeper/internal/migrationfs"
	"github.com/webdestroya/groundskeeper/internal/runners/pullmigrationsrunner"
	"github.com/webdestroya/groundskeeper/internal/utils"
	"github.com/webdestroya/groundskeeper/internal/utils/cmdutil"
	"github.com/webdestroya/x/logger"
)

var invokePullRunner func(context.Context, afero.Fs, *cmdutil.RepoPathRef, afero.Fs) error = pullmigrationsrunner.Run

type runner struct {
	fs   afero.Fs
	from *cmdutil.RepoPathRef

	path string

	force bool
}

func New() *cobra.Command {
	r := &runner{
		fs: utils.BaseFS(),
	}

	cmd := &cobra.Command{
		Use:     "pull",
		Short:   "Pull migration files. Used within a remote console session",
		GroupID: "debug",
		PreRunE: r.PreRunE,
		RunE:    r.RunE,
		Args:    cobra.NoArgs,
	}

	cmdutil.RepoPathRefCmd(cmd, &r.from, true)
	cmd.Flags().BoolVar(&r.force, "force", false, "Force a pull even if we already have migrations locally. BE CAREFUL")

	return cmd
}

func (r *runner) PreRunE(cmd *cobra.Command, args []string) error {

	ctx := cmd.Context()
	log := logger.Ctx(ctx)

	if ok, _ := afero.Exists(r.fs, config.DefaultProjectFilepath); ok {
		log.Error().Msg("Project configuration exists")
		if !r.force {
			return errors.New("This command is not necessary within a project. You already have the migrations.")
		}

		log.Info().Msg("Pull was forced, so loading project config")
		cfg, err := config.LoadProjectConfig(r.fs, config.DefaultProjectFilepath)
		if err != nil {
			return fmt.Errorf("failed to load project config: %w", err)
		}
		r.path = cfg.MigrationsPath

		log.Warn().Str("dir", r.path).Msg("Forcing migration sync")
	} else {
		r.path = config.DefaultProjectConfig().MigrationsPath
	}

	return nil
}

func (r *runner) RunE(cmd *cobra.Command, args []string) error {

	if err := migrationfs.Create(r.fs, r.path); err != nil {
		return err
	}

	migFs := afero.NewBasePathFs(r.fs, r.path)

	var cfgFs afero.Fs

	if !r.force {
		cfgFs = r.fs
	}
	return invokePullRunner(cmd.Context(), migFs, r.from, cfgFs)
}
