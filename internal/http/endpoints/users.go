package endpoints

import (
	"github.com/gin-gonic/gin"
	"github.com/nootey/observv/internal/http/handlers"
)

func UserEndpoints(apiGroup *gin.RouterGroup, handler *handlers.UserHandler) {
	apiGroup.GET("/", handler.GetUsers)
}
