package http

import (
	"github.com/gin-gonic/gin"
	"github.com/nootey/observv/internal/bootstrap"
	"github.com/nootey/observv/internal/http/endpoints"
	"github.com/nootey/observv/internal/http/handlers"
	"net/http"
)

type RouteInitializerHTTP struct {
	Router    *gin.Engine
	Container *bootstrap.Container
}

func NewRouteInitializerHTTP(router *gin.Engine, container *bootstrap.Container) *RouteInitializerHTTP {
	return &RouteInitializerHTTP{
		Router:    router,
		Container: container,
	}
}

func rootHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "HTTP server is running!"})
}

func (r *RouteInitializerHTTP) InitEndpoints() {
	api := r.Router.Group("/api")

	api.GET("/", rootHandler)
	r.initRoutes(api)
}

func (r *RouteInitializerHTTP) initRoutes(api *gin.RouterGroup) {
	userHandler := handlers.NewUserHandler(r.Container.UserRepo)

	// Public routes
	publicGroup := api.Group("")
	{
		publicUserRoutes := publicGroup.Group("/users")
		endpoints.UserEndpoints(publicUserRoutes, userHandler)
	}
}
