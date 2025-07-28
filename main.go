package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
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
	logDir    string
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

	// Create log directory
	logDir := config.LogDirectory
	if logDir == "" {
		logDir = "funding_logs"
	}
	if err := os.MkdirAll(logDir, 0755); err != nil {
		logger.Fatalf("Failed to create log directory: %v", err)
	}

	// Initialize funding monitor
	monitor := &FundingMonitor{
		exchanges: make(map[string]Exchange),
		logger:    logger,
		config:    config,
		logDir:    logDir,
	}

	// Initialize exchanges
	if err := monitor.initializeExchanges(); err != nil {
		logger.Fatalf("Failed to initialize exchanges: %v", err)
	}

	// Start background logging
	go monitor.startBackgroundLogging()

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
		"mexc": map[string]interface{}{
			"enabled":   true,
			"base_url":  "https://api.mexc.com",
			"api_key":   "",
			"api_secret": "",
		},
		"bitget": map[string]interface{}{
			"enabled":   true,
			"base_url":  "https://api.bitget.com",
			"api_key":   "",
			"api_secret": "",
		},
		"gate": map[string]interface{}{
			"enabled":   true,
			"base_url":  "https://api.gateio.ws",
			"api_key":   "",
			"api_secret": "",
		},
		"deribit": map[string]interface{}{
			"enabled":   true,
			"base_url":  "https://www.deribit.com",
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
		case "mexc":
			exchange = NewMEXCExchange(exchangeConfig, fm.logger)
		case "bitget":
			exchange = NewBitgetExchange(exchangeConfig, fm.logger)
		case "gate":
			exchange = NewGateExchange(exchangeConfig, fm.logger)
		case "deribit":
			exchange = NewDeribitExchange(exchangeConfig, fm.logger)
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
	router.HandleFunc("/api/logs/{symbol}", fm.getSymbolLogs).Methods("GET")
	router.HandleFunc("/api/logs", fm.getAllLogs).Methods("GET")

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

func (fm *FundingMonitor) getSymbolLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	symbol := vars["symbol"]
	
	// Get date from query parameter, default to today
	date := r.URL.Query().Get("date")
	if date == "" {
		date = time.Now().Format("02-01-2006")
	} else {
		// Convert from YYYY-MM-DD to DD-MM-YYYY if needed
		if len(date) == 10 && date[4] == '-' && date[7] == '-' {
			parsedDate, err := time.Parse("2006-01-02", date)
			if err == nil {
				date = parsedDate.Format("02-01-2006")
			}
		}
	}
	
	filename := filepath.Join(fm.logDir, symbol, fmt.Sprintf("%s.log", date))
	
	file, err := os.Open(filename)
	if err != nil {
		http.Error(w, "Log file not found", http.StatusNotFound)
		return
	}
	defer file.Close()
	
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	
	// Read and return the file content
	content, err := os.ReadFile(filename)
	if err != nil {
		http.Error(w, "Failed to read log file", http.StatusInternalServerError)
		return
	}
	
	w.Write(content)
}

func (fm *FundingMonitor) getAllLogs(w http.ResponseWriter, r *http.Request) {
	// List all available log files in the new directory structure
	var logFiles []map[string]interface{}
	
	// Walk through all subdirectories
	err := filepath.Walk(fm.logDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Skip the root directory
		if path == fm.logDir {
			return nil
		}
		
		// Only process .log files
		if !info.IsDir() && filepath.Ext(path) == ".log" {
			// Extract symbol and date from path
			relPath, err := filepath.Rel(fm.logDir, path)
			if err != nil {
				return err
			}
			
			// Path format: symbol/date.log
			parts := strings.Split(relPath, string(filepath.Separator))
			if len(parts) == 2 {
				symbol := parts[0]
				date := strings.TrimSuffix(parts[1], ".log")
				
				logFiles = append(logFiles, map[string]interface{}{
					"symbol":      symbol,
					"date":        date,
					"path":        relPath,
					"size":        info.Size(),
					"modified":    info.ModTime(),
				})
			}
		}
		
		return nil
	})
	
	if err != nil {
		http.Error(w, "Failed to read log directory", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"log_files": logFiles,
		"count":     len(logFiles),
	})
} 

// startBackgroundLogging starts a goroutine that logs funding rates to files periodically
func (fm *FundingMonitor) startBackgroundLogging() {
	interval := time.Duration(fm.config.LoggingInterval) * time.Minute
	if interval == 0 {
		interval = 1 * time.Minute // default to 1 minute
	}
	
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	fm.logger.Infof("Starting background logging every %v", interval)

	for {
		select {
		case <-ticker.C:
			fm.logFundingRatesToFiles()
		}
	}
}

// logFundingRatesToFiles logs funding rates for each pair to individual files
func (fm *FundingMonitor) logFundingRatesToFiles() {
	allRates, err := fm.getAllFundingRates()
	if err != nil {
		fm.logger.Errorf("Failed to get funding rates for logging: %v", err)
		return
	}

	// Group rates by symbol
	symbolRates := make(map[string][]FundingRate)
	for _, rate := range allRates {
		symbolRates[rate.Symbol] = append(symbolRates[rate.Symbol], rate)
	}

	// Log each symbol to its own file
	for symbol, rates := range symbolRates {
		fm.logSymbolToFile(symbol, rates)
	}
}

// logSymbolToFile logs funding rates for a specific symbol to a file
func (fm *FundingMonitor) logSymbolToFile(symbol string, rates []FundingRate) {
	// Create directory structure: funding_logs/pair/date.log
	pairDir := filepath.Join(fm.logDir, symbol)
	if err := os.MkdirAll(pairDir, 0755); err != nil {
		fm.logger.Errorf("Failed to create directory for %s: %v", symbol, err)
		return
	}
	
	// Create filename with date format DD-MM-YYYY
	timestamp := time.Now().Format("02-01-2006")
	filename := filepath.Join(pairDir, fmt.Sprintf("%s.log", timestamp))
	
	// Create log entry
	logEntry := struct {
		Timestamp time.Time      `json:"timestamp"`
		Symbol    string         `json:"symbol"`
		Rates     []FundingRate  `json:"rates"`
	}{
		Timestamp: time.Now(),
		Symbol:    symbol,
		Rates:     rates,
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(logEntry, "", "  ")
	if err != nil {
		fm.logger.Errorf("Failed to marshal log entry for %s: %v", symbol, err)
		return
	}

	// Append to file
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fm.logger.Errorf("Failed to open log file for %s: %v", symbol, err)
		return
	}
	defer file.Close()

	// Write with newline
	if _, err := file.Write(append(data, '\n')); err != nil {
		fm.logger.Errorf("Failed to write to log file for %s: %v", symbol, err)
	}
}

// getAllFundingRates gets funding rates from all exchanges
func (fm *FundingMonitor) getAllFundingRates() ([]FundingRate, error) {
	var allRates []FundingRate
	
	for name, exchange := range fm.exchanges {
		rates, err := exchange.GetFundingRates()
		if err != nil {
			fm.logger.Errorf("Failed to get funding rates from %s: %v", name, err)
			continue
		}
		allRates = append(allRates, rates...)
	}
	
	return allRates, nil
} 