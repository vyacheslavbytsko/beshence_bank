package middleware

import (
	"bank/internal/auth"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	ContextAuthClaimsKey         = "auth.claims"
	ContextAuthSessionIDKey      = "auth.session_id"
	ContextAuthAccountIDKey      = "auth.account_id"
	ContextAuthRefreshTokenIDKey = "auth.refresh_token_id"
)

func GetCurrentAccount(c *gin.Context) (string, bool) {
	accountIDValue, accountIDExists := c.Get(ContextAuthAccountIDKey)
	if !accountIDExists {
		return "", false
	}

	accountID, accountIDOk := accountIDValue.(string)
	if !accountIDOk || accountID == "" {
		return "", false
	}

	return accountID, true
}

func GetCurrentSession(c *gin.Context) (string, string, bool) {
	sessionIDValue, sessionIDExists := c.Get(ContextAuthSessionIDKey)
	refreshTokenIDValue, refreshTokenIDExists := c.Get(ContextAuthRefreshTokenIDKey)
	if !sessionIDExists || !refreshTokenIDExists {
		return "", "", false
	}

	sessionID, sessionIDOk := sessionIDValue.(string)
	refreshTokenID, refreshTokenIDOk := refreshTokenIDValue.(string)
	if !sessionIDOk || !refreshTokenIDOk || sessionID == "" || refreshTokenID == "" {
		return "", "", false
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

		if authClaims.TokenType != expectedTokenType {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{

				"errcode": -1,
				"error":   "invalid token type",
			})
			return
		}

		c.Set(ContextAuthClaimsKey, claims)
		c.Set(ContextAuthSessionIDKey, authClaims.SessionID)
		c.Set(ContextAuthAccountIDKey, authClaims.AccountID)
		c.Set(ContextAuthRefreshTokenIDKey, authClaims.RefreshTokenID)
		c.Next()
	}
}
