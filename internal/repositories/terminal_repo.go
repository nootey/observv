package repositories

import (
	"github.com/nootey/observv/internal/models"
	"gorm.io/gorm"
)

type TerminalRepository struct {
	DB *gorm.DB
}

func NewTerminalRepository(db *gorm.DB) *TerminalRepository {
	return &TerminalRepository{DB: db}
}

func (r *TerminalRepository) FindAllTerminals() ([]models.Terminal, error) {
	var Terminals []models.Terminal

	err := r.DB.
		Find(&Terminals).Error
	if err != nil {
		return nil, err
	}
	return Terminals, nil
}
