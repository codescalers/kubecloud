package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	ID       uuid.UUID `gorm:"primary_key; unique; type:uuid; column:id"`
	Username string    `json:"username" binding:"required"`
	Email    string    `json:"email" gorm:"unique" binding:"required"`
	Password string    `json:"-"`
	Salt     string    `json:"-"`
	Admin    bool      `json:"admin"`
}

// BeforeCreate generates a new uuid
func (user *User) BeforeCreate(tx *gorm.DB) (err error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return err
	}

	user.ID = id
	return
}
