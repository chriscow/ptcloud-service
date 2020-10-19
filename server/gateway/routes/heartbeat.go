package routes

import (
	"time"

	"github.com/kataras/iris/v12"
)

// Heartbeat returns JSON with the current system status and timestamp
// for monitoring
func Heartbeat(ctx iris.Context) {
	// write out a small JSON response
	ctx.JSON(iris.Map{
		"status":    "OK",
		"timestamp": time.Now().UTC().Unix(),
	})
}
