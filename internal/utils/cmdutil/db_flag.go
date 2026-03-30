package cmdutil

import (
	"github.com/spf13/cobra"
)

func SetupDatabaseFlag(cmd *cobra.Command) {

	cmd.Flags().String("dburl", "", "Database URL or SSM parameter path")

	// cmd.MarkFlagRequired("dburl")
}
