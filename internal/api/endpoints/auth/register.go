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

				"errcode": -1,
				"error":   "auth is not configured",
			})
			return
		}

		var request registerRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{

				"errcode": -1,
				"error":   "invalid request body",
			})
			return
		}

		passwordHash, err := auth.HashPassword(request.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{

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

					"errcode": -1,
					"error":   "username is already taken",
				})
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{

				"errcode": -1,
				"error":   "failed to create account",
			})
			return
		}

		tokens, err := auth.IssueTokenPairForNewSession(deps.DB, deps.RefreshJWTManager, deps.AccessJWTManager, account, c.GetHeader("User-Agent"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{

				"errcode": -1,
				"error":   "failed to generate tokens",
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"errcode":       0,
			"id":            account.ID,
			"username":      account.Username,
			"refresh_token": tokens.RefreshToken,
			"access_token":  tokens.AccessToken,
		})
	}
}
