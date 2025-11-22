package app

import (
	"log/slog"

	"github.com/sudo-odner/Backend-trainee-assignment-avito-2025/internal/config"
	"github.com/sudo-odner/Backend-trainee-assignment-avito-2025/internal/storage/postgresql"
	"github.com/sudo-odner/Backend-trainee-assignment-avito-2025/pkg/logger/sl"
)

func Run(cfg *config.Config, logger *slog.Logger) {
	// Init storage
	_, err := postgresql.New(
		cfg.Storage.Host,
		cfg.Storage.Port,
		cfg.Storage.User,
		cfg.Storage.Password,
		cfg.Storage.DBName,
		cfg.Storage.SSLMode)
	if err != nil {
		logger.Error("Storage not initialized", sl.Err(err))
		panic(err)
	}
	logger.Debug("Storage initialized")
	// TODO: Init transport
	// TODO: Run service

	return
}
