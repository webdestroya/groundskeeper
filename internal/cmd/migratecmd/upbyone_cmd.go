package migratecmd

import (
	"github.com/spf13/cobra"
	"github.com/webdestroya/groundskeeper/internal/runners/gooserunner"
	"github.com/webdestroya/groundskeeper/internal/utils"
)

func NewUpByOne() *cobra.Command {
	r := &runner{
		operation: gooserunner.GooseUpByOne,
		fs:        utils.BaseFS(),
	}

	cmd := &cobra.Command{
		Use:     "up",
		Short:   "Apply the next single migration",
		GroupID: "migration",
		PreRunE: r.PreRunE,
		RunE:    r.RunE,
		Args:    cobra.NoArgs,
		Aliases: []string{`up-by-one`},
	}

	r.setFlags(cmd)

	return cmd
}
