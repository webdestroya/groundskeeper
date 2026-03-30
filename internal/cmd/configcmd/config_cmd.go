package configcmd

import (
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/webdestroya/groundskeeper/internal/config"
	"github.com/webdestroya/groundskeeper/internal/globals"
	"github.com/webdestroya/groundskeeper/internal/utils"
	"github.com/webdestroya/x/logger"
)

type runner struct {
	fs      afero.Fs
	cfgFile string

	checkGithub bool
}

func New() *cobra.Command {
	r := &runner{
		fs: utils.BaseFS(),
	}

	cmd := &cobra.Command{
		Use:     "config",
		GroupID: "development",
		Short:   "Shows groundskeeper configuration",
		RunE:    r.RunE,
		Args:    cobra.NoArgs,
	}

	config.AddProjectConfigFlag(cmd, &r.cfgFile)

	cmd.Flags().BoolVar(&r.checkGithub, "github", false, "Check GitHub token resolution")
	cmd.Flags().MarkHidden("github")

	return cmd
}

func (r *runner) RunE(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	log := logger.Ctx(ctx)

	cfg, err := config.LoadProjectConfig(r.fs, r.cfgFile)
	if err != nil {
		log.Warn().Err(err).Msg("attempted to load project config")
	}

	log.Info().Msg("Groundskeeper Configuration")
	log.Info().Str("config_path", r.cfgFile).Send()
	if cfg != nil {
		log.Info().Str("migrations_path", cfg.MigrationsPath).Send()
	} else {
		log.Info().Str("migrations_path", "MISSING").Send()
	}
	log.Info().Int("concurrency", config.MaxConcurrency()).Send()
	log.Info().Str("database_url", config.DatabaseURL()).Send()
	log.Info().Bool("task_mode", config.IsTaskMode()).Send()

	if r.checkGithub {

		log.Info().Str("github_token", config.GithubToken()).Send()

		client, err := globals.GithubClient(ctx)
		log.Err(err).Msg("resolving the github client")
		if err == nil {
			log.Info().Msg("checking current user")
			user, _, err := client.Users.Get(ctx, "")
			log.Err(err).Msg("getting user info")
			if err == nil {
				log.Info().Str("github_user", user.GetLogin()).Send()
			}

		}

	}

	return nil
}
