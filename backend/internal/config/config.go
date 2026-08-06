package config

import (
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Port           string   `mapstructure:"PORT"`
	Environment    string   `mapstructure:"ENVIRONMENT"`
	DatabaseURL    string   `mapstructure:"DATABASE_URL"`
	AllowedOrigins []string `mapstructure:"ALLOWED_ORIGINS"`
}

func LoadConfig() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	viper.SetDefault("PORT", "8080")
	viper.SetDefault("ENVIRONMENT", "local")
	viper.SetDefault("DATABASE_URL", "postgresql://user:password@localhost/dbname")
	viper.SetDefault("ALLOWED_ORIGINS", "http://localhost:3000")

	_ = viper.ReadInConfig()

	cfg := &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, err
	}

	if len(cfg.AllowedOrigins) == 1 {
		cfg.AllowedOrigins = strings.Split(cfg.AllowedOrigins[0], ",")
	}

	return cfg, nil
}
