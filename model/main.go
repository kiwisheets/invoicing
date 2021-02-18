package model

import "gorm.io/gorm"

func Migrate(db *gorm.DB) {
	db.AutoMigrate(&LineItem{})
	db.AutoMigrate(&Invoice{})
	db.AutoMigrate(&Company{})
	db.AutoMigrate(&Address{})
	db.AutoMigrate(&Contact{})
	db.AutoMigrate(&Client{})
}
