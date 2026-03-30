package cmdutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRepoPathRefParse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantErr  bool
		wantRepo string
		wantRef  string
		wantPath string
	}{
		// valid cases
		{
			name:     "basic",
			input:    "Owner/Repo/path/to/migrations@v1.0.0",
			wantRepo: "Owner/Repo",
			wantRef:  "v1.0.0",
			wantPath: "path/to/migrations",
		},
		{
			name:     "branch ref",
			input:    "fakeorg/backend/db/migrate@main",
			wantRepo: "fakeorg/backend",
			wantRef:  "main",
			wantPath: "db/migrate",
		},
		{
			name:     "deep path",
			input:    "fakeorg/somerepo/internal/db/migrations@abc123",
			wantRepo: "fakeorg/somerepo",
			wantRef:  "abc123",
			wantPath: "internal/db/migrations",
		},
		{
			name:     "github.com prefix",
			input:    "github.com/fakeorg/backend/db/migrate@main",
			wantRepo: "fakeorg/backend",
			wantRef:  "main",
			wantPath: "db/migrate",
		},
		{
			name:     "https url prefix",
			input:    "https://github.com/fakeorg/backend/db/migrate@v2.0",
			wantRepo: "fakeorg/backend",
			wantRef:  "v2.0",
			wantPath: "db/migrate",
		},
		{
			name:     "leading slash",
			input:    "/Owner/Repo/path@tag",
			wantRepo: "Owner/Repo",
			wantRef:  "tag",
			wantPath: "path",
		},
		{
			name:     "single segment path",
			input:    "Owner/Repo/migrations@v1",
			wantRepo: "Owner/Repo",
			wantRef:  "v1",
			wantPath: "migrations",
		},
		{
			name:     "goofy1",
			input:    "github.com/Owner/Repo//things/yar/migrations@v1",
			wantRepo: "Owner/Repo",
			wantRef:  "v1",
			wantPath: "things/yar/migrations",
		},

		// invalid cases
		{
			name:    "no path segment",
			input:   "Owner/Repo@v1.0.0",
			wantErr: true,
		},
		{
			name:    "no at sign",
			input:   "just-a-string",
			wantErr: true,
		},
		{
			name:    "empty before at",
			input:   "@v1.0.0",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "only owner",
			input:   "Owner@tag",
			wantErr: true,
		},
		{
			name:    "empty ref",
			input:   "Owner/Repo/path@",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RepoPathRefParse(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantRepo, got.Repository())
				assert.Equal(t, tt.wantRef, got.Ref())
				assert.Equal(t, tt.wantPath, got.Path())
			}
		})
	}
}
