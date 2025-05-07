package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/nootey/observv/internal/repositories"
	"net/http"
)

type UserHandler struct {
	Repo *repositories.UserRepository
}

func NewUserHandler(repo *repositories.UserRepository) *UserHandler {
	return &UserHandler{
		Repo: repo,
	}
}

func (h *UserHandler) GetUsers(c *gin.Context) {
	users, err := h.Repo.FindAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"users": users})
}
