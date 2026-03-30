package migratecmd

import (
	"github.com/spf13/cobra"
	"github.com/webdestroya/groundskeeper/internal/runners/gooserunner"
	"github.com/webdestroya/groundskeeper/internal/utils"
)

func NewStatus() *cobra.Command {
	r := &runner{
		operation: gooserunner.GooseStatus,
		fs:        utils.BaseFS(),
	}

	cmd := &cobra.Command{
		Use:     "status",
		Short:   "Show migration status",
		GroupID: "migration",
		PreRunE: r.PreRunE,
		RunE:    r.RunE,
		Args:    cobra.NoArgs,
		Aliases: []string{`db:migrate:status`},
	}

	r.setFlags(cmd)

	return cmd
}
