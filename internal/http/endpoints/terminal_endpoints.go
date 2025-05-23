package endpoints

import (
	"github.com/gin-gonic/gin"
	"github.com/nootey/observv/internal/http/handlers"
)

func TerminalEndpoints(apiGroup *gin.RouterGroup, handler *handlers.TerminalHandler) {
	apiGroup.GET("/", handler.GetTerminals)
}
