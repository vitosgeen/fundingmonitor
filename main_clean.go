package main

import (
	"context"
	"fundingmonitor/internal/delivery"
	"fundingmonitor/internal/domain"
	"fundingmonitor/internal/infrastructure"
	"fundingmonitor/internal/usecase"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func main() {
	// Initialize logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		DisableColors: true,
	})
	logger.SetLevel(logrus.InfoLevel)

	// Load configuration
	config, err := infrastructure.LoadConfig()
	if err != nil {
		logger.Fatalf("Failed to load config: %v", err)
	}

	// Create log directory
	logDir := config.LogDirectory
	if logDir == "" {
		logDir = "funding_logs"
	}
	if err := os.MkdirAll(logDir, 0755); err != nil {
		logger.Fatalf("Failed to create log directory: %v", err)
	}

	// Initialize infrastructure
	factory := infrastructure.NewExchangeFactory(logger)

	// Create exchanges
	exchanges, err := factory.CreateExchanges(config)
	if err != nil {
		logger.Fatalf("Failed to initialize exchanges: %v", err)
	}

	// Create log repository
	logRepo := infrastructure.NewFileLogger(logDir, logger)

	// Create use cases
	multiExchangeUseCase := factory.CreateUseCases(exchanges, logRepo)

	// Create HTTP handlers
	handler := delivery.NewFundingHandler(multiExchangeUseCase)

	// Start background logging
	go startBackgroundLogging(multiExchangeUseCase, logger, config)

	// Start the server
	server := startServer(handler, config, logger)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Info("Server exited")
}

func startServer(handler *delivery.FundingHandler, config *domain.Config, logger *logrus.Logger) *http.Server {
	router := mux.NewRouter()

	// API routes
	router.HandleFunc("/api/funding", handler.GetFundingRates).Methods("GET")
	router.HandleFunc("/api/funding/{exchange}", handler.GetExchangeFunding).Methods("GET")
	router.HandleFunc("/api/health", handler.HealthCheck).Methods("GET")
	router.HandleFunc("/api/logs/{symbol}", handler.GetSymbolLogs).Methods("GET")
	router.HandleFunc("/api/logs", handler.GetAllLogs).Methods("GET")

	// WebSocket endpoint for real-time updates
	router.HandleFunc("/ws/funding", handler.FundingWebSocket)

	// Static files for web interface
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("static")))

	server := &http.Server{
		Addr:    ":" + config.Port,
		Handler: router,
	}

	go func() {
		logger.Infof("Starting server on port %s", config.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Server error: %v", err)
		}
	}()

	return server
}

func startBackgroundLogging(useCase *usecase.MultiExchangeUseCase, logger *logrus.Logger, config *domain.Config) {
	interval := time.Duration(config.LoggingInterval) * time.Minute
	if interval == 0 {
		interval = 1 * time.Minute // default to 1 minute
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	logger.Infof("Starting background logging every %v", interval)

	for {
		select {
		case <-ticker.C:
			if err := useCase.LogAllFundingRates(); err != nil {
				logger.Errorf("Failed to log funding rates: %v", err)
			}
		}
	}
}
