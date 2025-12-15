// pkg/auth/middleware.go
package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Middleware 创建认证中间件
func Middleware(authenticator Authenticator) gin.HandlerFunc {
	return func(c *gin.Context) {
		authorization := c.GetHeader("Authorization")

		if !authenticator.Authenticate(authorization) {
			c.String(http.StatusUnauthorized, "Unauthorised")
			c.Abort()
			return
		}

		c.Next()
	}
}
