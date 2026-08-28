package config

import (
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Port                   string   `mapstructure:"PORT"`
	Environment            string   `mapstructure:"ENVIRONMENT"`
	PostgresHost           string   `mapstructure:"POSTGRES_HOST"`
	PostgresUser           string   `mapstructure:"POSTGRES_USER"`
	PostgresPassword       string   `mapstructure:"POSTGRES_PASSWORD"`
	PostgresName           string   `mapstructure:"POSTGRES_DB"`
	PostgresPort           string   `mapstructure:"POSTGRESQL_PORT"`
	RedisHost              string   `mapstructure:"REDIS_HOST"`
	RedisPort              string   `mapstructure:"REDIS_PORT"`
	AllowedOrigins         []string `mapstructure:"ALLOWED_ORIGINS"`
	JWTAccessSecret        string   `mapstructure:"JWT_ACCESS_SECRET"`
	JWTRefreshSecret       string   `mapstructure:"JWT_REFRESH_SECRET"`
	AccessTokenExpiration  int64    `mapstructure:"ACCESS_TOKEN_EXPIRATION"`
	RefreshTokenExpiration int64    `mapstructure:"REFRESH_TOKEN_EXPIRATION"`
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
	viper.SetDefault("REDIS_HOST", "localhost")
	viper.SetDefault("REDIS_PORT", "6379")
	viper.SetDefault("ALLOWED_ORIGINS", "http://localhost:3000")
	viper.SetDefault("JWT_ACCESS_SECRET", "your_jwt_access_secret")
	viper.SetDefault("JWT_REFRESH_SECRET", "your_jwt_refresh_secret")
	viper.SetDefault("ACCESS_TOKEN_EXPIRATION", 5)      // 5 minutes
	viper.SetDefault("REFRESH_TOKEN_EXPIRATION", 24*60) // 1 days

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
