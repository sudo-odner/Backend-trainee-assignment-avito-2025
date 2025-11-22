package main

import (
	"github.com/sudo-odner/Backend-trainee-assignment-avito-2025/internal/app"
	"github.com/sudo-odner/Backend-trainee-assignment-avito-2025/internal/config"
	"github.com/sudo-odner/Backend-trainee-assignment-avito-2025/pkg/logger"
)

func main() {
	// Init config
	cfg := config.MustLoad("./config/local.yaml")
	// Init logger
	log := logger.SetupLogger(cfg.Env)
	// Init microservice
	app.Run(cfg, log)
}
