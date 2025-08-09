package delivery

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"fundingmonitor/internal/domain"

	"github.com/gorilla/mux"
)

// MockMultiExchangeUseCase for testing
type MockMultiExchangeUseCase struct {
	rates        []domain.FundingRate
	ratesErr     error
	exchangeInfo map[string]domain.ExchangeInfo
	logFiles     []domain.LogFile
	logErr       error
}

func (m *MockMultiExchangeUseCase) GetAllFundingRates() ([]domain.FundingRate, error) {
	return m.rates, m.ratesErr
}

func (m *MockMultiExchangeUseCase) GetExchangeFundingRates(exchangeName string) ([]domain.FundingRate, error) {
	if exchangeName == "nonexistent" {
		return nil, domain.ErrExchangeNotFound
	}
	if m.ratesErr != nil {
		return nil, m.ratesErr
	}
	return m.rates, nil
}

func (m *MockMultiExchangeUseCase) GetExchangeInfo() map[string]domain.ExchangeInfo {
	return m.exchangeInfo
}

func (m *MockMultiExchangeUseCase) LogAllFundingRates() error {
	return m.logErr
}

func (m *MockMultiExchangeUseCase) GetSymbolLogs(symbol string, date string) ([]byte, error) {
	return []byte("test log content"), m.logErr
}

func (m *MockMultiExchangeUseCase) GetAllLogs() ([]domain.LogFile, error) {
	return m.logFiles, m.logErr
}

func (m *MockMultiExchangeUseCase) GetHistoricalFundingRates(symbol string, exchange string) ([]domain.FundingRateHistory, error) {
	return []domain.FundingRateHistory{}, m.logErr
}

func TestFundingHandler_GetFundingRates(t *testing.T) {
	mockUseCase := &MockMultiExchangeUseCase{
		rates: []domain.FundingRate{
			{Symbol: "BTCUSDT", Exchange: "binance", FundingRate: 0.0001, Timestamp: time.Now()},
			{Symbol: "ETHUSDT", Exchange: "bybit", FundingRate: 0.0002, Timestamp: time.Now()},
		},
	}

	handler := NewFundingHandler(mockUseCase)

	req, err := http.NewRequest("GET", "/api/funding", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler.GetFundingRates(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, status)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}

	if rates, ok := response["rates"].([]interface{}); !ok || len(rates) != 2 {
		t.Errorf("Expected 2 rates, got %d", len(rates))
	}
}

func TestFundingHandler_HealthCheck(t *testing.T) {
	mockUseCase := &MockMultiExchangeUseCase{
		exchangeInfo: map[string]domain.ExchangeInfo{
			"binance": {Name: "binance", Healthy: true},
			"bybit":   {Name: "bybit", Healthy: false},
		},
	}

	handler := NewFundingHandler(mockUseCase)

	req, err := http.NewRequest("GET", "/api/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler.HealthCheck(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, status)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}

	if status, ok := response["status"].(string); !ok || status != "healthy" {
		t.Errorf("Expected status 'healthy', got %s", status)
	}

	if exchanges, ok := response["exchanges"].(float64); !ok || exchanges != 2 {
		t.Errorf("Expected 2 exchanges, got %v", exchanges)
	}
}

func TestFundingHandler_GetExchangeFunding(t *testing.T) {
	mockUseCase := &MockMultiExchangeUseCase{
		rates: []domain.FundingRate{
			{Symbol: "BTCUSDT", Exchange: "binance", FundingRate: 0.0001, Timestamp: time.Now()},
		},
	}

	handler := NewFundingHandler(mockUseCase)

	// Test existing exchange
	req, err := http.NewRequest("GET", "/api/funding/binance", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Set up mux vars for the request
	vars := map[string]string{"exchange": "binance"}
	req = mux.SetURLVars(req, vars)

	rr := httptest.NewRecorder()
	handler.GetExchangeFunding(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, status)
	}

	// Test non-existing exchange
	req, err = http.NewRequest("GET", "/api/funding/nonexistent", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Set up mux vars for the request
	vars = map[string]string{"exchange": "nonexistent"}
	req = mux.SetURLVars(req, vars)

	rr = httptest.NewRecorder()
	handler.GetExchangeFunding(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, status)
	}
}

func TestFundingHandler_GetFundingRatesTop(t *testing.T) {
	mockUseCase := &MockMultiExchangeUseCase{
		rates: []domain.FundingRate{
			{Symbol: "BTCUSDT", Exchange: "binance", FundingRate: 0.005, Timestamp: time.Now()},
			{Symbol: "ETHUSDT", Exchange: "bybit", FundingRate: -0.006, Timestamp: time.Now()},
			{Symbol: "XRPUSDT", Exchange: "okx", FundingRate: 0.002, Timestamp: time.Now()},
		},
	}

	handler := NewFundingHandler(mockUseCase)

	tests := []struct {
		name           string
		topParam       string
		expectedCount  int
		expectedStatus int
	}{
		{
			name:           "Default top param (should use 0.004)",
			topParam:       "",
			expectedCount:  2, // BTCUSDT (0.005), ETHUSDT (-0.006)
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Custom top param 0.002",
			topParam:       "0.002",
			expectedCount:  2, // BTCUSDT (0.005), ETHUSDT (-0.006)
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Custom top param 0.006 (should return none)",
			topParam:       "0.006",
			expectedCount:  0,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid top param (not decimal)",
			topParam:       "1",
			expectedCount:  0,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Percentage top param (0.4%)",
			topParam:       "0.4%",
			expectedCount:  2, // 0.4% = 0.004
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			url := "/api/funding/top"
			if tc.topParam != "" {
				url += "?top=" + tc.topParam
			}
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()
			handler.GetFundingRatesTop(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, rr.Code)
			}

			if tc.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Fatal(err)
				}
				rates, ok := response["rates"].([]interface{})
				if !ok {
					t.Fatalf("Expected rates in response")
				}
				if len(rates) != tc.expectedCount {
					t.Errorf("Expected %d rates, got %d", tc.expectedCount, len(rates))
				}
			}
		})
	}
}

func TestFundingHandler_GetFundingRatesTop_ErrorFromUseCase(t *testing.T) {
	mockUseCase := &MockMultiExchangeUseCase{
		ratesErr: assertAnError(),
	}
	handler := NewFundingHandler(mockUseCase)
	req, _ := http.NewRequest("GET", "/api/funding/top", nil)
	rr := httptest.NewRecorder()
	handler.GetFundingRatesTop(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}
}

// Helper to simulate an error
func assertAnError() error {
	return fmt.Errorf("mock error")
}
