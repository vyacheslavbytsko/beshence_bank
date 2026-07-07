package auth

import (
	"bank/internal/api"
	"bank/internal/database/models"
	"bank/internal/middleware"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func MeV1dot0(deps *api.Dependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		if deps == nil || deps.DB == nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"errcode": -1,
				"error":   "database is not configured",
			})
			return
		}

		accountID, ok := middleware.GetCurrentAccount(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"errcode": 401,
				"error":   "unauthorized",
			})
			return
		}

		// TODO: use middleware
		var account models.Account
		if err := deps.DB.Select("id", "username").Where("id = ?", accountID.String()).First(&account).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusUnauthorized, gin.H{
					"errcode": 401,
					"error":   "unauthorized",
				})
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{
				"errcode": -1,
				"error":   "failed to load account",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"errcode":  0,
			"id":       account.ID,
			"username": account.Username,
		})
	}
}
