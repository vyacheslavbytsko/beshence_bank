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

type loginRequest struct {
	Username string `json:"username" binding:"required,min=3,max=64"`
	Password string `json:"password" binding:"required,min=8,max=128"`
}

func LoginV1(deps *api.Dependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		if deps == nil || deps.DB == nil || deps.AccessJWTManager == nil || deps.RefreshJWTManager == nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"err":    "UNKNOWN",
				"errmsg": "auth is not configured",
			})
			return
		}

		var request loginRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"err":    "UNKNOWN",
				"errmsg": "invalid request body",
			})
			return
		}

		var account models.Account
		if err := deps.DB.Where("username = ?", request.Username).First(&account).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusUnauthorized, gin.H{
					"err":    "UNKNOWN",
					"errmsg": "invalid credentials",
				})
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{
				"err":    "UNKNOWN",
				"errmsg": "failed to load account",
			})
			return
		}

		ok, err := auth.VerifyPassword(request.Password, account.PasswordHash)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"err":    "UNKNOWN",
				"errmsg": "failed to verify password",
			})
			return
		}

		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"err":    "UNKNOWN",
				"errmsg": "invalid credentials",
			})
			return
		}

		tokens, err := auth.IssueTokenPairForNewSession(deps.DB, deps.RefreshJWTManager, deps.AccessJWTManager, account, c.GetHeader("User-Agent"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"err":    "UNKNOWN",
				"errmsg": "failed to generate tokens",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"err":           "0",
			"id":            account.ID,
			"username":      account.Username,
			"refresh_token": tokens.RefreshToken,
			"access_token":  tokens.AccessToken,
		})
	}
}
