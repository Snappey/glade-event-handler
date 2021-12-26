package sns

import (
	"event-handler/handler"

	"github.com/gin-gonic/gin"
)

func CreateRouter(port int, verify bool, handler *handler.Handler) *gin.Engine {
	router := gin.Default()
	router.POST("/events/sns", event(handler, verify))
	return router
}
