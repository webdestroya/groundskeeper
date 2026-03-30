package config

import (
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"go.yaml.in/yaml/v3"
)

const (
	DefaultProjectFilepath = `.groundskeeper.yml`
	ScratchMigrationPath   = `.gk_migrations`
)

type ProjectConfig struct {
	MigrationsPath string `yaml:"path,omitempty"`
}

var (
	_ yaml.Marshaler = (*ProjectConfig)(nil)
)

func (c ProjectConfig) MarshalYAML() (any, error) {
	return &struct {
		Version        int    `yaml:"version"`
		MigrationsPath string `yaml:"path,omitempty"`
	}{
		Version:        1,
		MigrationsPath: c.MigrationsPath,
	}, nil
}

func DefaultProjectConfig() *ProjectConfig {
	return &ProjectConfig{
		MigrationsPath: ScratchMigrationPath,
	}
}

func AddProjectConfigFlag(cmd *cobra.Command, dest *string) {
	cmd.Flags().StringVarP(dest, "config", "c", DefaultProjectFilepath, "Use custom project configuration `file`")
	_ = cmd.MarkFlagFilename("config", "yml", "yaml")
}

func LoadProjectConfig(fs afero.Fs, path string) (*ProjectConfig, error) {
	data, err := afero.ReadFile(fs, path)
	if err != nil {
		return nil, err
	}

	cfg := &ProjectConfig{}

	err = yaml.Unmarshal(data, cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
