package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type FundingMonitor struct {
	exchanges map[string]Exchange
	logger    *logrus.Logger
	config    *Config
}



func main() {
	// Initialize logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	// Load configuration
	config, err := loadConfig()
	if err != nil {
		logger.Fatalf("Failed to load config: %v", err)
	}

	// Initialize funding monitor
	monitor := &FundingMonitor{
		exchanges: make(map[string]Exchange),
		logger:    logger,
		config:    config,
	}

	// Initialize exchanges
	if err := monitor.initializeExchanges(); err != nil {
		logger.Fatalf("Failed to initialize exchanges: %v", err)
	}

	// Start the server
	server := monitor.startServer()

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

func loadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// Set defaults
	viper.SetDefault("port", "8080")
	viper.SetDefault("exchanges", map[string]interface{}{
		"binance": map[string]interface{}{
			"enabled":   true,
			"base_url":  "https://api.binance.com",
			"api_key":   "",
			"api_secret": "",
		},
		"bybit": map[string]interface{}{
			"enabled":   true,
			"base_url":  "https://api.bybit.com",
			"api_key":   "",
			"api_secret": "",
		},
		"okx": map[string]interface{}{
			"enabled":   true,
			"base_url":  "https://www.okx.com",
			"api_key":   "",
			"api_secret": "",
		},
	})

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func (fm *FundingMonitor) initializeExchanges() error {
	for name, exchangeConfig := range fm.config.Exchanges {
		if !exchangeConfig.Enabled {
			continue
		}

		var exchange Exchange
		switch name {
		case "binance":
			exchange = NewBinanceExchange(exchangeConfig, fm.logger)
		case "bybit":
			exchange = NewBybitExchange(exchangeConfig, fm.logger)
		case "okx":
			exchange = NewOKXExchange(exchangeConfig, fm.logger)
		default:
			fm.logger.Warnf("Unknown exchange: %s", name)
			continue
		}

		fm.exchanges[name] = exchange
		fm.logger.Infof("Initialized exchange: %s", name)
	}

	return nil
}

func (fm *FundingMonitor) startServer() *http.Server {
	router := mux.NewRouter()

	// API routes
	router.HandleFunc("/api/funding", fm.getFundingRates).Methods("GET")
	router.HandleFunc("/api/funding/{exchange}", fm.getExchangeFunding).Methods("GET")
	router.HandleFunc("/api/health", fm.healthCheck).Methods("GET")

	// WebSocket endpoint for real-time updates
	router.HandleFunc("/ws/funding", fm.fundingWebSocket)

	// Static files for web interface
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("static")))

	server := &http.Server{
		Addr:    ":" + fm.config.Port,
		Handler: router,
	}

	go func() {
		fm.logger.Infof("Starting server on port %s", fm.config.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fm.logger.Fatalf("Server error: %v", err)
		}
	}()

	return server
}

func (fm *FundingMonitor) getFundingRates(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var allRates []FundingRate
	for name, exchange := range fm.exchanges {
		rates, err := exchange.GetFundingRates()
		if err != nil {
			fm.logger.Errorf("Failed to get funding rates from %s: %v", name, err)
			continue
		}

		for _, rate := range rates {
			rate.Exchange = name
			allRates = append(allRates, rate)
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"timestamp": time.Now().Unix(),
		"rates":     allRates,
	})
}

func (fm *FundingMonitor) getExchangeFunding(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	exchangeName := vars["exchange"]

	exchange, exists := fm.exchanges[exchangeName]
	if !exists {
		http.Error(w, "Exchange not found", http.StatusNotFound)
		return
	}

	rates, err := exchange.GetFundingRates()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get funding rates: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"exchange":  exchangeName,
		"timestamp": time.Now().Unix(),
		"rates":     rates,
	})
}

func (fm *FundingMonitor) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"exchanges": len(fm.exchanges),
	})
}

func (fm *FundingMonitor) fundingWebSocket(w http.ResponseWriter, r *http.Request) {
	// WebSocket implementation for real-time funding rate updates
	// This would require additional implementation
	http.Error(w, "WebSocket not implemented yet", http.StatusNotImplemented)
} 