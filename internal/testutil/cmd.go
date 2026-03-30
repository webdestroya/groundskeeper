package testutil

import (
	"bytes"
	"testing"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/webdestroya/x/logger"
)

func RunCommand(t *testing.T, cmd *cobra.Command, args []string) (*bytes.Buffer, error) {
	t.Helper()

	buf := new(bytes.Buffer)

	if args == nil {
		args = make([]string, 0)
	}

	logger.SetLogger(zerolog.New(buf).Level(zerolog.TraceLevel))

	cmd.SetArgs(args)
	cmd.SetIn(nil)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()

	return buf, err
}
