package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Session struct {
	ID             uuid.UUID `gorm:"column:id;type:char(36);primaryKey" json:"id"`
	AccountID      uuid.UUID `gorm:"column:account_id;type:char(36);not null;index" json:"account_id"`
	Name           string    `gorm:"column:name;size:255;not null" json:"name"`
	RefreshTokenID uuid.UUID `gorm:"column:refresh_token_id;type:char(36);not null" json:"refresh_token_id"`
	CreatedAt      time.Time `gorm:"column:created_at;not null" json:"created_at"`

	Account Account `gorm:"foreignKey:AccountID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
}

func (s *Session) TableName() string {
	return "sessions"
}

func (s *Session) BeforeCreate(_ *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}

	return nil
}
