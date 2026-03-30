package createcmd

import (
	"path/filepath"
	"regexp"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/webdestroya/groundskeeper/internal/config"
	"github.com/webdestroya/groundskeeper/internal/testutil"
	"go.yaml.in/yaml/v3"
)

func writeProjectConfig(t *testing.T, fs afero.Fs, migrationsPath string) {
	t.Helper()
	cfg := &config.ProjectConfig{
		MigrationsPath: migrationsPath,
	}
	data, err := yaml.Marshal(cfg)
	require.NoError(t, err)
	require.NoError(t, afero.WriteFile(fs, config.DefaultProjectFilepath, data, 0o644))
}

func findMigrationFile(t *testing.T, fs afero.Fs, dir string) string {
	t.Helper()
	matches, err := afero.Glob(fs, filepath.Join(dir, "*.sql"))
	require.NoError(t, err)
	require.Len(t, matches, 1, "expected exactly one .sql file in %s", dir)
	return matches[0]
}

var timestampPattern = regexp.MustCompile(`^\d{14}_`)

func TestCreate_Basic(t *testing.T) {
	memfs := afero.NewMemMapFs()
	testutil.SetBaseFS(t, memfs)

	writeProjectConfig(t, memfs, "db/migrations")
	require.NoError(t, memfs.MkdirAll("db/migrations", 0o755))

	cmd := New()
	_, err := testutil.RunCommand(t, cmd, []string{"create_users_table"})
	require.NoError(t, err)

	f := findMigrationFile(t, memfs, "db/migrations")
	base := filepath.Base(f)

	assert.True(t, timestampPattern.MatchString(base), "filename should start with a 14-digit timestamp")
	assert.Contains(t, base, "_create_users_table.sql")

	content, err := afero.ReadFile(memfs, f)
	require.NoError(t, err)
	assert.Equal(t, migrationContent, string(content))
}

func TestCreate_WithPathOverride(t *testing.T) {
	memfs := afero.NewMemMapFs()
	testutil.SetBaseFS(t, memfs)

	cmd := New()
	_, err := testutil.RunCommand(t, cmd, []string{"--path", "custom/dir", "add_index"})
	require.NoError(t, err)

	dirExists, err := afero.DirExists(memfs, "custom/dir")
	require.NoError(t, err)
	assert.True(t, dirExists, "custom/dir should be created")

	f := findMigrationFile(t, memfs, "custom/dir")
	assert.Contains(t, filepath.Base(f), "_add_index.sql")

	content, err := afero.ReadFile(memfs, f)
	require.NoError(t, err)
	assert.Equal(t, migrationContent, string(content))
}

func TestCreate_MultiWordName(t *testing.T) {
	memfs := afero.NewMemMapFs()
	testutil.SetBaseFS(t, memfs)

	writeProjectConfig(t, memfs, "db/migrations")
	require.NoError(t, memfs.MkdirAll("db/migrations", 0o755))

	cmd := New()
	_, err := testutil.RunCommand(t, cmd, []string{"create users table"})
	require.NoError(t, err)

	f := findMigrationFile(t, memfs, "db/migrations")
	assert.Contains(t, filepath.Base(f), "_create_users_table.sql")
}

func TestCreate_NoProjectConfig(t *testing.T) {
	memfs := afero.NewMemMapFs()
	testutil.SetBaseFS(t, memfs)

	cmd := New()
	_, err := testutil.RunCommand(t, cmd, []string{"some_migration"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Project has not been initialized")
}

func TestCreate_NoMigrationName(t *testing.T) {
	memfs := afero.NewMemMapFs()
	testutil.SetBaseFS(t, memfs)

	writeProjectConfig(t, memfs, "db/migrations")

	cmd := New()
	_, err := testutil.RunCommand(t, cmd, nil)
	require.Error(t, err)
}

func TestCreate_NameIsUppercased(t *testing.T) {
	memfs := afero.NewMemMapFs()
	testutil.SetBaseFS(t, memfs)

	writeProjectConfig(t, memfs, "db/migrations")
	require.NoError(t, memfs.MkdirAll("db/migrations", 0o755))

	cmd := New()
	_, err := testutil.RunCommand(t, cmd, []string{"CreateUsersTable"})
	require.NoError(t, err)

	f := findMigrationFile(t, memfs, "db/migrations")
	base := filepath.Base(f)

	// SnakeCase only inserts underscores at delimiters (spaces, hyphens),
	// not at camelCase transitions. The name is lowercased first, so
	// "CreateUsersTable" -> "createuserstable" -> "createuserstable"
	assert.Contains(t, base, "_createuserstable.sql")
	assert.NotContains(t, base, "C")
}
