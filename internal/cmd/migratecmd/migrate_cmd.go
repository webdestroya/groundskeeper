package migratecmd

import (
	"github.com/spf13/cobra"
	"github.com/webdestroya/groundskeeper/internal/runners/gooserunner"
	"github.com/webdestroya/groundskeeper/internal/utils"
)

func NewMigrate() *cobra.Command {
	r := &runner{
		operation: gooserunner.GooseUpAll,
		fs:        utils.BaseFS(),
	}

	cmd := &cobra.Command{
		Use:     "migrate",
		Short:   "Apply all pending migrations",
		GroupID: "migration",
		PreRunE: r.PreRunE,
		RunE:    r.RunE,
		Args:    cobra.NoArgs,
		Aliases: []string{`db:migrate`},
	}

	r.setFlags(cmd)

	return cmd
}
