package Infrastructure

import (
	"net/http"
	"strings"

	Domain "ShopOps/Domain"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(jwtService JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Bearer token is required"})
			c.Abort()
			return
		}

		token, err := jwtService.ValidateToken(tokenString)
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		userID, err := jwtService.ExtractUserID(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to extract user ID from token"})
			c.Abort()
			return
		}

		phone, err := jwtService.ExtractPhone(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to extract phone from token"})
			c.Abort()
			return
		}

		role, err := jwtService.ExtractRole(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to extract role from token"})
			c.Abort()
			return
		}

		c.Set("userID", userID)
		c.Set("phone", phone)
		c.Set("role", role)
		c.Next()
	}
}

func BusinessMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		businessID := c.Param("businessId")
		if businessID == "" {
			businessID = c.Query("business_id")
		}

		if businessID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Business ID is required"})
			c.Abort()
			return
		}

		c.Set("businessID", businessID)
		c.Next()
	}
}

func OwnerOnlyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "Role not found in context"})
			c.Abort()
			return
		}

		roleStr, ok := role.(string)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": "Invalid role type"})
			c.Abort()
			return
		}

		if roleStr != string(Domain.RoleBusinessOwner) && roleStr != string(Domain.RoleAdmin) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only business owners can perform this action"})
			c.Abort()
			return
		}

		c.Next()
	}
}
