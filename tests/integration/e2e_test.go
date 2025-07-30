package integration

import (
	"context"
	"encoding/json"
	"fundingmonitor/internal/domain"
	"fundingmonitor/internal/infrastructure"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

// TestE2E_RealApplication tests the real application with actual HTTP server
func TestE2E_RealApplication(t *testing.T) {
	// Skip if running in CI or if you want to skip E2E tests
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Create temporary directory
	tempDir := t.TempDir()
	
	// Create logger
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	
	// Create test configuration
	config := &domain.Config{
		Port:         "0", // Use port 0 to get a random available port
		LogDirectory: tempDir,
		Exchanges: map[string]domain.ExchangeConfig{
			"binance": {
				Enabled:  true,
				BaseURL:  "https://api.binance.com",
				APIKey:   "",
				APISecret: "",
			},
			"bybit": {
				Enabled:  true,
				BaseURL:  "https://api.bybit.com",
				APIKey:   "",
				APISecret: "",
			},
		},
		LoggingInterval: 1,
	}

	// Create factory
	factory := infrastructure.NewExchangeFactory(logger)
	
	// Create exchanges (these will be real implementations)
	exchanges, err := factory.CreateExchanges(config)
	if err != nil {
		t.Fatalf("Failed to create exchanges: %v", err)
	}

	// Create log repository
	logRepo := infrastructure.NewFileLogger(tempDir, logger)
	
	// Create use case
	useCase := factory.CreateUseCases(exchanges, logRepo)
	
	// Start background logging in a goroutine
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	go func() {
		ticker := time.NewTicker(2 * time.Second) // Log every 2 seconds for testing
		defer ticker.Stop()
		
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := useCase.LogAllFundingRates(); err != nil {
					logger.Errorf("Failed to log funding rates: %v", err)
				}
			}
		}
	}()

	// Wait a bit for initial logging
	time.Sleep(3 * time.Second)
	
	// Test that log files were created
	logFiles, err := logRepo.GetAllLogs()
	if err != nil {
		t.Fatalf("Failed to get log files: %v", err)
	}
	
	if len(logFiles) == 0 {
		t.Error("Expected log files to be created, but none found")
	}
	
	// Test that we can get funding rates from use case
	rates, err := useCase.GetAllFundingRates()
	if err != nil {
		t.Fatalf("Failed to get funding rates: %v", err)
	}
	
	// Note: In a real E2E test, we might get 0 rates if exchanges are down
	// This is expected behavior, so we just check that the call doesn't panic
	t.Logf("Retrieved %d funding rates from exchanges", len(rates))
	
	// Test exchange info
	exchangeInfo := useCase.GetExchangeInfo()
	if len(exchangeInfo) != 2 {
		t.Errorf("Expected 2 exchanges, got %d", len(exchangeInfo))
	}
	
	for name, info := range exchangeInfo {
		t.Logf("Exchange %s: healthy=%v", name, info.Healthy)
	}
}

// TestE2E_Configuration tests configuration loading
func TestE2E_Configuration(t *testing.T) {
	// Test loading configuration
	config, err := infrastructure.LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	
	// Verify config has expected structure
	if config.Port == "" {
		t.Error("Expected port to be set")
	}
	
	if len(config.Exchanges) == 0 {
		t.Error("Expected exchanges to be configured")
	}
	
	// Check that at least some exchanges are enabled
	enabledCount := 0
	for _, exchange := range config.Exchanges {
		if exchange.Enabled {
			enabledCount++
		}
	}
	
	if enabledCount == 0 {
		t.Error("Expected at least one exchange to be enabled")
	}
}

// TestE2E_FileSystem tests file system operations
func TestE2E_FileSystem(t *testing.T) {
	tempDir := t.TempDir()
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	
	fileLogger := infrastructure.NewFileLogger(tempDir, logger)
	
	// Test creating and reading log files
	rates := []domain.FundingRate{
		{Symbol: "BTCUSDT", Exchange: "binance", FundingRate: 0.0001, Timestamp: time.Now()},
		{Symbol: "ETHUSDT", Exchange: "bybit", FundingRate: 0.0002, Timestamp: time.Now()},
	}
	
	// Create log files
	err := fileLogger.LogFundingRates("BTCUSDT", rates)
	if err != nil {
		t.Fatalf("Failed to log funding rates: %v", err)
	}
	
	err = fileLogger.LogFundingRates("ETHUSDT", rates)
	if err != nil {
		t.Fatalf("Failed to log funding rates: %v", err)
	}
	
	// Verify files were created
	date := time.Now().Format("02-01-2006")
	expectedFiles := []string{
		filepath.Join(tempDir, "BTCUSDT", date+".log"),
		filepath.Join(tempDir, "ETHUSDT", date+".log"),
	}
	
	for _, file := range expectedFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Errorf("Expected log file to be created: %s", file)
		}
	}
	
	// Test reading log files
	logFiles, err := fileLogger.GetAllLogs()
	if err != nil {
		t.Fatalf("Failed to get all logs: %v", err)
	}
	
	if len(logFiles) != 2 {
		t.Errorf("Expected 2 log files, got %d", len(logFiles))
	}
	
	// Test reading specific symbol logs
	content, err := fileLogger.GetSymbolLogs("BTCUSDT", date)
	if err != nil {
		t.Fatalf("Failed to get symbol logs: %v", err)
	}
	
	if len(content) == 0 {
		t.Error("Expected log content to be non-empty")
	}
	
	// Verify JSON structure
	var logEntry struct {
		Timestamp time.Time           `json:"timestamp"`
		Symbol    string              `json:"symbol"`
		Rates     []domain.FundingRate `json:"rates"`
	}
	
	if err := json.Unmarshal(content, &logEntry); err != nil {
		t.Fatalf("Failed to unmarshal log entry: %v", err)
	}
	
	if logEntry.Symbol != "BTCUSDT" {
		t.Errorf("Expected symbol 'BTCUSDT', got %s", logEntry.Symbol)
	}
	
	if len(logEntry.Rates) != 2 {
		t.Errorf("Expected 2 rates, got %d", len(logEntry.Rates))
	}
} 