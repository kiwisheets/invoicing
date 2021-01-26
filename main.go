package main

//go:generate go run github.com/99designs/gqlgen generate

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/kiwisheets/auth/directive"
	"github.com/kiwisheets/gql-server/client"
	"github.com/kiwisheets/orm"
	"github.com/kiwisheets/server"
	"github.com/kiwisheets/util"
	"github.com/maxtroughear/goenv"

	"github.com/kiwisheets/invoicing/config"
	"github.com/kiwisheets/invoicing/graphql/generated"
	"github.com/kiwisheets/invoicing/graphql/resolver"
	"github.com/kiwisheets/invoicing/logger"
	"github.com/kiwisheets/invoicing/model"
	"github.com/kiwisheets/invoicing/nrextension"
	"github.com/kiwisheets/invoicing/nrhook"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/aymerick/raymond"
	"github.com/emvi/hide"
	"github.com/newrelic/go-agent/v3/integrations/logcontext/nrlogrusplugin"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/sethgrid/pester"
	"github.com/sirupsen/logrus"
	gormlogger "gorm.io/gorm/logger"
)

const (
	appName = "Invoicing"
)

func main() {
	cfg := config.Server()

	nrLicenseKey := goenv.MustGetSecretFromEnv("NR_LICENSE_KEY")

	logrus.SetOutput(os.Stdout)
	logrus.SetFormatter(nrlogrusplugin.ContextFormatter{})
	logrus.AddHook(nrhook.NewNrHook(appName, nrLicenseKey))
	if cfg.Environment == "development" {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	hostname, _ := os.Hostname()
	log := logrus.WithFields(logrus.Fields{
		"service":  appName,
		"env":      cfg.Environment,
		"hostname": hostname,
	})

	app, err := newrelic.NewApplication(
		newrelic.ConfigAppName(appName),
		newrelic.ConfigLicense(nrLicenseKey),
		newrelic.ConfigDistributedTracerEnabled(true),
		func(cfg *newrelic.Config) {
			cfg.ErrorCollector.RecordPanics = true
		},
	)
	if err != nil {
		logrus.Errorf("failed to start new relic agent %v", err)
	}
	defer app.Shutdown(30 * time.Second)

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
		db.Config.Logger = gormlogger.Default.LogMode(gormlogger.Info)
		directive.Development(true)

		db.Migrator().DropTable(&model.LineItem{})
		db.Migrator().DropTable(&model.Invoice{})
	}

	db.AutoMigrate(&model.LineItem{})
	db.AutoMigrate(&model.Invoice{})

	// create models owned by the invoice domain
	// db.AutoMigrate(&gqlservermodel.Company{})
	// db.AutoMigrate(&gqlservermodel.Client{})
	// db.AutoMigrate(&gqlservermodel.Contact{})
	// db.AutoMigrate(&gqlservermodel.Address{})

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

	gqlHandler.Use(logger.LogrusExtension{
		Logger: log,
	})
	gqlHandler.Use(nrextension.NrExtension{
		NrApp: app,
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
