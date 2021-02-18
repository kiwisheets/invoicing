package resolver

import (
	"github.com/cheshir/go-mq"
	gqlClient "github.com/kiwisheets/gql-server/client"
	"gorm.io/gorm"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	DB              *gorm.DB
	CreateProducer  mq.SyncProducer
	RenderProducer  mq.SyncProducer
	GqlServerClient *gqlClient.Client
}
