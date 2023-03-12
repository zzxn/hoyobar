package middleware

import (
	"github.com/gin-gonic/gin"
)

func Auth() gin.HandlerFunc {
   return func(c *gin.Context) {
       authToken := c.GetHeader("Auth")
       if authToken == "" {
           return
       }
       c.Set("auth_token", authToken)
   }
}
