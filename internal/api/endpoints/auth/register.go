package auth

import (
	"bank/internal/api"
	"bank/internal/auth"
	"errors"
	"net/http"

	"bank/internal/database/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type registerRequest struct {
	Username string `json:"username" binding:"required,min=3,max=64"`
	Password string `json:"password" binding:"required,min=8,max=128"`
}

func RegisterV1dot0(deps *api.Dependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		if deps == nil || deps.DB == nil || deps.AccessJWTManager == nil || deps.RefreshJWTManager == nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"errcode": -1,
				"error":   "auth is not configured",
			})
			return
		}

		var request registerRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"errcode": -1,
				"error":   "invalid request body",
			})
			return
		}

		passwordHash, err := auth.HashPassword(request.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"errcode": -1,
				"error":   "failed to hash password",
			})
			return
		}

		account := models.Account{
			Username:     request.Username,
			PasswordHash: passwordHash,
		}

		if err := deps.DB.Create(&account).Error; err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				c.JSON(http.StatusConflict, gin.H{
					"success": false,
					"errcode": -1,
					"error":   "username is already taken",
				})
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"errcode": -1,
				"error":   "failed to create account",
			})
			return
		}

		tokens, err := auth.IssueTokenPairForNewSession(deps.DB, deps.AccessJWTManager, deps.RefreshJWTManager, account, c.GetHeader("User-Agent"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"errcode": -1,
				"error":   "failed to generate tokens",
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"success":       true,
			"id":            account.ID,
			"username":      account.Username,
			"access_token":  tokens.AccessToken,
			"refresh_token": tokens.RefreshToken,
		})
	}
}
