package middleware

import (
	"strings"

	apperrors "github.com/evelyndaianabejarano-coder/aluna-be/internal/errors"
	"github.com/evelyndaianabejarano-coder/aluna-be/internal/services"
	"github.com/gin-gonic/gin"
)

const (
	ContextKeyUserID = "userID"
	ContextKeyRole   = "role"
)

func Auth(tokenSvc services.TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			apperrors.Respond(c, apperrors.ErrInvalidToken)
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(header, "Bearer ")
		claims, err := tokenSvc.ValidateToken(tokenString)
		if err != nil {
			apperrors.Respond(c, apperrors.ErrInvalidToken)
			c.Abort()
			return
		}

		c.Set(ContextKeyUserID, claims.UserID)
		c.Set(ContextKeyRole, claims.Role)
		c.Next()
	}
}
