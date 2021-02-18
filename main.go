package main

//go:generate go run github.com/99designs/gqlgen generate

import (
	"fmt"
	"net/http"
	"time"

	"github.com/kiwisheets/auth/directive"
	"github.com/kiwisheets/gql-server/client"
	"github.com/kiwisheets/orm"
	"github.com/kiwisheets/server"
	"github.com/kiwisheets/util"
	"github.com/maxtroughear/logrusextension"
	"github.com/maxtroughear/nrextension"

	"github.com/kiwisheets/invoicing/config"
	"github.com/kiwisheets/invoicing/graphql/generated"
	"github.com/kiwisheets/invoicing/graphqlapi"
	"github.com/kiwisheets/invoicing/model"
	"github.com/kiwisheets/invoicing/resolver"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/aymerick/raymond"
	"github.com/sethgrid/pester"
	"github.com/sirupsen/logrus"
)

func main() {
	cfg := config.Server()

	app := graphqlapi.NewDefault()
	if app.NrApp != nil {
		defer app.NrApp.Shutdown(30 * time.Second)
	}

	raymond.RegisterHelper("total", model.InvoiceTotalHelper)
	raymond.RegisterHelper("itemCost", model.InvoiceItemCostHelper)
	raymond.RegisterHelper("itemTotal", model.InvoiceItemTotalHelper)

	db := orm.Init(&cfg.Database)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	model.Migrate(db)

	// messaging

	mq := util.NewMQ()
	defer mq.Close()
	go handleMQErrors(mq.Error())

	createProducer, err := mq.SyncProducer("invoice_create")
	if err != nil {
		panic(fmt.Errorf("failed to create producer: invoice_create: %s", err))
	}

	renderProducer, err := mq.SyncProducer("invoice_render")
	if err != nil {
		panic(fmt.Errorf("failed to create producer: invoice_render: %s", err))
	}

	c := generated.Config{
		Resolvers: &resolver.Resolver{
			DB:             db,
			CreateProducer: createProducer,
			RenderProducer: renderProducer,
			GqlServerClient: client.NewClient(pester.New(), cfg.GqlServerURL, func(req *http.Request) {
				if cfg.CfClientID != "" && cfg.CfClientSecret != "" {
					req.Header.Set("CF-Access-Client-Id", cfg.CfClientID)
					req.Header.Set("CF-Access-Client-Secret", cfg.CfClientSecret)
				}
			}),
		},
		Directives: generated.DirectiveRoot{
			IsAuthenticated:       directive.IsAuthenticated,
			IsSecureAuthenticated: directive.IsSecureAuthenticated,
			HasPerm:               directive.HasPerm,
			HasPerms:              directive.HasPerms,
		},
	}

	gqlHandler := handler.New(generated.NewExecutableSchema(c))

	gqlHandler.Use(logrusextension.LogrusExtension{
		Logger: app.Logger,
	})
	gqlHandler.Use(nrextension.NrExtension{
		NrApp: app.NrApp,
	})

	server.Setup(gqlHandler, &cfg.GraphQL, db)

	// register dataloader
	// router.Use(dataloaderMiddleware)

	server.Run()

}

func handleMQErrors(errors <-chan error) {
	for err := range errors {
		logrus.Printf("mq error: %s", err)
	}
}
