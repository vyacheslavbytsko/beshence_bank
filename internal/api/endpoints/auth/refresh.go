package auth

import (
	"bank/internal/api"
	"bank/internal/auth"
	"bank/internal/database/models"
	"bank/internal/middleware"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RefreshV1dot0(deps *api.Dependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		if deps == nil || deps.DB == nil || deps.AccessJWTManager == nil || deps.RefreshJWTManager == nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "auth is not configured",
			})
			return
		}

		accountID, ok := middleware.GetCurrentAccount(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorized",
			})
			return
		}

		sessionID, tokenRefreshTokenID, ok := middleware.GetCurrentSession(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorized",
			})
			return
		}

		var session models.Session
		if err := deps.DB.Where("id = ? AND account_id = ?", sessionID, accountID).First(&session).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "invalid session",
				})
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to load session",
			})
			return
		}

		if session.RefreshTokenID != tokenRefreshTokenID {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "refresh token is no longer valid",
			})
			return
		}

		var account models.Account
		if err := deps.DB.Select("id", "login").Where("id = ?", accountID).First(&account).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "unauthorized",
				})
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to load account",
			})
			return
		}

		tokens, err := auth.IssueTokenPairForExistingSession(deps.DB, deps.AccessJWTManager, deps.RefreshJWTManager, account, session, tokenRefreshTokenID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "refresh token is no longer valid",
				})
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to generate tokens",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"errcode":       0,
			"id":            account.ID,
			"username":      account.Username,
			"refresh_token": tokens.RefreshToken,
			"access_token":  tokens.AccessToken,
		})
	}
}
