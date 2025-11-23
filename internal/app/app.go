package app

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/sudo-odner/Backend-trainee-assignment-avito-2025/internal/config"
	"github.com/sudo-odner/Backend-trainee-assignment-avito-2025/internal/storage/postgresql"
	"github.com/sudo-odner/Backend-trainee-assignment-avito-2025/internal/transport/router"
	"github.com/sudo-odner/Backend-trainee-assignment-avito-2025/pkg/logger/sl"
)

func Run(cfg *config.Config, logger *slog.Logger) {
	// Init storage
	storage, err := postgresql.New(
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

	// Init transport
	handler := router.New(logger, storage)
	logger.Debug("Router initialized")

	// Run service
	addr := cfg.HttpServer.Host + ":" + cfg.HttpServer.Port
	srv := http.Server{
		Addr:    addr,
		Handler: handler,
	}

	logger.Info(fmt.Sprintf("server listening on '%s'", addr))
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error(fmt.Sprintf("error listening on '%s'", addr), sl.Err(err))
		return
	}
	logger.Info(fmt.Sprintf("server listening on '%s'", addr))
}
