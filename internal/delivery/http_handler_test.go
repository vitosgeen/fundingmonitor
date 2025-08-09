package delivery

import (
	"encoding/json"
	"fundingmonitor/internal/domain"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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
