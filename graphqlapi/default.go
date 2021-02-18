package graphqlapi

import (
	"os"

	"github.com/emvi/hide"
	"github.com/joho/godotenv"
	"github.com/kiwisheets/util"
	"github.com/maxtroughear/goenv"
	"github.com/maxtroughear/logrusnrhook"
	"github.com/newrelic/go-agent/v3/integrations/logcontext/nrlogrusplugin"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/sirupsen/logrus"
)

type App struct {
	NrApp   *newrelic.Application
	AppName string
	Logger  *logrus.Entry
}

type env struct {
	appName      string
	nrLicenseKey string
	environment  string
	hashCfg      util.HashConfig
}

func NewDefault() App {
	env := getEnv()

	hide.UseHash(hide.NewHashID(env.hashCfg.Salt, env.hashCfg.MinLength))

	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.DebugLevel)

	hostname, _ := os.Hostname()

	app := App{
		AppName: env.appName,
		Logger: logrus.WithFields(logrus.Fields{
			"service":  env.appName,
			"env":      env.environment,
			"hostname": hostname,
		}),
	}

	if env.environment == "production" {
		logrus.SetLevel(logrus.InfoLevel)
		logrus.SetFormatter(nrlogrusplugin.ContextFormatter{})
		logrus.AddHook(logrusnrhook.NewNrHook(env.appName, env.nrLicenseKey, true))

		var err error
		if app.NrApp, err = newrelic.NewApplication(
			newrelic.ConfigAppName(env.appName),
			newrelic.ConfigLicense(env.nrLicenseKey),
			newrelic.ConfigDistributedTracerEnabled(true),
			func(cfg *newrelic.Config) {
				cfg.ErrorCollector.RecordPanics = true
			},
			// newrelic.ConfigLogger(nrlogrus.StandardLogger()),
		); err != nil {
			logrus.Errorf("failed to start new relic agent %v", err)
		}
	}

	return app
}

func getEnv() env {
	godotenv.Load()
	return env{
		appName:      goenv.CanGet("APP_NAME", "unnamed"),
		nrLicenseKey: goenv.CanGet("NR_LICENSE_KEY", ""),
		environment:  goenv.MustGet("ENVIRONMENT"),
		hashCfg: util.HashConfig{
			Salt:      goenv.MustGetSecretFromEnv("HASH_SALT"),
			MinLength: goenv.CanGetInt32("HASH_MIN_LENGTH", 10),
		},
	}
}
