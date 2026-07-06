package bank

import (
	"bank/internal/api"
	"bank/internal/database/models"
	"bank/internal/middleware"
	"bank/internal/misc"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type addEventRequest struct {
	EventID  string  `json:"id" binding:"required"`
	ParentID *string `json:"parent_id"`
	Payload  string  `json:"payload" binding:"required"`
}

func AddEventV1dot0(deps *api.Dependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		if deps == nil || deps.DB == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"errcode": -1, "error": "database is not configured"})
			return
		}

		accountID, ok := middleware.GetCurrentAccount(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"errcode": -1, "error": "unauthorized"})
			return
		}

		vaultID, err := uuid.Parse(c.Param("vaultId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"errcode": -1,
				"error":   "invalid vault id",
			})
			return
		}

		chainName := c.Param("chainName")

		// TODO: check for name

		if _, err := loadVaultForAccount(deps.DB, vaultID, accountID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{
					"errcode": -1,
					"error":   "vault not found",
				})
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{
				"errcode": -1,
				"error":   "failed to load vault",
			})
			return
		}

		var request addEventRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"errcode": -1,
				"error":   "invalid request body",
			})
			return
		}

		eventID, err := uuid.Parse(request.EventID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"errcode": -1,
				"error":   "invalid event id",
			})
			return
		}

		parentID, err := misc.ParseOptionalUUID(request.ParentID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"errcode": -1,
				"error":   "invalid parent_id",
			})
			return
		}

		tx := deps.DB.Begin()
		if tx.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"errcode": -1,
				"error":   "failed to start transaction",
			})
			return
		}

		rollback := func() {
			_ = tx.Rollback().Error
		}

		if _, err := loadVaultForAccount(tx, vaultID, accountID); err != nil {
			rollback()
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{
					"errcode": -1,
					"error":   "vault not found",
				})
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{
				"errcode": -1,
				"error":   "failed to load vault",
			})
			return
		}

		var chain models.Chain

		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("name = ? AND vault_id = ?", chainName, vaultID).Take(&chain).Error; err != nil {
			rollback()
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{
					"errcode": -1,
					"error":   "chain not found",
				})
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{
				"errcode": -1,
				"error":   "failed to load chain",
			})
			return
		}

		if !misc.SameOptionalUUID(parentID, chain.LastEventID) {
			rollback()
			c.JSON(http.StatusConflict, gin.H{
				"errcode":       -1,
				"error":         "last event id is not same",
				"last_event_id": misc.OptionalUUIDToString(chain.LastEventID),
			})
			return
		}

		event := models.Event{
			ID:        eventID,
			ChainName: chainName,
			VaultID:   vaultID,
			ParentID:  parentID,
			Payload:   request.Payload,
			CreatedAt: time.Now(),
		}

		result := tx.Clauses(clause.OnConflict{Columns: []clause.Column{
			{Name: "id"},
			{Name: "chain_name"},
			{Name: "vault_id"},
		}, DoNothing: true}).Create(&event)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
				rollback()
				c.JSON(http.StatusConflict, gin.H{
					"errcode":       -1,
					"error":         "last event id is not same",
					"last_event_id": misc.OptionalUUIDToString(chain.LastEventID),
				})
				return
			}

			rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"errcode": -1,
				"error":   "failed to insert event",
			})
			return
		}

		if result.RowsAffected == 0 {
			var existing models.Event
			if err := tx.Where("vault_id = ? AND chain_name = ? AND id = ?", vaultID, chainName, eventID).Take(&existing).Error; err != nil {
				rollback()
				if errors.Is(err, gorm.ErrRecordNotFound) {
					c.JSON(http.StatusConflict, gin.H{
						"errcode":       -1,
						"error":         "last event id is not same",
						"last_event_id": misc.OptionalUUIDToString(chain.LastEventID),
					})
					return
				}

				c.JSON(http.StatusInternalServerError, gin.H{
					"errcode": -1,
					"error":   "failed to check existing event",
				})
				return
			}

			if !misc.SameOptionalUUID(existing.ParentID, parentID) || existing.Payload != request.Payload {
				rollback()
				c.JSON(http.StatusConflict, gin.H{
					"errcode":       -1,
					"error":         "last event id is not same",
					"last_event_id": misc.OptionalUUIDToString(chain.LastEventID),
				})
				return
			}

			if err := tx.Commit().Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"errcode": -1,
					"error":   "failed to complete transaction",
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"errcode": 0,
			})
			return
		}

		if err := tx.Model(&models.Chain{}).Where("name = ? AND vault_id = ?", chainName, vaultID).Update("last_event_id", eventID).Error; err != nil {
			rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"errcode": -1,
				"error":   "failed to update chain state",
			})
			return
		}

		if err := tx.Commit().Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"errcode": -1,
				"error":   "failed to complete transaction",
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"errcode": 0,
			"error":   "event created successfully",
		})
	}
}
