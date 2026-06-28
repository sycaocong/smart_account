package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"

	"github.com/smartx/account/internal/api"
	"github.com/smartx/account/internal/service"
	"github.com/smartx/account/internal/storage"
)

func main() {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	db := storage.NewMemoryStorage()

	txService := service.NewTransactionService(db)
	balanceService := service.NewBalanceService(db, txService)
	accountService := service.NewAccountService(db, balanceService)
	apiKeyService := service.NewAPIKeyService(db)

	handler := api.NewHandler(accountService, balanceService, txService, apiKeyService, logger)

	go func() {
		if err := handler.Start(":8081"); err != nil && err != http.ErrServerClosed {
			logger.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	logger.Info().Msg("Account service started on :8081")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Info().Msg("Shutting down account service...")



	logger.Info().Msg("Account service shutdown complete")
}