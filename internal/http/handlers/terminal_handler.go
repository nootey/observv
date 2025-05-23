package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/nootey/observv/internal/repositories"
	"net/http"
)

type TerminalHandler struct {
	Repo *repositories.TerminalRepository
}

func NewTerminalHandler(repo *repositories.TerminalRepository) *TerminalHandler {
	return &TerminalHandler{
		Repo: repo,
	}
}

func (h *TerminalHandler) GetTerminals(c *gin.Context) {
	Terminals, err := h.Repo.FindAllTerminals()
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"Terminals": Terminals})
}
