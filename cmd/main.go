package main

import (
	"github.com/nootey/observv/internal/observability"
	"github.com/nootey/observv/pkg/config"
)

const AppVersion = "1.1.1"

func main() {
	cfg := config.LoadConfig()
	cfg.AppVersion = AppVersion // inject

	observability.ExposeAppVersion(AppVersion)
	Execute(cfg)
}
