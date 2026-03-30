package main

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/webdestroya/groundskeeper/internal/cmd"
)

func main() {
	code, _ := cmd.Execute(func(c *cobra.Command) {
		c.SetErr(os.Stderr)
		c.SetOut(os.Stdout)
		c.SetIn(os.Stdin)
	})
	os.Exit(code)
}
