package cmdutil

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

type RepoPathRef struct {
	owner      string
	repository string
	ref        string
	path       string
}

func (r *RepoPathRef) Validate() error {
	return nil
}

func (r RepoPathRef) Owner() string          { return r.owner }
func (r RepoPathRef) RepositoryName() string { return r.repository }
func (r RepoPathRef) Repository() string     { return r.owner + `/` + r.repository }
func (r RepoPathRef) Ref() string            { return r.ref }
func (r RepoPathRef) Path() string           { return r.path }

func (r RepoPathRef) String() string {
	return filepath.Join(r.owner, r.repository, r.path) + `@` + r.ref
}

func RepoPathRefFlag(cmd *cobra.Command, dest *string) {
	cmd.Flags().StringVar(dest, "from", "", "Migrations location `URI` (github.com/ORG/REPO/path/to/migrations@REF)")
}

// sets up the parsing from the command
// returns a PreRunE
func RepoPathRefCmd(root *cobra.Command, dest **RepoPathRef, required bool) {

	var fromVal string

	RepoPathRefFlag(root, &fromVal)

	preRunE := func(cmd *cobra.Command, _ []string) error {
		if fromVal != "" {
			rprVal, err := RepoPathRefParse(fromVal)
			if err != nil {
				return fmt.Errorf(`invalid migration location: %w`, err)
			}
			*dest = rprVal
			return nil
		}

		if required {
			return errors.New("Please provide a location for migrations using the --from flag")
		}

		return nil
	}
	root.PreRunE = ChainCallbacks(preRunE, root.PreRunE)

}

// Parses "repo-path-tag"
//
// Format:
//
//	[github.com/]Owner/Repo/Path/To/migrations@TAG
//
// Returns (repoName, ref, path)
func RepoPathRefParse(value string) (*RepoPathRef, error) {

	before, ref, ok := strings.Cut(value, `@`)
	if !ok || ref == "" {
		return nil, errors.New("missing ref")
	}

	// TODO: other ways to specify refs? any prefixes to strip

	if uri, err := url.Parse(before); err == nil {
		before = uri.Path
	}

	before = strings.TrimPrefix(before, `/`)
	before = strings.TrimPrefix(before, `github.com/`)

	parts := strings.SplitN(before, `/`, 3)
	if len(parts) != 3 {
		return nil, errors.New("invalid repo-path-ref")
	}

	owner := parts[0]
	repoName := parts[1]
	path := parts[2]

	return &RepoPathRef{
		owner:      owner,
		repository: repoName,
		ref:        ref,
		path:       strings.Trim(filepath.Clean(path), `/`),
	}, nil

}
