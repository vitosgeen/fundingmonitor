package main

import (
	"fundingmonitor/internal/domain"
	"fundingmonitor/internal/infrastructure"
	"log"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

func main() {
	// Create a temporary directory for testing
	tempDir := "test_logs"
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		log.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create logger
	logger := logrus.New()
	fileLogger := infrastructure.NewFileLogger(tempDir, logger)

	// Create test funding rates
	rates := []domain.FundingRate{
		{
			Symbol:      "1INCH_USDT",
			Exchange:    "mexc",
			FundingRate: 0.000072,
			MarkPrice:   0.00,
			IndexPrice:  0.00,
			Timestamp:   time.Now(),
		},
		{
			Symbol:      "1INCH_USDT",
			Exchange:    "gate",
			FundingRate: -0.000308,
			MarkPrice:   0.24,
			IndexPrice:  0.24,
			Timestamp:   time.Now(),
		},
	}

	// Test logging
	if err := fileLogger.LogFundingRates("1INCH_USDT", rates); err != nil {
		log.Fatalf("Failed to log funding rates: %v", err)
	}

	// Read and display the logged content
	date := time.Now().Format("02-01-2006")
	content, err := fileLogger.GetSymbolLogs("1INCH_USDT", date)
	if err != nil {
		log.Fatalf("Failed to read log file: %v", err)
	}

	log.Printf("New log format output:")
	log.Printf("---")
	log.Printf("%s", string(content))
	log.Printf("---")
	log.Printf("Test completed successfully!")
}
