package repositories

import (
	"github.com/nootey/observv/internal/models"
	"gorm.io/gorm"
)

type UserRepository struct {
	DB *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{DB: db}
}

func (r *UserRepository) FindAllUsers() ([]models.User, error) {
	var users []models.User

	err := r.DB.
		Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}
