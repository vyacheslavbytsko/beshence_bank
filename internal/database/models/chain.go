package models

import (
	"errors"
	"regexp"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var chainNamePattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

var (
	ErrInvalidChainName   = errors.New("chain name must contain only English letters, numbers, '_' or '-' characters")
	ErrChainVaultRequired = errors.New("chain vault is required")
)

type Chain struct {
	ID          uuid.UUID  `gorm:"column:id;type:char(36);primaryKey" json:"id"`
	Name        string     `gorm:"column:name;size:128;not null" json:"name"`
	VaultID     uuid.UUID  `gorm:"column:vault_id;type:char(36);not null" json:"vault_id"`
	LastEventID *uuid.UUID `gorm:"column:last_event_id;type:char(36)" json:"last_event_id,omitempty"`
	CreatedAt   time.Time  `gorm:"column:created_at;not null" json:"created_at"`

	LastEvent *Event `gorm:"foreignKey:LastEventID;references:EventID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"-"`
	Vault     Vault  `gorm:"foreignKey:VaultID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
}

func (c *Chain) TableName() string {
	return "vaults"
}

func (c *Chain) BeforeCreate(_ *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}

	return c.Validate()
}

func (c *Chain) Validate() error {
	if !chainNamePattern.MatchString(c.Name) {
		return ErrInvalidChainName
	}

	if c.VaultID == uuid.Nil {
		return ErrChainVaultRequired
	}

	return nil
}
