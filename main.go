package main

//go:generate go run github.com/99designs/gqlgen generate

import (
	"github.com/kiwisheets/auth/directive"
	"github.com/kiwisheets/gql-server/gqlclient"
	"github.com/kiwisheets/server/graphqlapi"

	"github.com/kiwisheets/invoicing/config"
	"github.com/kiwisheets/invoicing/graphql/generated"
	"github.com/kiwisheets/invoicing/model"
	"github.com/kiwisheets/invoicing/mq"
	"github.com/kiwisheets/invoicing/resolver"
)

func main() {
	cfg := config.Server()

	app := graphqlapi.NewDefault()
	defer app.Shutdown()

	db := model.Init(&cfg.Database)
	defer db.Close()

	mq := mq.Init()
	defer mq.Close()

	c := generated.Config{
		Resolvers: &resolver.Resolver{
			DB:        db.DB,
			MQ:        mq,
			GqlClient: gqlclient.NewClient(&cfg.GqlClient),
		},
		Directives: generated.DirectiveRoot{
			IsAuthenticated:       directive.IsAuthenticated,
			IsSecureAuthenticated: directive.IsSecureAuthenticated,
			HasPerm:               directive.HasPerm,
			HasPerms:              directive.HasPerms,
		},
	}

	server := app.SetupServer(generated.NewExecutableSchema(c), &cfg.GraphQL, db.DB)
	server.Run(app.Logger)
}
