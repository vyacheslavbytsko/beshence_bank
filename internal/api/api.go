package api

import (
	"bank/internal/auth"

	"gorm.io/gorm"
)

type Dependencies struct {
	DB                *gorm.DB
	RefreshJWTManager *auth.JWT
	AccessJWTManager  *auth.JWT
}

func NewDependencies(
	db *gorm.DB,
	refresh *auth.JWT,
	access *auth.JWT,
) *Dependencies {
	return &Dependencies{
		DB:                db,
		RefreshJWTManager: refresh,
		AccessJWTManager:  access,
	}
}
