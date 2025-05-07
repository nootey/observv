package bootstrap

import "github.com/nootey/observv/pkg/config"

type Container struct {
	Config *config.Config
}

func NewContainer(cfg *config.Config) *Container {
	return &Container{
		Config: cfg,
	}
}
