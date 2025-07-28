package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func TestHealthCheck(t *testing.T) {
	// Create a test monitor
	logger := logrus.New()
	config := &Config{
		Port:      "8080",
		Exchanges: make(map[string]ExchangeConfig),
	}
	
	monitor := &FundingMonitor{
		exchanges: make(map[string]Exchange),
		logger:    logger,
		config:    config,
	}

	// Create a test request
	req, err := http.NewRequest("GET", "/api/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a response recorder
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(monitor.healthCheck)

	// Call the handler
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body
	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("failed to unmarshal response: %v", err)
	}

	// Check that the response contains expected fields
	if _, ok := response["status"]; !ok {
		t.Error("response missing 'status' field")
	}
	if _, ok := response["timestamp"]; !ok {
		t.Error("response missing 'timestamp' field")
	}
	if _, ok := response["exchanges"]; !ok {
		t.Error("response missing 'exchanges' field")
	}
}

func TestFundingRateStruct(t *testing.T) {
	// Test creating a funding rate
	now := time.Now()
	fundingRate := FundingRate{
		Symbol:           "BTCUSDT",
		Exchange:         "binance",
		FundingRate:      0.0001,
		NextFundingTime:  now.Add(8 * time.Hour),
		Timestamp:        now,
		MarkPrice:        45000.50,
		IndexPrice:       44998.25,
		LastFundingRate:  0.0002,
	}

	// Test JSON marshaling
	data, err := json.Marshal(fundingRate)
	if err != nil {
		t.Errorf("failed to marshal funding rate: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled FundingRate
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Errorf("failed to unmarshal funding rate: %v", err)
	}

	// Verify the data
	if unmarshaled.Symbol != fundingRate.Symbol {
		t.Errorf("symbol mismatch: got %v, want %v", unmarshaled.Symbol, fundingRate.Symbol)
	}
	if unmarshaled.Exchange != fundingRate.Exchange {
		t.Errorf("exchange mismatch: got %v, want %v", unmarshaled.Exchange, fundingRate.Exchange)
	}
	if unmarshaled.FundingRate != fundingRate.FundingRate {
		t.Errorf("funding rate mismatch: got %v, want %v", unmarshaled.FundingRate, fundingRate.FundingRate)
	}
} 