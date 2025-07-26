package middleware

import (
	"net/http"
	"strings"

	"github.com/caner-cetin/hospital-tracker/internal/models"
	"github.com/caner-cetin/hospital-tracker/internal/services"
	"github.com/gin-gonic/gin"
)

func AuthRequired(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "authorization header required",
			})
			c.Abort()
			return
		}

		bearerToken := strings.Split(authHeader, " ")
		if len(bearerToken) != 2 || bearerToken[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "invalid authorization header format",
			})
			c.Abort()
			return
		}

		claims, err := authService.ValidateToken(bearerToken[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "invalid token",
			})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("hospital_id", claims.HospitalID)
		c.Set("user_type", claims.UserType)
		c.Next()
	}
}

func AuthorizedOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		userType, exists := c.Get("user_type")
		if !exists {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "unauthorized",
			})
			c.Abort()
			return
		}

		if userType != models.UserTypeAuthorized {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error: "access denied: authorized users only",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
