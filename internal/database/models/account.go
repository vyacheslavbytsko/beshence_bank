package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Account struct {
	ID           uuid.UUID `gorm:"column:id;type:char(36);primaryKey" json:"id"`
	Username     string    `gorm:"column:username;size:64;not null;unique" json:"username"`
	PasswordHash string    `gorm:"column:password_hash;not null" json:"-"`
	CreatedAt    time.Time `gorm:"column:created_at;not null" json:"created_at"`
}

func (a *Account) TableName() string {
	return "accounts"
}

func (a *Account) BeforeCreate(_ *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}

	return nil
}
