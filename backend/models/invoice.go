package models

import (
	"time"
)

type Invoice struct {
	ID     int        `json:"id" gorm:"primaryKey"`
	UserID int        `json:"user_id" binding:"required"`
	Total  float64    `json:"total"`
	Nodes  []NodeItem `json:"nodes" gorm:"foreignKey:invoice_id"`
	// TODO:
	Tax       float64   `json:"tax"`
	CreatedAt time.Time `json:"created_at"`
	FileData  []byte    `json:"-" gorm:"type:bytea;column:file_data"`
}

type NodeItem struct {
	ID            int       `json:"id" gorm:"primaryKey"`
	InvoiceID     int       `json:"invoice_id"`
	NodeID        uint32    `json:"node_id"`
	ContractID    uint64    `json:"contract_id"`
	RentCreatedAt time.Time `json:"rent_created_at"`
	PeriodInHours float64   `json:"period"`
	Cost          float64   `json:"cost"`
}
