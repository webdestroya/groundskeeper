package config

import (
	"github.com/spf13/viper"
)

func SetDefaults(v *viper.Viper) {
	v.SetDefault(KeyDatabaseUrl, "")
	v.SetDefault(KeyGithubToken, "")
	v.SetDefault(KeyConcurrency, 10)
}
