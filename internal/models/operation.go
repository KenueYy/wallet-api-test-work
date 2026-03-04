package models

import (
	"github.com/google/uuid"
)

type OperationType string

const (
	DEPOSIT  OperationType = "DEPOSIT"
	WITHDRAW OperationType = "WITHDRAW"
)

type Operation struct {
	WalletID  uuid.UUID     `json:"wallet_id" binding:"required uuid"`
	Operation OperationType `json:"operationType" binding:"required oneof=DEPOSIT WITHDRAW"`
	Amount    int64         `json:"amount" binding:"required gt=0"`
}
