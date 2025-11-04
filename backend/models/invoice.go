package models

import (
	"time"

	"gorm.io/gorm"
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
	InvoiceID     int       `json:"invoice_id" gorm:"index;constraint:OnDelete:CASCADE"`
	NodeID        uint32    `json:"node_id"`
	ContractID    uint64    `json:"contract_id"`
	RentCreatedAt time.Time `json:"rent_created_at"`
	PeriodInHours float64   `json:"period"`
	Cost          float64   `json:"cost"`
}

type GormInvoiceRepository struct {
	db *gorm.DB
}

func NewGormInvoiceRepository(db DB) *GormInvoiceRepository {
	return &GormInvoiceRepository{db: db.GetDB()}
}

// CreateInvoice creates new invoice
func (r *GormInvoiceRepository) CreateInvoice(invoice *Invoice) error {
	return r.db.Create(&invoice).Error
}

// GetInvoice returns an invoice by ID
func (r *GormInvoiceRepository) GetInvoice(id int) (Invoice, error) {
	var invoice Invoice
	err := r.db.First(&invoice, id).Error
	if err != nil {
		return Invoice{}, err
	}

	var nodes []NodeItem
	if err = r.db.Model(&invoice).Association("Nodes").Find(&nodes); err != nil {
		return Invoice{}, err
	}

	invoice.Nodes = nodes
	return invoice, nil
}

// ListUserInvoices returns all invoices of user
func (r *GormInvoiceRepository) ListUserInvoices(userID int) ([]Invoice, error) {
	var invoices []Invoice
	err := r.db.Where("user_id = ?", userID).Find(&invoices).Error
	if err != nil {
		return nil, err
	}

	for idx := range invoices {
		invoices[idx], err = r.GetInvoice(invoices[idx].ID)
		if err != nil {
			return nil, err
		}
	}
	return invoices, nil
}

// ListInvoices returns all invoices (admin)
func (r *GormInvoiceRepository) ListInvoices() ([]Invoice, error) {
	var invoices []Invoice
	err := r.db.Find(&invoices).Error

	if err != nil {
		return nil, err
	}

	for idx := range invoices {
		invoices[idx], err = r.GetInvoice(invoices[idx].ID)
		if err != nil {
			return nil, err
		}
	}
	return invoices, nil
}

func (r *GormInvoiceRepository) UpdateInvoicePDF(id int, data []byte) error {
	return r.db.Model(&Invoice{}).Where("id = ?", id).Updates(map[string]interface{}{"file_data": data}).Error
}
