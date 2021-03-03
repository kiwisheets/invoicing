package resolver

import (
	gqlClient "github.com/kiwisheets/gql-server/client"
	"github.com/kiwisheets/invoicing/mq"
	"gorm.io/gorm"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	DB        *gorm.DB
	MQ        *mq.MQ
	GqlClient *gqlClient.Client
}
