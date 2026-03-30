package createcmd

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/webdestroya/groundskeeper/internal/config"
	"github.com/webdestroya/groundskeeper/internal/migrationfs"
	"github.com/webdestroya/groundskeeper/internal/utils"
	"github.com/webdestroya/x/util"
)

const (
	timestampFormat  = "20060102150405"
	migrationContent = `-- +goose Up
SELECT 'up SQL query';

-- +goose Down
SELECT 'down SQL query';
`
)

type runner struct {
	fs      afero.Fs
	name    string
	cfgFile string

	migrationsPath string
}

func New() *cobra.Command {
	r := &runner{
		fs: utils.BaseFS(),
	}

	cmd := &cobra.Command{
		Use:     "create MIGRATION_NAME",
		Short:   "Create a new migration file",
		GroupID: "development",
		PreRunE: r.PreRunE,
		RunE:    r.RunE,
		Args:    cobra.MinimumNArgs(1),
		Aliases: []string{"new"},
		Example: heredoc.Doc(`
		$ groundskeeper create create_users_table
		`),
	}

	cmd.Flags().StringVar(&r.migrationsPath, "path", "", "Override migrations `dir`ectory. (Rarely needed)")

	config.AddProjectConfigFlag(cmd, &r.cfgFile)

	return cmd
}

func (r *runner) PreRunE(cmd *cobra.Command, args []string) error {

	if r.migrationsPath == "" {

		cfg, err := config.LoadProjectConfig(r.fs, r.cfgFile)
		if err != nil {
			return errors.New("Project has not been initialized. Please run `groundskeeper init` first.")
		}

		r.migrationsPath = cfg.MigrationsPath
	}

	if len(args) > 0 {
		r.name = strings.Join(args, " ")
	}

	r.name = strings.TrimSpace(r.name)

	if r.name == "" {
		return errors.New("You need to provide a migration name")
	}

	r.name = strings.ToLower(r.name)

	return nil
}

func (r *runner) RunE(cmd *cobra.Command, args []string) error {

	if err := migrationfs.Create(r.fs, r.migrationsPath); err != nil {
		return err
	}

	migrationFs := afero.NewBasePathFs(r.fs, r.migrationsPath).(*afero.BasePathFs)

	version := time.Now().UTC().Format(timestampFormat)
	filename := fmt.Sprintf("%v_%v.sql", version, util.SnakeCase(r.name))

	if ok, _ := afero.Exists(migrationFs, filename); ok {
		return fmt.Errorf("migration file already exists?")
	}

	if err := afero.WriteFile(migrationFs, filename, []byte(migrationContent), 0o644); err != nil {
		return fmt.Errorf("failed to write migration file: %w", err)
	}

	realpath, _ := migrationFs.RealPath(filename)
	cmd.Println("Write migration file:", realpath)

	return nil
}
