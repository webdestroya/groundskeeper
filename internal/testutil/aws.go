package testutil

import (
	"testing"

	"github.com/webdestroya/awsmocker"
	"github.com/webdestroya/groundskeeper/internal/awsclients"
)

// StartMocker starts awsmocker with the given mocks and configures the global
// AWS config to route through the mock. The config is reset to nil when the
// test finishes.
func StartMocker(t *testing.T, mocks []*awsmocker.MockedEndpoint) {
	t.Helper()
	info := awsmocker.Start(t, awsmocker.WithMocks(mocks...))
	cfg := info.Config()
	awsclients.ForcedAwsConfig = &cfg
	t.Cleanup(func() {
		awsclients.ForcedAwsConfig = nil
	})
}
