package integration

import (
	"encoding/json"
	"fundingmonitor/internal/delivery"
	"fundingmonitor/internal/domain"
	"fundingmonitor/internal/infrastructure"
	"fundingmonitor/internal/usecase"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// TestServer represents a test server with all dependencies
type TestServer struct {
	handler    *delivery.FundingHandler
	useCase    *usecase.MultiExchangeUseCase
	exchanges  map[string]domain.ExchangeRepository
	logRepo    domain.LogRepository
	tempDir    string
	server     *httptest.Server
}

// setupTestServer creates a test server with real implementations
func setupTestServer(t *testing.T) *TestServer {
	// Create temporary directory for logs
	tempDir := t.TempDir()
	
	// Create logger
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce noise in tests
	
	// Create mock exchanges
	exchanges := make(map[string]domain.ExchangeRepository)
	
	// Mock Binance
	binanceMock := &MockExchangeRepository{
		name:    "binance",
		healthy: true,
		rates: []domain.FundingRate{
			{Symbol: "BTCUSDT", Exchange: "binance", FundingRate: 0.0001, Timestamp: time.Now()},
			{Symbol: "ETHUSDT", Exchange: "binance", FundingRate: 0.0002, Timestamp: time.Now()},
		},
	}
	exchanges["binance"] = binanceMock
	
	// Mock Bybit
	bybitMock := &MockExchangeRepository{
		name:    "bybit",
		healthy: true,
		rates: []domain.FundingRate{
			{Symbol: "BTCUSDT", Exchange: "bybit", FundingRate: 0.0003, Timestamp: time.Now()},
		},
	}
	exchanges["bybit"] = bybitMock
	
	// Create log repository
	logRepo := infrastructure.NewFileLogger(tempDir, logger)
	
	// Create use case
	useCase := usecase.NewMultiExchangeUseCase(exchanges, logRepo)
	
	// Create handler
	handler := delivery.NewFundingHandler(useCase)
	
	// Create router
	router := mux.NewRouter()
	router.HandleFunc("/api/funding", handler.GetFundingRates).Methods("GET")
	router.HandleFunc("/api/funding/{exchange}", handler.GetExchangeFunding).Methods("GET")
	router.HandleFunc("/api/health", handler.HealthCheck).Methods("GET")
	router.HandleFunc("/api/logs/{symbol}", handler.GetSymbolLogs).Methods("GET")
	router.HandleFunc("/api/logs", handler.GetAllLogs).Methods("GET")
	
	// Create test server
	server := httptest.NewServer(router)
	
	return &TestServer{
		handler:   handler,
		useCase:   useCase,
		exchanges: exchanges,
		logRepo:   logRepo,
		tempDir:   tempDir,
		server:    server,
	}
}

func (ts *TestServer) cleanup() {
	if ts.server != nil {
		ts.server.Close()
	}
}

// MockExchangeRepository for integration tests
type MockExchangeRepository struct {
	name     string
	healthy  bool
	rates    []domain.FundingRate
	err      error
}

func (m *MockExchangeRepository) GetFundingRates() ([]domain.FundingRate, error) {
	return m.rates, m.err
}

func (m *MockExchangeRepository) GetName() string {
	return m.name
}

func (m *MockExchangeRepository) IsHealthy() bool {
	return m.healthy
}

func TestIntegration_GetAllFundingRates(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.cleanup()
	
	// Make HTTP request
	resp, err := http.Get(ts.server.URL + "/api/funding")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()
	
	// Check response
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
	
	// Parse response
	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	
	// Verify response structure
	if rates, ok := response["rates"].([]interface{}); !ok {
		t.Error("Expected 'rates' field in response")
	} else if len(rates) != 3 {
		t.Errorf("Expected 3 rates, got %d", len(rates))
	}
	
	if _, ok := response["timestamp"]; !ok {
		t.Error("Expected 'timestamp' field in response")
	}
}

func TestIntegration_GetExchangeFunding(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.cleanup()
	
	// Test existing exchange
	resp, err := http.Get(ts.server.URL + "/api/funding/binance")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
	
	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	
	if exchange, ok := response["exchange"].(string); !ok || exchange != "binance" {
		t.Errorf("Expected exchange 'binance', got %v", exchange)
	}
	
	// Test non-existing exchange
	resp, err = http.Get(ts.server.URL + "/api/funding/nonexistent")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, resp.StatusCode)
	}
}

func TestIntegration_HealthCheck(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.cleanup()
	
	resp, err := http.Get(ts.server.URL + "/api/health")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
	
	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	
	if status, ok := response["status"].(string); !ok || status != "healthy" {
		t.Errorf("Expected status 'healthy', got %s", status)
	}
	
	if exchanges, ok := response["exchanges"].(float64); !ok || exchanges != 2 {
		t.Errorf("Expected 2 exchanges, got %v", exchanges)
	}
}

func TestIntegration_LoggingFlow(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.cleanup()
	
	// Trigger logging
	err := ts.useCase.LogAllFundingRates()
	if err != nil {
		t.Fatalf("Failed to log funding rates: %v", err)
	}
	
	// Check if log files were created
	date := time.Now().Format("02-01-2006")
	expectedFiles := []string{
		filepath.Join(ts.tempDir, "BTCUSDT", date+".log"),
		filepath.Join(ts.tempDir, "ETHUSDT", date+".log"),
	}
	
	for _, file := range expectedFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Errorf("Expected log file to be created: %s", file)
		}
	}
	
	// Test getting logs via API
	resp, err := http.Get(ts.server.URL + "/api/logs/BTCUSDT?date=" + date)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
	
	// Test getting all logs
	resp, err = http.Get(ts.server.URL + "/api/logs")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
	
	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	
	if count, ok := response["count"].(float64); !ok || count < 1 {
		t.Errorf("Expected at least 1 log file, got %v", count)
	}
}

func TestIntegration_ErrorHandling(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.cleanup()
	
	// Test invalid endpoint
	resp, err := http.Get(ts.server.URL + "/api/invalid")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, resp.StatusCode)
	}
	
	// Test non-existent log file
	resp, err = http.Get(ts.server.URL + "/api/logs/NONEXISTENT?date=01-01-2023")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, resp.StatusCode)
	}
} 