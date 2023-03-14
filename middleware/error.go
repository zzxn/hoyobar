package middleware

import (
	"hoyobar/util/myerr"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		for _, e := range c.Errors {
			err := e.Err
			if myErr, ok := err.(*myerr.MyError); ok {
				log.Printf("error: %v, cause: %v\n", myErr, myErr.Cause())
				c.JSON(http.StatusInternalServerError, gin.H{
					"ecode": myErr.Ecode,
					"emsg":  myErr.Emsg,
				})
			} else {
				log.Println("Unknown error:", e)
				c.JSON(http.StatusInternalServerError, gin.H{
					"ecode": myerr.ErrUnknown.Ecode,
					"emsg":  myerr.ErrUnknown.Emsg,
				})
			}
			return
		}

	}
}
