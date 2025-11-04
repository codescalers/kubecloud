package models

import (
	"time"

	"gorm.io/gorm"
)

const (
	ChargeBalanceMode = "charge_balance"
	RedeemVoucherMode = "redeem_voucher"
	AdminCreditMode   = "admin_credit"
)

type PendingRecord struct {
	ID       int    `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID   int    `json:"user_id" gorm:"not null"`
	Username string `json:"username"`
	// TFTs are multiplied by 1e7
	TFTAmount            uint64    `json:"tft_amount" gorm:"not null"`
	TransferredTFTAmount uint64    `json:"transferred_tft_amount" gorm:"not null"`
	TransferMode         string    `json:"transfer_mode"`
	CreatedAt            time.Time `json:"created_at" gorm:"not null"`
	UpdatedAt            time.Time `json:"updated_at" gorm:"not null"`
}

type GormPendingRecordRepository struct {
	db *gorm.DB
}

func NewGormPendingRecordRepository(db DB) *GormPendingRecordRepository {
	return &GormPendingRecordRepository{db: db.GetDB()}
}

func (r *GormPendingRecordRepository) CreatePendingRecord(record *PendingRecord) error {
	record.CreatedAt = time.Now()
	return r.db.Create(record).Error
}

func (r *GormPendingRecordRepository) ListAllPendingRecords() ([]PendingRecord, error) {
	var pendingRecords []PendingRecord
	return pendingRecords, r.db.Find(&pendingRecords).Error
}

func (r *GormPendingRecordRepository) ListOnlyPendingRecords() ([]PendingRecord, error) {
	var pendingRecords []PendingRecord
	return pendingRecords, r.db.Where("tft_amount > transferred_tft_amount").Find(&pendingRecords).Error
}

func (r *GormPendingRecordRepository) ListUserPendingRecords(userID int) ([]PendingRecord, error) {
	var pendingRecords []PendingRecord
	return pendingRecords, r.db.Where("user_id = ?", userID).Find(&pendingRecords).Error
}

func (r *GormPendingRecordRepository) UpdatePendingRecordTransferredAmount(id int, amount uint64) error {
	return r.db.Model(&PendingRecord{}).
		Where("id = ?", id).
		UpdateColumn("transferred_tft_amount", gorm.Expr("transferred_tft_amount + ?", amount)).
		UpdateColumn("updated_at", gorm.Expr("?", time.Now())).
		Error
}
