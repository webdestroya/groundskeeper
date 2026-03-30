package config

import (
	"github.com/spf13/viper"
)

var ConfigBackend *viper.Viper = GetViper()

func DatabaseURL() string {
	return ConfigBackend.GetString(KeyDatabaseUrl)
}

func GithubToken() string {
	return ConfigBackend.GetString(KeyGithubToken)
}

func MaxConcurrency() int {
	return ConfigBackend.GetInt(KeyConcurrency)
}

func IsTaskMode() bool {
	return ConfigBackend.IsSet(KeyTaskMode)
}

func GetViper() *viper.Viper {
	v := viper.NewWithOptions()
	v.SetEnvPrefix(ServiceName)
	v.AutomaticEnv()
	SetDefaults(v)

	v.BindEnv(KeyDatabaseUrl, "GROUNDSKEEPER_DATABASE_URL", "GK_DATABASE_URL", "DATABASE_URL") //nolint:errcheck // only errors with zero args
	v.BindEnv(KeyGithubToken, "GROUNDSKEEPER_GITHUB_TOKEN", "GITHUB_TOKEN")                    //nolint:errcheck // only errors with zero args

	return v
}

func Set(key string, value any) {
	ConfigBackend.Set(key, value)
}
