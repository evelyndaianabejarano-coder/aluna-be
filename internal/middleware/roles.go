package middleware

import (
	apperrors "github.com/evelyndaianabejarano-coder/aluna-be/internal/errors"
	"github.com/evelyndaianabejarano-coder/aluna-be/internal/models"
	"github.com/gin-gonic/gin"
)

func RequireRole(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}

	return func(c *gin.Context) {
		role, exists := c.Get(ContextKeyRole)
		if !exists {
			apperrors.Respond(c, apperrors.ErrUnauthorized)
			c.Abort()
			return
		}

		userRole, ok := role.(models.Role)
		if !ok {
			apperrors.Respond(c, apperrors.ErrUnauthorized)
			c.Abort()
			return
		}

		if _, permitted := allowed[string(userRole)]; !permitted {
			apperrors.Respond(c, apperrors.ErrUnauthorized)
			c.Abort()
			return
		}

		c.Next()
	}
}
