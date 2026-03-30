package configcmd

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/webdestroya/groundskeeper/internal/config"
	"github.com/webdestroya/groundskeeper/internal/testutil"
	"go.yaml.in/yaml/v3"
)

func writeProjectConfig(t *testing.T, fs afero.Fs, path string, migrationsPath string) {
	t.Helper()
	data, err := yaml.Marshal(&config.ProjectConfig{MigrationsPath: migrationsPath})
	require.NoError(t, err)
	require.NoError(t, afero.WriteFile(fs, path, data, 0o644))
}

func TestConfig_NoProjectConfig(t *testing.T) {
	memfs := afero.NewMemMapFs()
	testutil.SetBaseFS(t, memfs)

	cmd := New()
	out, err := testutil.RunCommand(t, cmd, []string{})
	require.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, `"migrations_path":"MISSING"`)
	assert.Contains(t, output, `"concurrency":10`)
	assert.Contains(t, output, `"Groundskeeper Configuration"`)
}

func TestConfig_WithProjectConfig(t *testing.T) {
	memfs := afero.NewMemMapFs()
	testutil.SetBaseFS(t, memfs)

	writeProjectConfig(t, memfs, config.DefaultProjectFilepath, "db/migrations")

	cmd := New()
	out, err := testutil.RunCommand(t, cmd, []string{})
	require.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, `"migrations_path":"db/migrations"`)
	assert.Contains(t, output, `"config_path":".groundskeeper.yml"`)
	assert.NotContains(t, output, `"migrations_path":"MISSING"`)
}

func TestConfig_CustomConfigPath(t *testing.T) {
	memfs := afero.NewMemMapFs()
	testutil.SetBaseFS(t, memfs)

	writeProjectConfig(t, memfs, "custom.yml", "custom/dir")

	cmd := New()
	out, err := testutil.RunCommand(t, cmd, []string{"--config", "custom.yml"})
	require.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, `"config_path":"custom.yml"`)
	assert.Contains(t, output, `"migrations_path":"custom/dir"`)
}

func TestConfig_ShowsDatabaseURL(t *testing.T) {
	memfs := afero.NewMemMapFs()
	testutil.SetBaseFS(t, memfs)

	config.Set(config.KeyDatabaseUrl, "postgres://localhost/testdb")
	t.Cleanup(func() { config.Set(config.KeyDatabaseUrl, "") })

	cmd := New()
	out, err := testutil.RunCommand(t, cmd, []string{})
	require.NoError(t, err)

	assert.Contains(t, out.String(), `"database_url":"postgres://localhost/testdb"`)
}

func TestConfig_ShowsDefaultConcurrency(t *testing.T) {
	memfs := afero.NewMemMapFs()
	testutil.SetBaseFS(t, memfs)

	cmd := New()
	out, err := testutil.RunCommand(t, cmd, []string{})
	require.NoError(t, err)

	assert.Contains(t, out.String(), `"concurrency":10`)
}

func TestConfig_RejectsPositionalArgs(t *testing.T) {
	memfs := afero.NewMemMapFs()
	testutil.SetBaseFS(t, memfs)

	cmd := New()
	_, err := testutil.RunCommand(t, cmd, []string{"extra"})
	require.Error(t, err)
}

func TestConfig_GithubCheck(t *testing.T) {
	memfs := afero.NewMemMapFs()
	testutil.SetBaseFS(t, memfs)

	mux := testutil.MockGithub(t)
	mux.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"id":1,"login":"testbot"}`)
	})

	cmd := New()
	out, err := testutil.RunCommand(t, cmd, []string{"--github"})
	require.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, `"github_user":"testbot"`)
	assert.Contains(t, output, `"checking current user"`)
	assert.Contains(t, output, `"resolving the github client"`)
}
