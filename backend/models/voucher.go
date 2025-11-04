package models

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Voucher struct holds all data for vouchers, voucher used only by one user.db.
// Voucher struct holds all data for vouchers, voucher used only by one user.db.
type Voucher struct {
	ID        int       `json:"id" gorm:"primaryKey;autoIncrement"`
	Code      string    `json:"code" gorm:"unique;not null" validate:"required"`
	Value     float64   `json:"value" gorm:"not null" validate:"required,gt=0"`
	Redeemed  bool      `json:"redeemed" gorm:"default:false"`
	CreatedAt time.Time `json:"created_at" validate:"required"`
	ExpiresAt time.Time `json:"expires_at" validate:"required,gtfield=CreatedAt"`
}

type GormVoucherRepository struct {
	db *gorm.DB
}

func NewGormVoucherRepository(db DB) *GormVoucherRepository {
	return &GormVoucherRepository{db: db.GetDB()}
}

// CreateVoucher creates new voucher in system
func (r *GormVoucherRepository) CreateVoucher(voucher *Voucher) error {
	return r.db.Create(voucher).Error
}

// ListAllVouchers gets all vouchers in system
func (r *GormVoucherRepository) ListAllVouchers() ([]Voucher, error) {
	var vouchers []Voucher

	err := r.db.Find(&vouchers).Error
	if err != nil {
		return nil, err
	}
	return vouchers, nil
}

// GetVoucherByCode returns voucher by its code
func (r *GormVoucherRepository) GetVoucherByCode(code string) (Voucher, error) {
	var voucher Voucher
	query := r.db.First(&voucher, "code = ?", code)
	return voucher, query.Error
}

// RedeemVoucher updates status if voucher
func (r *GormVoucherRepository) RedeemVoucher(code string) error {
	result := r.db.Model(&Voucher{}).
		Where("code = ?", code).
		Update("redeemed", true)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("no voucher found with Code %s", code)
	}

	return nil
}
