package initcmd

import (
	"fmt"
	"path"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/webdestroya/groundskeeper/internal/config"
	"github.com/webdestroya/groundskeeper/internal/migrationfs"
	"github.com/webdestroya/groundskeeper/internal/utils"
	"github.com/webdestroya/x/logger"
	"go.yaml.in/yaml/v3"
)

type runner struct {
	fs   afero.Fs
	path string
}

func New() *cobra.Command {
	r := &runner{
		fs: utils.BaseFS(),
	}

	cmd := &cobra.Command{
		Use:     "init",
		GroupID: "development",
		Short:   "Initialize a groundskeeper project",
		PreRunE: r.PreRunE,
		RunE:    r.RunE,
		Args:    cobra.NoArgs,
	}

	cmd.Flags().StringVar(&r.path, "path", path.Join(`db`, `migrations`), "Specify the `dir`ectory where migrations will reside.")
	_ = cmd.MarkFlagDirname("path")

	return cmd
}

func (r *runner) PreRunE(cmd *cobra.Command, args []string) error {

	if ok, _ := afero.Exists(r.fs, config.DefaultProjectFilepath); ok {
		return fmt.Errorf("project already initialized")
	}

	return nil
}

func (r *runner) RunE(cmd *cobra.Command, args []string) error {

	log := logger.Ctx(cmd.Context())

	cfg := &config.ProjectConfig{
		MigrationsPath: r.path,
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to create config?: %w", err)
	}

	log.Info().Str("path", r.path).Msg("Creating migrations folder")
	if err := migrationfs.Create(r.fs, r.path); err != nil {
		return fmt.Errorf("failed to create initial directory: %w", err)
	}

	log.Info().Str("path", config.DefaultProjectFilepath).Msg("Creating project file")
	if err := afero.WriteFile(r.fs, config.DefaultProjectFilepath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
