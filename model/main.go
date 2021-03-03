package model

import (
	"github.com/aymerick/raymond"
	"github.com/kiwisheets/orm"
	"github.com/kiwisheets/util"
	"gorm.io/gorm"
)

type Orm struct {
	DB *gorm.DB
}

func (o *Orm) Close() {
	sqlDB, _ := o.DB.DB()
	sqlDB.Close()
}

func Init(cfg *util.DatabaseConfig) *Orm {
	orm := Orm{
		DB: orm.Init(cfg),
	}

	migrate(orm.DB)
	registerRaymondHelpers()

	return &orm
}

func registerRaymondHelpers() {
	raymond.RegisterHelper("total", InvoiceTotalHelper)
	raymond.RegisterHelper("itemCost", InvoiceItemCostHelper)
	raymond.RegisterHelper("itemTotal", InvoiceItemTotalHelper)
}

func migrate(db *gorm.DB) {
	db.AutoMigrate(&LineItem{})
	db.AutoMigrate(&Invoice{})
	db.AutoMigrate(&Company{})
	db.AutoMigrate(&Address{})
	db.AutoMigrate(&Contact{})
	db.AutoMigrate(&Client{})
}
