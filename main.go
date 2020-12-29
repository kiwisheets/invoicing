package main

//go:generate go run github.com/99designs/gqlgen generate

import (
	"fmt"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/aymerick/raymond"
	"github.com/emvi/hide"
	"github.com/kiwisheets/auth/directive"
	"github.com/kiwisheets/gql-server/client"
	"github.com/kiwisheets/invoicing/config"
	"github.com/kiwisheets/invoicing/graphql/generated"
	"github.com/kiwisheets/invoicing/graphql/resolver"
	"github.com/kiwisheets/invoicing/model"
	"github.com/kiwisheets/orm"
	"github.com/kiwisheets/server"
	"github.com/kiwisheets/util"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm/logger"
)

func main() {
	cfg := config.Server()

	logrus.SetOutput(os.Stdout)
	if cfg.Environment == "development" {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
		logrus.SetFormatter(&logrus.JSONFormatter{})
	}

	raymond.RegisterHelper("total", func(invoice *model.InvoiceTemplateData) string {
		return "$10.00"
	})

	raymond.RegisterHelper("itemtotal", func(item *model.LineItemInput) string {
		return "$2.50"
	})

	hide.UseHash(hide.NewHashID(cfg.Hash.Salt, cfg.Hash.MinLength))

	db := orm.Init(&cfg.Database)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	if cfg.GraphQL.Environment == "development" {
		db.Config.Logger = logger.Default.LogMode(logger.Info)
		directive.Development(true)

		db.Migrator().DropTable(&model.LineItem{})
		db.Migrator().DropTable(&model.Invoice{})
	}

	db.AutoMigrate(&model.LineItem{})
	db.AutoMigrate(&model.Invoice{})

	// i := model.Invoice{
	// 	CompanyID: 1,
	// 	CreatedBy: 1,
	// 	Client:    1,
	// 	LineItems: []model.LineItem{
	// 		{
	// 			Name:        "Test Item",
	// 			Description: "Item description",
	// 			UnitCost:    2.50,
	// 			TaxRate:     util.Float64(0),
	// 			Quantity:    1,
	// 		},
	// 	},
	// }

	// db.Create(&i)

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
			DB:              db,
			CreateProducer:  createProducer,
			RenderProducer:  renderProducer,
			GqlServerClient: client.NewClient(http.DefaultClient, cfg.GqlServerURL),
		},
		Directives: generated.DirectiveRoot{
			IsAuthenticated:       directive.IsAuthenticated,
			IsSecureAuthenticated: directive.IsSecureAuthenticated,
			HasPerm:               directive.HasPerm,
			HasPerms:              directive.HasPerms,
		},
	}

	gqlHandler := handler.New(generated.NewExecutableSchema(c))

	server.Setup(gqlHandler, &cfg.GraphQL, db)

	// register dataloader
	// router.Use(dataloaderMiddleware)

	// TODO: POSTSTART (no poststart) update apollo federated graph using endpoint, isolated script task
	// Find way to do this

	server.Run()

}

func handleMQErrors(errors <-chan error) {
	for err := range errors {
		logrus.Printf("mq error: %s", err)
	}
}
