package database

import (
	"bank/internal/database/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func New(databaseURL string) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		TranslateError:                           true,
		DisableForeignKeyConstraintWhenMigrating: true,
	})
}

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&models.Account{},
		&models.Session{},
		&models.Vault{},
		&models.Chain{},
		&models.Event{},
	); err != nil {
		return err
	}

	return nil
}
