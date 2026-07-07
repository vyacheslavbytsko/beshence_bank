package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Event struct {
	ID        uuid.UUID  `gorm:"column:id;type:char(36);primaryKey" json:"id"`
	ChainName string     `gorm:"column:chain_name;size:128;primaryKey" json:"chain_name"`
	VaultID   uuid.UUID  `gorm:"column:vault_id;type:char(36);primaryKey" json:"vault_id"`
	ParentID  *uuid.UUID `gorm:"column:parent_id;type:char(36)" json:"parent_id,omitempty"`
	Payload   string     `gorm:"column:payload;type:text;not null" json:"payload"`
	CreatedAt time.Time  `gorm:"column:created_at;not null" json:"created_at"`

	Chain  *Chain `gorm:"foreignKey:ChainName,VaultID;references:Name,VaultID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
	Parent *Event `gorm:"foreignKey:ParentID,ChainName,VaultID;references:ID,ChainName,VaultID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"-"`
}

func (e *Event) BeforeCreate(_ *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}

	return nil
}
