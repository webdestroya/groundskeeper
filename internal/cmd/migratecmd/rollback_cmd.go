package migratecmd

import (
	"github.com/spf13/cobra"
	"github.com/webdestroya/groundskeeper/internal/runners/gooserunner"
	"github.com/webdestroya/groundskeeper/internal/utils"
)

func NewRollback() *cobra.Command {
	r := &runner{
		operation: gooserunner.GooseDown,
		fs:        utils.BaseFS(),
	}

	cmd := &cobra.Command{
		Use:     "rollback",
		Short:   "Rollback the previous migration",
		GroupID: "migration",
		PreRunE: r.PreRunE,
		RunE:    r.RunE,
		Args:    cobra.NoArgs,
		Aliases: []string{`db:rollback`, `down`},
	}

	r.setFlags(cmd)

	return cmd
}
