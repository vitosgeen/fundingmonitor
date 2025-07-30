package infrastructure

import (
	"fundingmonitor/internal/domain"
	"os"
	"path/filepath"
	"strings"
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

	contentStr := string(content)

	// Check that the content contains the expected text format
	if !strings.Contains(contentStr, "[") || !strings.Contains(contentStr, "] Symbol: BTCUSDT") {
		t.Errorf("Expected text format with timestamp and symbol, got: %s", contentStr)
	}

	// Check that we have the expected number of rate lines
	rateLines := strings.Count(contentStr, "Exchange:")
	if rateLines != 2 {
		t.Errorf("Expected 2 rate lines, got %d", rateLines)
	}

	// Check that both exchanges are present
	if !strings.Contains(contentStr, "Exchange: binance") {
		t.Errorf("Expected binance exchange in log")
	}
	if !strings.Contains(contentStr, "Exchange: bybit") {
		t.Errorf("Expected bybit exchange in log")
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

	testContent := `[2024-01-01 00:00:00] Symbol: BTCUSDT
  Exchange: binance, Funding Rate: 0.000100, Mark Price: 50000.00, Index Price: 50000.00

`
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
		testContent := `[2024-01-01 00:00:00] Symbol: ` + symbol + `
  Exchange: binance, Funding Rate: 0.000100, Mark Price: 50000.00, Index Price: 50000.00

`
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
