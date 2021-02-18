package model

import (
	"time"

	"github.com/emvi/hide"
)

type Contact struct {
	ID               hide.ID `gorm:"type: bigserial;primary_key" json:"id"` // int64
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time
	ClientID         hide.ID
	Email            *string
	Phone            *string
	Mobile           *string
	PreferredContact *PreferredContact
	Firstname        string
	Lastname         string
}
