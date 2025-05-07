package bootstrap

import "github.com/nootey/observv/pkg/config"
	"github.com/nootey/observv/pkg/config"
	"gorm.io/gorm"

type Container struct {
	Config   *config.Config
	DB       *gorm.DB
}

func NewContainer(cfg *config.Config) *Container {
	return &Container{
		Config:   cfg,
		DB:       db,
	}
}
