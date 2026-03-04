package models

import (
	"github.com/google/uuid"
)

type Wallet struct {
	ID      uuid.UUID `gorm:"type:uuid; primaryKey; default:gen_random_uuid()"`
	Balance int64     `gorm:"default:0; check:balance>=0"`
	Version uint      `gorm:"version"`
}
