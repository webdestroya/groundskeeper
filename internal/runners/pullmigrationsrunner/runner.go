//go:build !groundskeeper.usegit

package pullmigrationsrunner

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/go-github/v84/github"
	"github.com/spf13/afero"
	"github.com/webdestroya/groundskeeper/internal/config"
	"github.com/webdestroya/groundskeeper/internal/globals"
	"github.com/webdestroya/groundskeeper/internal/runners"
	"github.com/webdestroya/groundskeeper/internal/utils/cmdutil"
	"github.com/webdestroya/x/logger"
	"go.yaml.in/yaml/v3"
	"golang.org/x/sync/errgroup"
)

type runner struct {
	from   *cmdutil.RepoPathRef
	destFS afero.Fs
}

func Run(ctx context.Context, migrationsFS afero.Fs, from *cmdutil.RepoPathRef, cfgFS afero.Fs) error {

	if err := New(migrationsFS, from).Run(ctx); err != nil {
		return err
	}

	if cfgFS != nil {
		if ok, _ := afero.Exists(cfgFS, config.DefaultProjectFilepath); !ok {
			logger.Ctx(ctx).Debug().Msg("writing project config")
			// Write the config after we pull
			cfg := config.DefaultProjectConfig()

			data, err := yaml.Marshal(cfg)
			if err != nil {
				return fmt.Errorf("failed to create config?: %w", err)
			}
			if err := afero.WriteFile(cfgFS, config.DefaultProjectFilepath, data, 0o644); err != nil {
				return fmt.Errorf("failed to write config file: %w", err)
			}
		}
	}

	return nil
}

func New(migrationsFS afero.Fs, from *cmdutil.RepoPathRef) runners.Runner {
	return &runner{
		destFS: migrationsFS,
		from:   from,
	}
}

func (r runner) Run(ctx context.Context) error {

	start := time.Now()

	defer func() {
		logger.Ctx(ctx).Info().Dur("duration", time.Since(start)).Msg("pull finished")
	}()

	owner := r.from.Owner()
	repo := r.from.RepositoryName()

	client, err := globals.GithubClient(ctx)
	if err != nil {
		return err
	}

	entries, err := listMigrationBlobs(ctx, client, owner, repo, r.from.Ref(), r.from.Path())
	if err != nil {
		return err
	}

	logger.Ctx(ctx).Info().Int("file_count", len(entries)).Msg("Fetching migration files from GitHub")

	if err := fetchBlobs(ctx, client, owner, repo, r.destFS, entries); err != nil {
		return err
	}

	return nil
}

// migrationEntry is a blob found under the migrations path in the repo tree.
type migrationEntry struct {
	name string // filename only (no directory prefix)
	sha  string // blob SHA for fetching content
}

// listMigrationBlobs resolves the ref, fetches the repo tree, and returns
// blob entries that live directly under migrationsPath.
func listMigrationBlobs(ctx context.Context, client *github.Client, owner, repo, ref, migrationsPath string) ([]migrationEntry, error) {
	log := logger.Ctx(ctx)
	// Resolve the ref (branch, tag, or SHA) to a commit so we can get
	// the root tree SHA.
	commit, _, err := client.Repositories.GetCommit(ctx, owner, repo, ref, nil)
	if err != nil {
		return nil, fmt.Errorf("resolve ref %q: %w", ref, err)
	}
	log.Info().Str("sha", commit.GetSHA()).Msg("resolving ref sha")

	rootTreeSHA := commit.GetCommit().GetTree().GetSHA()
	if rootTreeSHA == "" {
		return nil, fmt.Errorf("commit %q has no tree", ref)
	}
	log.Info().Str("treeSHA", rootTreeSHA).Msg("Resolved root tree SHA")

	// Try a recursive tree fetch first. This returns every entry in the
	// repo in a single API call and works for the vast majority of repos.
	tree, _, err := client.Git.GetTree(ctx, owner, repo, rootTreeSHA, true)
	if err != nil {
		return nil, fmt.Errorf("get tree: %w", err)
	}

	log.Info().Int("treesize", len(tree.Entries)).Msg("loading git tree")

	// If the tree was truncated (very large repo), fall back to walking
	// the path segments to find the subtree SHA, then list just that
	// subtree non-recursively.
	if tree.GetTruncated() {
		log.Debug().Msg("GetTree truncated, using fallback")
		return listMigrationBlobsFallback(ctx, client, owner, repo, rootTreeSHA, migrationsPath)
	}

	return filterTreeEntries(ctx, tree, migrationsPath), nil
}

// filterTreeEntries returns blob entries that are direct children of prefix.
func filterTreeEntries(ctx context.Context, tree *github.Tree, prefix string) []migrationEntry {
	log := logger.Ctx(ctx)

	// Ensure the prefix ends with "/" so "db/migrations" doesn't match
	// "db/migrations_old/foo.sql".
	dirPrefix := strings.TrimSuffix(prefix, "/") + "/"

	var entries []migrationEntry
	for _, e := range tree.Entries {
		if e.GetType() != "blob" {
			continue
		}
		entryPath := e.GetPath()
		if !strings.HasPrefix(entryPath, dirPrefix) {
			continue
		}
		// Only direct children (no further slashes after the prefix).
		relPath := strings.TrimPrefix(entryPath, dirPrefix)
		if strings.Contains(relPath, "/") {
			log.Warn().Str("path", relPath).Msg("ignoring directory in migrations path")
			continue
		}

		if !strings.EqualFold(filepath.Ext(relPath), `.sql`) {
			log.Warn().Str("path", relPath).Msg("ignoring non SQL file")
			continue
		}

		entries = append(entries, migrationEntry{
			name: relPath,
			sha:  e.GetSHA(),
		})
	}
	return entries
}

// listMigrationBlobsFallback handles the case where the recursive tree was
// truncated. It walks each path segment to resolve the subtree SHA for
// migrationsPath, then lists that subtree directly.
func listMigrationBlobsFallback(
	ctx context.Context,
	client *github.Client,
	owner, repo, rootTreeSHA, migrationsPath string,
) ([]migrationEntry, error) {
	treeSHA := rootTreeSHA

	// Walk each path segment: "db/migrations" -> ["db", "migrations"]
	segments := strings.SplitSeq(strings.Trim(migrationsPath, "/"), "/")
	for seg := range segments {
		tree, _, err := client.Git.GetTree(ctx, owner, repo, treeSHA, false)
		if err != nil {
			return nil, fmt.Errorf("get tree for segment %q: %w", seg, err)
		}

		found := false
		for _, e := range tree.Entries {
			if e.GetPath() == seg && e.GetType() == "tree" {
				treeSHA = e.GetSHA()
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("path segment %q not found in tree", seg)
		}
	}

	// Now treeSHA points at the migrations directory. List it directly.
	migTree, _, err := client.Git.GetTree(ctx, owner, repo, treeSHA, false)
	if err != nil {
		return nil, fmt.Errorf("get migrations tree: %w", err)
	}

	var entries []migrationEntry
	for _, e := range migTree.Entries {
		if e.GetType() != "blob" {
			continue
		}
		entries = append(entries, migrationEntry{
			name: e.GetPath(),
			sha:  e.GetSHA(),
		})
	}
	return entries, nil
}

// fetchBlobs downloads each blob in parallel and writes it into the
// "migrations" directory on the provided filesystem.
func fetchBlobs(
	ctx context.Context,
	client *github.Client,
	owner, repo string,
	memfs afero.Fs,
	entries []migrationEntry,
) error {

	maxConcurrent := config.MaxConcurrency()
	logger.Ctx(ctx).Info().Int("concurrency", maxConcurrent).Msg("git fetch concurrency limit")

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(maxConcurrent)

	var mu sync.Mutex

	for _, entry := range entries {
		g.Go(func() error {
			raw, _, err := client.Git.GetBlobRaw(ctx, owner, repo, entry.sha)
			if err != nil {
				return fmt.Errorf("fetch blob %s (%s): %w", entry.name, entry.sha, err)
			}

			filePath := entry.name

			mu.Lock()
			defer mu.Unlock()
			return afero.WriteFile(memfs, filePath, raw, 0o644)
		})
	}

	return g.Wait()
}
