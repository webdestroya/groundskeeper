package migrationfs

import (
	"fmt"

	"github.com/spf13/afero"
	"github.com/webdestroya/groundskeeper/internal/config"
)

func Create(fs afero.Fs, path string) error {
	if ok, _ := afero.DirExists(fs, path); ok {
		return nil
	}

	if err := fs.MkdirAll(path, 0o755); err != nil {
		return fmt.Errorf("unable to create migrations directory: %w", err)
	}

	return nil
}

func FromConfig(base afero.Fs, cfg *config.ProjectConfig) (afero.Fs, error) {

	path := cfg.MigrationsPath

	if err := Create(base, path); err != nil {
		return nil, err
	}

	return afero.NewBasePathFs(base, path), nil
}

func HasMigrations(fs afero.Fs) (bool, error) {

	entries, err := afero.Glob(fs, `*.sql`)
	if err != nil {
		return false, err
	}

	return len(entries) > 0, nil
}
