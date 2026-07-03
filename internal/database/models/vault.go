package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Vault struct {
	ID        uuid.UUID `gorm:"column:id;type:char(36);primaryKey" json:"id"`
	Name      string    `gorm:"column:name;size:128;not null" json:"name"`
	AccountID uuid.UUID `gorm:"column:account_id;type:char(36);not null" json:"account_id"`
	CreatedAt time.Time `gorm:"column:created_at;not null" json:"created_at"`

	Account Account `gorm:"foreignKey:AccountID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
}

func (v *Vault) TableName() string {
	return "vaults"
}

func (v *Vault) BeforeCreate(_ *gorm.DB) error {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}

	return nil
}
