package bootstrap

import (
	"github.com/nootey/observv/internal/repositories"
	"github.com/nootey/observv/pkg/config"
	"gorm.io/gorm"
)

type Container struct {
	Config   *config.Config
	DB       *gorm.DB
	UserRepo *repositories.UserRepository
}

func NewContainer(cfg *config.Config, db *gorm.DB) *Container {

	userRepo := repositories.NewUserRepository(db)

	return &Container{
		Config:   cfg,
		DB:       db,
		UserRepo: userRepo,
	}
}
