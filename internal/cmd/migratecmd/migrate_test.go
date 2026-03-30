package migratecmd

import (
	"path"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"github.com/webdestroya/groundskeeper/internal/config"
	"github.com/webdestroya/groundskeeper/internal/testutil"
	"go.yaml.in/yaml/v3"
)

func TestLocalFlow(t *testing.T) {
	testutil.SetBaseFS(t, buildFakeFs(t))

	sqliteDBPath := `sqlite://` + t.TempDir() + `/testdb.db`

	args := []string{"--dburl", sqliteDBPath}

	// Status shows all pending
	out, err := testutil.RunCommand(t, NewStatus(), args)
	require.NoError(t, err)
	require.Contains(t, out.String(), `"state":"pending","name":"20260309040736_create_authors.sql"`)

	// Run single migration
	out, err = testutil.RunCommand(t, NewUpByOne(), args)
	require.NoError(t, err)
	require.Contains(t, out.String(), `20260309040736`)
	require.Contains(t, out.String(), `migration completed`)
	out, err = testutil.RunCommand(t, NewStatus(), args)
	require.NoError(t, err)
	require.Contains(t, out.String(), `"state":"applied","name":"20260309040736_create_authors.sql"`)

	// rollback previous
	_, err = testutil.RunCommand(t, NewRollback(), args)
	require.NoError(t, err)
	out, err = testutil.RunCommand(t, NewStatus(), args)
	require.NoError(t, err)
	require.Contains(t, out.String(), `"state":"pending","name":"20260309040736_create_authors.sql"`)

	// migrate all
	_, err = testutil.RunCommand(t, NewMigrate(), args)
	require.NoError(t, err)
	out, err = testutil.RunCommand(t, NewStatus(), args)
	require.NoError(t, err)
	require.Contains(t, out.String(), `"state":"applied","name":"20260309040736_create_authors.sql"`)
	require.Contains(t, out.String(), `"state":"applied","name":"20260309041055_create_books.sql"`)
	require.Contains(t, out.String(), `"state":"applied","name":"20260309041100_create_ratings.sql"`)
}

func buildFakeFs(t *testing.T) afero.Fs {
	t.Helper()

	fs := afero.NewMemMapFs()

	osfs := afero.NewOsFs()

	data, err := yaml.Marshal(&config.ProjectConfig{
		MigrationsPath: "migrations",
	})
	require.NoError(t, err)
	require.NoError(t, afero.WriteFile(fs, config.DefaultProjectFilepath, data, 0o644))

	require.NoError(t, fs.Mkdir("migrations", 0o755))
	entries, err := afero.ReadDir(osfs, "testdata")
	require.NoError(t, err)
	for _, entry := range entries {
		data, err := afero.ReadFile(osfs, path.Join("testdata", entry.Name()))
		require.NoError(t, err)

		require.NoError(t, afero.WriteFile(fs, path.Join("migrations", entry.Name()), data, 0o644))
	}

	return afero.NewReadOnlyFs(fs)
}
