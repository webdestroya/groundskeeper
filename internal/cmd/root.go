package cmd

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/webdestroya/groundskeeper/internal/cmd/rootcmd"
	"github.com/webdestroya/x/logger"
)

func Execute(extras ...func(*cobra.Command)) (int, error) {
	cobra.EnableTraverseRunHooks = true
	rootCmd := rootcmd.New()
	for _, extraFn := range extras {
		extraFn(rootCmd)
	}

	ctx := context.Background()

	logger.SetConsoleMode()
	ctx = logger.WithContext(ctx)

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	outCmd, err := rootCmd.ExecuteContextC(ctx)
	if err != nil {
		stop()

		if exitErr, ok := errors.AsType[*exec.ExitError](err); ok {
			return exitErr.ExitCode(), err
		}

		if errors.Is(err, context.Canceled) {
			logger.Ctx(ctx).Info().Msg("Interrupt received, exiting")
			return 0, nil
		}

		outCmd.PrintErrln(err)

		return 1, err
	}

	return 0, nil
}
