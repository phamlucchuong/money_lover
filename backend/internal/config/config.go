package config

import (
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Port             string   `mapstructure:"PORT"`
	Environment      string   `mapstructure:"ENVIRONMENT"`
	PostgresHost     string   `mapstructure:"POSTGRES_HOST"`
	PostgresUser     string   `mapstructure:"POSTGRES_USER"`
	PostgresPassword string   `mapstructure:"POSTGRES_PASSWORD"`
	PostgresName     string   `mapstructure:"POSTGRES_DB"`
	PostgresPort     string   `mapstructure:"POSTGRESQL_PORT"`
	AllowedOrigins   []string `mapstructure:"ALLOWED_ORIGINS"`
}

func LoadConfig() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	viper.SetDefault("PORT", "8080")
	viper.SetDefault("ENVIRONMENT", "local")
	viper.SetDefault("POSTGRES_HOST", "localhost")
	viper.SetDefault("POSTGRES_USER", "user")
	viper.SetDefault("POSTGRES_PASSWORD", "password")
	viper.SetDefault("POSTGRES_DB", "dbname")
	viper.SetDefault("POSTGRESQL_PORT", "5432")
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
