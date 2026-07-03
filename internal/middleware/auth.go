package middleware

import (
	"bank/internal/auth"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	ContextAuthClaimsKey         = "auth.claims"
	ContextAuthSessionIDKey      = "auth.session_id"
	ContextAuthAccountIDKey      = "auth.account_id"
	ContextAuthRefreshTokenIDKey = "auth.refresh_token_id"
)

func GetCurrentAccount(c *gin.Context) (uuid.UUID, bool) {
	accountIDValue, accountIDExists := c.Get(ContextAuthAccountIDKey)
	if !accountIDExists {
		return uuid.Nil, false
	}

	accountID, accountIDOk := accountIDValue.(uuid.UUID)
	if !accountIDOk || accountID == uuid.Nil {
		return uuid.Nil, false
	}

	return accountID, true
}

func GetCurrentSession(c *gin.Context) (uuid.UUID, uuid.UUID, bool) {
	sessionIDValue, sessionIDExists := c.Get(ContextAuthSessionIDKey)
	refreshTokenIDValue, refreshTokenIDExists := c.Get(ContextAuthRefreshTokenIDKey)
	if !sessionIDExists || !refreshTokenIDExists {
		return uuid.Nil, uuid.Nil, false
	}

	sessionID, sessionIDOk := sessionIDValue.(uuid.UUID)
	refreshTokenID, refreshTokenIDOk := refreshTokenIDValue.(uuid.UUID)
	if !sessionIDOk || !refreshTokenIDOk || sessionID == uuid.Nil || refreshTokenID == uuid.Nil {
		return uuid.Nil, uuid.Nil, false
	}

	return sessionID, refreshTokenID, true
}

func RequireAuth(jwt *auth.JWT, tt auth.TokenType, h gin.HandlerFunc) gin.HandlerFunc {
	mw := CheckAuth(jwt, tt)
	return func(c *gin.Context) {
		mw(c)
		if c.IsAborted() {
			return
		}
		h(c)
	}
}

func CheckAuth(jwtManager *auth.JWT, expectedTokenType auth.TokenType) gin.HandlerFunc {
	return func(c *gin.Context) {
		if jwtManager == nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{

				"errcode": -1,
				"error":   "auth is not configured",
			})
			return
		}

		authorizationHeader := c.GetHeader("Authorization")
		if !strings.HasPrefix(authorizationHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{

				"errcode": -1,
				"error":   "missing or invalid authorization header",
			})
			return
		}

		token := strings.TrimSpace(strings.TrimPrefix(authorizationHeader, "Bearer "))
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{

				"errcode": -1,
				"error":   "missing bearer token",
			})
			return
		}

		claims, err := jwtManager.ParseToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{

				"errcode": -1,
				"error":   "invalid token",
			})
			return
		}

		authClaims, claimsOk := auth.ClaimsFromToken(claims)
		if !claimsOk {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{

				"errcode": -1,
				"error":   "invalid token claims",
			})
			return
		}

		sessionID, err := uuid.Parse(authClaims.SessionID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{

				"errcode": -1,
				"error":   "invalid token claims",
			})
			return
		}

		accountID, err := uuid.Parse(authClaims.AccountID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{

				"errcode": -1,
				"error":   "invalid token claims",
			})
			return
		}

		refreshTokenID := uuid.Nil
		if authClaims.RefreshTokenID != "" {
			refreshTokenID, err = uuid.Parse(authClaims.RefreshTokenID)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{

					"errcode": -1,
					"error":   "invalid token claims",
				})
				return
			}
		}

		if authClaims.TokenType != expectedTokenType {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{

				"errcode": -1,
				"error":   "invalid token type",
			})
			return
		}

		c.Set(ContextAuthClaimsKey, claims)
		c.Set(ContextAuthSessionIDKey, sessionID)
		c.Set(ContextAuthAccountIDKey, accountID)
		c.Set(ContextAuthRefreshTokenIDKey, refreshTokenID)
		c.Next()
	}
}
