package pullcmd

import (
	"context"
	"errors"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/webdestroya/groundskeeper/internal/config"
	"github.com/webdestroya/groundskeeper/internal/testutil"
	"github.com/webdestroya/groundskeeper/internal/utils/cmdutil"
	"go.yaml.in/yaml/v3"
)

type pullRunnerCall struct {
	from  *cmdutil.RepoPathRef
	cfgFs afero.Fs
}

func stubPullRunner(t *testing.T, retErr error) *pullRunnerCall {
	t.Helper()

	original := invokePullRunner
	t.Cleanup(func() { invokePullRunner = original })

	captured := &pullRunnerCall{}
	invokePullRunner = func(_ context.Context, _ afero.Fs, from *cmdutil.RepoPathRef, cfgFs afero.Fs) error {
		captured.from = from
		captured.cfgFs = cfgFs
		return retErr
	}

	return captured
}

func writeProjectConfig(t *testing.T, fs afero.Fs, migrationsPath string) {
	t.Helper()
	data, err := yaml.Marshal(&config.ProjectConfig{MigrationsPath: migrationsPath})
	require.NoError(t, err)
	require.NoError(t, afero.WriteFile(fs, config.DefaultProjectFilepath, data, 0o644))
}

func TestPull_MissingFromFlag(t *testing.T) {
	memfs := afero.NewMemMapFs()
	testutil.SetBaseFS(t, memfs)

	cmd := New()
	_, err := testutil.RunCommand(t, cmd, []string{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "provide a location")
}

func TestPull_InvalidFromFlag(t *testing.T) {
	memfs := afero.NewMemMapFs()
	testutil.SetBaseFS(t, memfs)

	cmd := New()
	_, err := testutil.RunCommand(t, cmd, []string{"--from", "bad-value"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid migration location")
}

func TestPull_Basic(t *testing.T) {
	memfs := afero.NewMemMapFs()
	testutil.SetBaseFS(t, memfs)

	captured := stubPullRunner(t, nil)

	cmd := New()
	_, err := testutil.RunCommand(t, cmd, []string{"--from", "Org/Repo/db/migrations@main"})
	require.NoError(t, err)

	// Runner was called with correct RepoPathRef
	require.NotNil(t, captured.from)
	assert.Equal(t, "Org", captured.from.Owner())
	assert.Equal(t, "Repo", captured.from.RepositoryName())
	assert.Equal(t, "db/migrations", captured.from.Path())
	assert.Equal(t, "main", captured.from.Ref())

	// cfgFs is non-nil in normal (non-force) mode so project config gets written
	assert.NotNil(t, captured.cfgFs)

	// Migrations directory was created
	dirExists, err := afero.DirExists(memfs, config.ScratchMigrationPath)
	require.NoError(t, err)
	assert.True(t, dirExists)
}

func TestPull_WithGithubPrefix(t *testing.T) {
	memfs := afero.NewMemMapFs()
	testutil.SetBaseFS(t, memfs)

	captured := stubPullRunner(t, nil)

	cmd := New()
	_, err := testutil.RunCommand(t, cmd, []string{"--from", "github.com/someorg/somerepo/db/internal/migrations@main"})
	require.NoError(t, err)

	require.NotNil(t, captured.from)
	assert.Equal(t, "someorg", captured.from.Owner())
	assert.Equal(t, "somerepo", captured.from.RepositoryName())
	assert.Equal(t, "db/internal/migrations", captured.from.Path())
	assert.Equal(t, "main", captured.from.Ref())
}

func TestPull_ProjectConfigExists_NoForce(t *testing.T) {
	memfs := afero.NewMemMapFs()
	testutil.SetBaseFS(t, memfs)

	writeProjectConfig(t, memfs, "db/migrations")

	cmd := New()
	_, err := testutil.RunCommand(t, cmd, []string{"--from", "Org/Repo/db/migrations@main"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not necessary within a project")
}

func TestPull_ProjectConfigExists_WithForce(t *testing.T) {
	memfs := afero.NewMemMapFs()
	testutil.SetBaseFS(t, memfs)

	writeProjectConfig(t, memfs, "custom/migrations")
	captured := stubPullRunner(t, nil)

	cmd := New()
	_, err := testutil.RunCommand(t, cmd, []string{"--from", "Org/Repo/db/migrations@v1.2.3", "--force"})
	require.NoError(t, err)

	// Runner was called
	require.NotNil(t, captured.from)
	assert.Equal(t, "v1.2.3", captured.from.Ref())

	// cfgFs is nil in force mode (no config write)
	assert.Nil(t, captured.cfgFs)

	// Migrations directory uses the path from the existing project config
	dirExists, err := afero.DirExists(memfs, "custom/migrations")
	require.NoError(t, err)
	assert.True(t, dirExists)
}

func TestPull_RunnerError(t *testing.T) {
	memfs := afero.NewMemMapFs()
	testutil.SetBaseFS(t, memfs)

	stubPullRunner(t, errors.New("github API exploded"))

	cmd := New()
	_, err := testutil.RunCommand(t, cmd, []string{"--from", "Org/Repo/db/migrations@main"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "github API exploded")
}

func TestPull_RejectsPositionalArgs(t *testing.T) {
	memfs := afero.NewMemMapFs()
	testutil.SetBaseFS(t, memfs)

	cmd := New()
	_, err := testutil.RunCommand(t, cmd, []string{"--from", "Org/Repo/db/migrations@main", "extra-arg"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown command")
}
