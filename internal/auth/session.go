package auth

import (
	"bank/internal/database/models"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const maxSessionNameLength = 255

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

func IssueTokenPairForNewSession(db *gorm.DB, refreshJWTManager *JWT, accessJWTManager *JWT, account models.Account, sessionName string) (TokenPair, error) {
	sessionName = normalizeSessionName(sessionName)

	refreshTokenID := uuid.New()
	session := models.Session{
		AccountID:      account.ID,
		Name:           sessionName,
		RefreshTokenID: refreshTokenID,
	}

	var tokens TokenPair
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&session).Error; err != nil {
			return err
		}

		pair, err := generateTokenPairForSession(refreshJWTManager, accessJWTManager, account, session.ID, refreshTokenID)
		if err != nil {
			return err
		}

		tokens = pair
		return nil
	})
	if err != nil {
		return TokenPair{}, err
	}

	return tokens, nil
}

func IssueTokenPairForExistingSession(db *gorm.DB, refreshJWTManager *JWT, accessJWTManager *JWT, account models.Account, session models.Session, existingRefreshTokenID uuid.UUID) (TokenPair, error) {
	newRefreshTokenID := uuid.New()

	var tokens TokenPair
	err := db.Transaction(func(tx *gorm.DB) error {
		pair, err := generateTokenPairForSession(refreshJWTManager, accessJWTManager, account, session.ID, newRefreshTokenID)
		if err != nil {
			return err
		}

		updateResult := tx.Model(&models.Session{}).
			Where("id = ? AND account_id = ? AND refresh_token_id = ?", session.ID, account.ID, existingRefreshTokenID).
			Update("refresh_token_id", newRefreshTokenID)
		if updateResult.Error != nil {
			return updateResult.Error
		}

		if updateResult.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		tokens = pair
		return nil
	})
	if err != nil {
		return TokenPair{}, err
	}

	return tokens, nil
}

func generateTokenPairForSession(refreshJWTManager *JWT, accessJWTManager *JWT, account models.Account, sessionID uuid.UUID, refreshTokenID uuid.UUID) (TokenPair, error) {
	refreshToken, err := refreshJWTManager.GenerateToken(sessionID, account.ID, refreshTokenID)
	if err != nil {
		return TokenPair{}, err
	}

	accessToken, err := accessJWTManager.GenerateToken(sessionID, account.ID, refreshTokenID)
	if err != nil {
		return TokenPair{}, err
	}

	return TokenPair{
		RefreshToken: refreshToken,
		AccessToken:  accessToken,
	}, nil
}

func normalizeSessionName(sessionName string) string {
	sessionName = strings.TrimSpace(sessionName)
	if sessionName == "" {
		return "unknown"
	}

	runes := []rune(sessionName)
	if len(runes) > maxSessionNameLength {
		return string(runes[:maxSessionNameLength])
	}

	return sessionName
}
