package cmdutil

import "github.com/spf13/cobra"

type runEFunc = func(*cobra.Command, []string) error

func ChainCallbacks(funcs ...runEFunc) runEFunc {
	return func(cmd *cobra.Command, args []string) error {
		for _, fn := range funcs {
			if fn == nil {
				continue
			}
			if err := fn(cmd, args); err != nil {
				return err
			}
		}
		return nil
	}
}
