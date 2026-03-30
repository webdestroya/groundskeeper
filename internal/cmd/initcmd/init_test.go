package initcmd

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/webdestroya/groundskeeper/internal/config"
	"github.com/webdestroya/groundskeeper/internal/testutil"
	"go.yaml.in/yaml/v3"
)

func readProjectConfig(t *testing.T, fs afero.Fs) *config.ProjectConfig {
	t.Helper()
	data, err := afero.ReadFile(fs, config.DefaultProjectFilepath)
	require.NoError(t, err)
	cfg := &config.ProjectConfig{}
	require.NoError(t, yaml.Unmarshal(data, cfg))
	return cfg
}

func TestInit_NoFlags(t *testing.T) {
	memfs := afero.NewMemMapFs()
	testutil.SetBaseFS(t, memfs)

	cmd := New()
	_, err := testutil.RunCommand(t, cmd, nil)
	require.NoError(t, err)

	exists, err := afero.Exists(memfs, config.DefaultProjectFilepath)
	require.NoError(t, err)
	assert.True(t, exists, ".groundskeeper.yml should exist")

	cfg := readProjectConfig(t, memfs)
	assert.Equal(t, "db/migrations", cfg.MigrationsPath)

	dirExists, err := afero.DirExists(memfs, "db/migrations")
	require.NoError(t, err)
	assert.True(t, dirExists, "db/migrations directory should exist")
}

func TestInit_WithPathFlag(t *testing.T) {
	memfs := afero.NewMemMapFs()
	testutil.SetBaseFS(t, memfs)

	cmd := New()
	_, err := testutil.RunCommand(t, cmd, []string{"--path", "custom/migrations"})
	require.NoError(t, err)

	exists, err := afero.Exists(memfs, config.DefaultProjectFilepath)
	require.NoError(t, err)
	assert.True(t, exists, ".groundskeeper.yml should exist")

	cfg := readProjectConfig(t, memfs)
	assert.Equal(t, "custom/migrations", cfg.MigrationsPath)

	dirExists, err := afero.DirExists(memfs, "custom/migrations")
	require.NoError(t, err)
	assert.True(t, dirExists, "custom/migrations directory should exist")
}

func TestInit_AlreadyInitialized(t *testing.T) {
	memfs := afero.NewMemMapFs()
	testutil.SetBaseFS(t, memfs)

	originalContent := []byte("existing config")
	require.NoError(t, afero.WriteFile(memfs, config.DefaultProjectFilepath, originalContent, 0o644))

	cmd := New()
	_, err := testutil.RunCommand(t, cmd, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "project already initialized")

	data, err := afero.ReadFile(memfs, config.DefaultProjectFilepath)
	require.NoError(t, err)
	assert.Equal(t, originalContent, data, ".groundskeeper.yml should not be overwritten")
}
