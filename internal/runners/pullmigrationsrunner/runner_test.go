//go:build !groundskeeper.usegit

package pullmigrationsrunner

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/webdestroya/groundskeeper/internal/config"
	"github.com/webdestroya/groundskeeper/internal/testutil"
	"github.com/webdestroya/groundskeeper/internal/utils/cmdutil"
)

func parseRef(t *testing.T, value string) *cmdutil.RepoPathRef {
	t.Helper()
	ref, err := cmdutil.RepoPathRefParse(value)
	require.NoError(t, err)
	return ref
}

// testdata migration files used by the happy path test.
var migrationFiles = []struct {
	name string
	sha  string
}{
	{name: "20260309040736_create_authors.sql", sha: "sha_authors"},
	{name: "20260309041055_create_books.sql", sha: "sha_books"},
	{name: "20260309041100_create_ratings.sql", sha: "sha_ratings"},
}

// loadTestdata reads a migration file from the migratecmd testdata directory.
func loadTestdata(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(path.Join("..", "..", "cmd", "migratecmd", "testdata", name)) //nolint:forbidigo // it's a test
	require.NoError(t, err)
	return data
}

// commitAndTreeHandlers registers mock handlers for the standard
// GetCommit + GetTree flow that most tests share.
func commitAndTreeHandlers(t *testing.T, mux *http.ServeMux, treeEntries []map[string]any, truncated bool) {
	t.Helper()

	mux.HandleFunc("/repos/TestOrg/TestRepo/commits/main", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{
			"sha": "abc123",
			"commit": {
				"tree": { "sha": "tree456" }
			}
		}`)
	})

	treeResp := map[string]any{
		"sha":       "tree456",
		"tree":      treeEntries,
		"truncated": truncated,
	}

	mux.HandleFunc("/repos/TestOrg/TestRepo/git/trees/tree456", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(treeResp)
	})
}

func TestPull_HappyPath(t *testing.T) {
	mux := testutil.MockGithub(t)

	// Build tree entries for the 3 SQL files under db/migrations/
	treeEntries := make([]map[string]any, 0, len(migrationFiles))
	blobContent := make(map[string][]byte)
	for _, mf := range migrationFiles {
		treeEntries = append(treeEntries, map[string]any{
			"path": "db/migrations/" + mf.name,
			"type": "blob",
			"sha":  mf.sha,
			"mode": "100644",
		})
		blobContent[mf.sha] = loadTestdata(t, mf.name)
	}

	commitAndTreeHandlers(t, mux, treeEntries, false)

	// Blob handler: return raw content based on SHA
	mux.HandleFunc("/repos/TestOrg/TestRepo/git/blobs/", func(w http.ResponseWriter, r *http.Request) {
		sha := path.Base(r.URL.Path)
		data, ok := blobContent[sha]
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/vnd.github.v3.raw")
		w.Write(data)
	})

	destFS := afero.NewMemMapFs()
	cfgFS := afero.NewMemMapFs()
	from := parseRef(t, "TestOrg/TestRepo/db/migrations@main")

	err := Run(context.Background(), destFS, from, cfgFS)
	require.NoError(t, err)

	// All 3 SQL files should be on the destination FS
	for _, mf := range migrationFiles {
		data, err := afero.ReadFile(destFS, mf.name)
		require.NoError(t, err, "missing file: %s", mf.name)
		assert.Equal(t, string(blobContent[mf.sha]), string(data))
	}

	// Project config should have been written
	exists, err := afero.Exists(cfgFS, config.DefaultProjectFilepath)
	require.NoError(t, err)
	assert.True(t, exists, "project config should be written when cfgFS is provided")
}

func TestPull_OnlyRubyFiles(t *testing.T) {
	mux := testutil.MockGithub(t)

	// Tree contains 5 .rb files, no .sql files
	treeEntries := make([]map[string]any, 0, 5)
	for i := range 5 {
		treeEntries = append(treeEntries, map[string]any{
			"path": fmt.Sprintf("db/migrations/%03d_migration.rb", i+1),
			"type": "blob",
			"sha":  fmt.Sprintf("sha_rb_%d", i+1),
			"mode": "100644",
		})
	}

	commitAndTreeHandlers(t, mux, treeEntries, false)

	// No blob handler needed -- blobs should never be fetched

	destFS := afero.NewMemMapFs()
	from := parseRef(t, "TestOrg/TestRepo/db/migrations@main")

	err := Run(context.Background(), destFS, from, nil)
	require.NoError(t, err)

	// No files should be on the destination FS
	entries, err := afero.ReadDir(destFS, ".")
	require.NoError(t, err)
	assert.Empty(t, entries, "no SQL files should be pulled")
}

func TestPull_InvalidRef(t *testing.T) {
	mux := testutil.MockGithub(t)

	mux.HandleFunc("/repos/TestOrg/TestRepo/commits/nonexistent", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprint(w, `{"message":"No commit found for SHA: nonexistent"}`)
	})

	destFS := afero.NewMemMapFs()
	from := parseRef(t, "TestOrg/TestRepo/db/migrations@nonexistent")

	err := Run(context.Background(), destFS, from, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "resolve ref")
}

func TestPull_InvalidRepo(t *testing.T) {
	mux := testutil.MockGithub(t)

	mux.HandleFunc("/repos/BadOrg/BadRepo/commits/main", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"message":"Not Found"}`)
	})

	destFS := afero.NewMemMapFs()
	from := parseRef(t, "BadOrg/BadRepo/db/migrations@main")

	err := Run(context.Background(), destFS, from, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "resolve ref")
}

func TestPull_PathNotFound_TruncatedTree(t *testing.T) {
	mux := testutil.MockGithub(t)

	// Return a truncated tree so the fallback path kicks in.
	// The empty tree means the fallback won't find the "db" segment.
	commitAndTreeHandlers(t, mux, []map[string]any{}, true)

	destFS := afero.NewMemMapFs()
	from := parseRef(t, "TestOrg/TestRepo/db/migrations@main")

	err := Run(context.Background(), destFS, from, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found in tree")
}

func TestPull_GitHub503(t *testing.T) {
	mux := testutil.MockGithub(t)

	mux.HandleFunc("/repos/TestOrg/TestRepo/commits/main", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprint(w, `{"message":"Service Unavailable"}`)
	})

	destFS := afero.NewMemMapFs()
	from := parseRef(t, "TestOrg/TestRepo/db/migrations@main")

	err := Run(context.Background(), destFS, from, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "resolve ref")
}

func TestPull_GitHub403(t *testing.T) {
	mux := testutil.MockGithub(t)

	mux.HandleFunc("/repos/TestOrg/TestRepo/commits/main", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, `{"message":"API rate limit exceeded"}`)
	})

	destFS := afero.NewMemMapFs()
	from := parseRef(t, "TestOrg/TestRepo/db/migrations@main")

	err := Run(context.Background(), destFS, from, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "resolve ref")
}
