package testutil

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/webdestroya/groundskeeper/internal/utils"
)

func SetBaseFS(t *testing.T, fs afero.Fs) {
	t.Helper()
	utils.ForcedBaseFS = fs
	t.Cleanup(func() {
		utils.ForcedBaseFS = afero.NewMemMapFs()
	})
}
