package Infrastructure

import (
	"log"

	"github.com/gin-gonic/gin"
)

// JSONError logs the error and sends a JSON error response with given status.
// If err is non-nil the error's message is used; otherwise msg is used.
func JSONError(ctx *gin.Context, status int, err error, msg string) {
	if err != nil {
		log.Printf("ERROR status=%d method=%s path=%s err=%v", status, ctx.Request.Method, ctx.Request.URL.Path, err)
		ctx.JSON(status, gin.H{"error": err.Error()})
		return
	}
	log.Printf("ERROR status=%d method=%s path=%s message=%s", status, ctx.Request.Method, ctx.Request.URL.Path, msg)
	ctx.JSON(status, gin.H{"error": msg})
	return
}
