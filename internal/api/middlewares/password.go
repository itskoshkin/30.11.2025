package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	"link-availability-checker/internal/config"
)

func AskPassword() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Password")
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"Error": "Password required"})
			return
		}

		if header != viper.GetString(config.ApiPassword) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"Error": "Invalid password"})
			return
		}

		c.Next()
	}
}
