package misc

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const EndpointPing = "/ping"
const EndpointWellKnownBank = "/.well-known/beshence/bank"

func PingV1dot0(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"ping": "beshence-pong!",
		"api": gin.H{
			"urls":     []string{"https://127.0.0.1:27462/api"},
			"versions": []string{"v1.0"},
		},
		"auth": gin.H{
			"methods": []string{"usernameAndPassword"},
		},
	})
}
