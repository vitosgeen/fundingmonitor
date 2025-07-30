package infrastructure

import (
	"encoding/json"
	"fundingmonitor/internal/domain"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func TestFileLogger_LogFundingRates(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()
	logger := logrus.New()
	fileLogger := NewFileLogger(tempDir, logger)

	rates := []domain.FundingRate{
		{Symbol: "BTCUSDT", Exchange: "binance", FundingRate: 0.0001, Timestamp: time.Now()},
		{Symbol: "ETHUSDT", Exchange: "bybit", FundingRate: 0.0002, Timestamp: time.Now()},
	}

	// Test logging funding rates
	err := fileLogger.LogFundingRates("BTCUSDT", rates)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check if file was created
	date := time.Now().Format("02-01-2006")
	expectedPath := filepath.Join(tempDir, "BTCUSDT", date+".log")
	
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Fatalf("Expected log file to be created at %s", expectedPath)
	}

	// Read and verify the content
	content, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

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

func TestFileLogger_GetSymbolLogs(t *testing.T) {
	tempDir := t.TempDir()
	logger := logrus.New()
	fileLogger := NewFileLogger(tempDir, logger)

	// Create a test log file
	symbolDir := filepath.Join(tempDir, "BTCUSDT")
	if err := os.MkdirAll(symbolDir, 0755); err != nil {
		t.Fatal(err)
	}

	date := time.Now().Format("02-01-2006")
	logFile := filepath.Join(symbolDir, date+".log")
	
	testContent := `{"timestamp":"2024-01-01T00:00:00Z","symbol":"BTCUSDT","rates":[]}`
	if err := os.WriteFile(logFile, []byte(testContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Test getting symbol logs
	content, err := fileLogger.GetSymbolLogs("BTCUSDT", date)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if string(content) != testContent {
		t.Errorf("Expected content %s, got %s", testContent, string(content))
	}

	// Test getting non-existent log
	_, err = fileLogger.GetSymbolLogs("BTCUSDT", "01-01-2023")
	if err != domain.ErrLogFileNotFound {
		t.Errorf("Expected ErrLogFileNotFound, got %v", err)
	}
}

func TestFileLogger_GetAllLogs(t *testing.T) {
	tempDir := t.TempDir()
	logger := logrus.New()
	fileLogger := NewFileLogger(tempDir, logger)

	// Create test log files
	symbols := []string{"BTCUSDT", "ETHUSDT"}
	date := time.Now().Format("02-01-2006")

	for _, symbol := range symbols {
		symbolDir := filepath.Join(tempDir, symbol)
		if err := os.MkdirAll(symbolDir, 0755); err != nil {
			t.Fatal(err)
		}

		logFile := filepath.Join(symbolDir, date+".log")
		testContent := `{"timestamp":"2024-01-01T00:00:00Z","symbol":"` + symbol + `","rates":[]}`
		if err := os.WriteFile(logFile, []byte(testContent), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Test getting all logs
	logFiles, err := fileLogger.GetAllLogs()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(logFiles) != 2 {
		t.Errorf("Expected 2 log files, got %d", len(logFiles))
	}

	// Check that we have logs for both symbols
	symbolsFound := make(map[string]bool)
	for _, logFile := range logFiles {
		symbolsFound[logFile.Symbol] = true
	}

	for _, symbol := range symbols {
		if !symbolsFound[symbol] {
			t.Errorf("Expected to find log for symbol %s", symbol)
		}
	}
} 