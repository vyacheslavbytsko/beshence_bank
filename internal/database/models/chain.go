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
	Name        string     `gorm:"column:name;size:128;primaryKey" json:"name"`
	VaultID     uuid.UUID  `gorm:"column:vault_id;type:char(36);primaryKey;index:idx_chain_vault_created" json:"vault_id"`
	LastEventID *uuid.UUID `gorm:"column:last_event_id;type:char(36)" json:"last_event_id,omitempty"`
	CreatedAt   time.Time  `gorm:"column:created_at;not null;index:idx_chain_vault_created" json:"created_at"`

	LastEvent *Event `gorm:"foreignKey:LastEventID,Name,VaultID;references:ID,ChainName,VaultID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"-"`
	Vault     Vault  `gorm:"foreignKey:VaultID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
}

func (c *Chain) TableName() string {
	return "chains"
}

func (c *Chain) BeforeCreate(_ *gorm.DB) error {
	return c.Validate()
}

func (c *Chain) Validate() error {
	if !chainNamePattern.MatchString(c.Name) {
		return ErrInvalidChainName
	}

	if c.VaultID == uuid.Nil {
		return ErrChainVaultRequired
	}

	// TODO: check if vault already has chain with this name. i know we have it in api but still

	return nil
}
