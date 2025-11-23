package main

import (
	"flag"

	"github.com/sudo-odner/Backend-trainee-assignment-avito-2025/internal/app"
	"github.com/sudo-odner/Backend-trainee-assignment-avito-2025/internal/config"
	"github.com/sudo-odner/Backend-trainee-assignment-avito-2025/pkg/logger"
)

func main() {
	configPath := flag.String("config", "./config/local.yaml", "path to config file")
	flag.Parse()
	// Init config
	cfg := config.MustLoad(*configPath)
	// Init logger
	log := logger.SetupLogger(cfg.Env)
	// Init microservice
	app.Run(cfg, log)
}
