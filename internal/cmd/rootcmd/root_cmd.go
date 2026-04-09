package rootcmd

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
	"github.com/webdestroya/groundskeeper/internal/cmd/configcmd"
	"github.com/webdestroya/groundskeeper/internal/cmd/createcmd"
	"github.com/webdestroya/groundskeeper/internal/cmd/initcmd"
	"github.com/webdestroya/groundskeeper/internal/cmd/migratecmd"
	"github.com/webdestroya/groundskeeper/internal/cmd/pullcmd"
	"github.com/webdestroya/groundskeeper/internal/utils/buildinfo"
)

func New() *cobra.Command {

	cmd := &cobra.Command{
		Use:     "groundskeeper",
		Version: buildinfo.Version,
		Short:   `Manages migrations for your database.`,
		CompletionOptions: cobra.CompletionOptions{
			HiddenDefaultCmd: true,
		},

		Long: heredoc.Doc(`
			DEPLOY TASK USAGE:
			  $ groundskeeper migrate --from github.com/OWNER/REPO/db/migrations@v1.2.3 --dburl ssm:/ssm/path/DATABASE_URL

			LOCAL DEVELOPMENT USAGE:

			Initialize a new project and create a db/migrations folder
			  $ groundskeeper init
			
			Create a new migration file
			  $ groundskeeper new create_user_table  

			Check status of migrations
			  $ groundskeeper status

			Run migrations locally
			  $ groundskeeper migrate
		`),
	}

	cmd.AddGroup(&cobra.Group{
		ID:    "development",
		Title: "Development Commands:",
	})

	cmd.AddGroup(&cobra.Group{
		ID:    "debugging",
		Title: "Debugging Commands:",
	})

	cmd.AddGroup(&cobra.Group{
		ID:    "migration",
		Title: "Migration Commands:",
	})

	cmd.AddCommand(createcmd.New())
	cmd.AddCommand(initcmd.New())
	cmd.AddCommand(configcmd.New())

	cmd.AddCommand(migratecmd.NewMigrate())
	cmd.AddCommand(migratecmd.NewRollback())
	cmd.AddCommand(migratecmd.NewStatus())
	cmd.AddCommand(migratecmd.NewUpByOne())

	cmd.AddCommand(pullcmd.New())

	return cmd
}
