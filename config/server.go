package config

import (
	"github.com/joho/godotenv"
	"github.com/kiwisheets/util"
	"github.com/maxtroughear/goenv"
)

func Server() *Config {
	godotenv.Load()

	return &Config{
		Version:     goenv.MustGet("APP_VERSION"),
		Environment: goenv.MustGet("ENVIRONMENT"),
		GraphQL: util.GqlConfig{
			ComplexityLimit:   200,
			Environment:       goenv.MustGet("ENVIRONMENT"),
			APIPath:           goenv.CanGet("API_PATH", "/"),
			PlaygroundPath:    goenv.CanGet("PLAYGROUND_PATH", "/graphql"),
			PlaygroundAPIPath: goenv.CanGet("PLAYGROUND_API_PATH", "/api/"),
			PlaygroundEnabled: goenv.MustGet("ENVIRONMENT") == "development",
			Port:              goenv.MustGet("PORT"),
			Redis:             util.RedisConfig{Address: goenv.MustGet("REDIS_ADDRESS")},
		},
		Database: util.DatabaseConfig{
			Host:           goenv.MustGet("POSTGRES_HOST"),
			Port:           goenv.MustGet("POSTGRES_PORT"),
			User:           goenv.MustGet("POSTGRES_USER"),
			Password:       goenv.MustGetSecretFromEnv("POSTGRES_PASSWORD"),
			Database:       goenv.MustGet("POSTGRES_DB"),
			MaxConnections: goenv.CanGetInt32("POSTGRES_MAX_CONNECTIONS", 20),
			SSLMode:        goenv.CanGet("POSTGRES_SSLMODE", "disable"),
			SSLCAPath:      goenv.CanGet("POSTGRES_SSL_CA_PATH", ""),
			Options:        goenv.CanGet("POSTGRES_OPTIONS", ""),
		},
		GqlClient: util.ClientConfig{
			BaseURL:        goenv.MustGet("GQL_SERVER_URL"),
			CfClientID:     goenv.CanGet("CF_CLIENT_ID", ""),
			CfClientSecret: goenv.CanGet("CF_CLIENT_SECRET", ""),
		},
	}
}
