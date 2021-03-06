package config

import "github.com/kiwisheets/util"

type Config struct {
	Version     string
	Environment string
	GraphQL     util.GqlConfig
	Hash        util.HashConfig
	Database    util.DatabaseConfig
	GqlClient   util.ClientConfig
}
