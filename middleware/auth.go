package middleware

import (
	"log"

	"github.com/gin-gonic/gin"
)

func ReadAuthToken(callback func(authToken string, c *gin.Context)) gin.HandlerFunc {
	return func(c *gin.Context) {
		authToken := c.GetHeader("Auth")
		log.Println("Got auth:", authToken)
		if authToken != "" {
			callback(authToken, c)
		}
		c.Next()
	}
}
