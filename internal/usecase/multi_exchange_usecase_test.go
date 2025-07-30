package usecase

import (
	"fundingmonitor/internal/domain"
	"testing"
)

// MockExchangeRepository implements domain.ExchangeRepository for testing
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

// MockLogRepository implements domain.LogRepository for testing
type MockLogRepository struct {
	logErr   error
	getErr   error
	logFiles []domain.LogFile
}

func (m *MockLogRepository) LogFundingRates(symbol string, rates []domain.FundingRate) error {
	return m.logErr
}

func (m *MockLogRepository) GetSymbolLogs(symbol string, date string) ([]byte, error) {
	return []byte("test"), m.getErr
}

func (m *MockLogRepository) GetAllLogs() ([]domain.LogFile, error) {
	return m.logFiles, m.getErr
}

func TestMultiExchangeUseCase_GetAllFundingRates(t *testing.T) {
	// Create mock exchanges
	binanceMock := &MockExchangeRepository{
		name:    "binance",
		healthy: true,
		rates: []domain.FundingRate{
			{Symbol: "BTCUSDT", Exchange: "binance", FundingRate: 0.0001},
			{Symbol: "ETHUSDT", Exchange: "binance", FundingRate: 0.0002},
		},
	}

	bybitMock := &MockExchangeRepository{
		name:    "bybit",
		healthy: true,
		rates: []domain.FundingRate{
			{Symbol: "BTCUSDT", Exchange: "bybit", FundingRate: 0.0003},
		},
	}

	exchanges := map[string]domain.ExchangeRepository{
		"binance": binanceMock,
		"bybit":   bybitMock,
	}

	logRepo := &MockLogRepository{}
	useCase := NewMultiExchangeUseCase(exchanges, logRepo)

	// Test getting all funding rates
	rates, err := useCase.GetAllFundingRates()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should have 3 rates total
	if len(rates) != 3 {
		t.Fatalf("Expected 3 rates, got %d", len(rates))
	}

	// Check that exchange names are set correctly
	for _, rate := range rates {
		if rate.Exchange == "" {
			t.Errorf("Expected exchange name to be set, got empty")
		}
	}
}

func TestMultiExchangeUseCase_GetExchangeFundingRates(t *testing.T) {
	binanceMock := &MockExchangeRepository{
		name:    "binance",
		healthy: true,
		rates: []domain.FundingRate{
			{Symbol: "BTCUSDT", Exchange: "binance", FundingRate: 0.0001},
		},
	}

	exchanges := map[string]domain.ExchangeRepository{
		"binance": binanceMock,
	}

	logRepo := &MockLogRepository{}
	useCase := NewMultiExchangeUseCase(exchanges, logRepo)

	// Test getting rates from existing exchange
	rates, err := useCase.GetExchangeFundingRates("binance")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(rates) != 1 {
		t.Fatalf("Expected 1 rate, got %d", len(rates))
	}

	// Test getting rates from non-existing exchange
	_, err = useCase.GetExchangeFundingRates("nonexistent")
	if err != domain.ErrExchangeNotFound {
		t.Fatalf("Expected ErrExchangeNotFound, got %v", err)
	}
}

func TestMultiExchangeUseCase_GetExchangeInfo(t *testing.T) {
	binanceMock := &MockExchangeRepository{
		name:    "binance",
		healthy: true,
	}

	bybitMock := &MockExchangeRepository{
		name:    "bybit",
		healthy: false,
	}

	exchanges := map[string]domain.ExchangeRepository{
		"binance": binanceMock,
		"bybit":   bybitMock,
	}

	logRepo := &MockLogRepository{}
	useCase := NewMultiExchangeUseCase(exchanges, logRepo)

	info := useCase.GetExchangeInfo()

	if len(info) != 2 {
		t.Fatalf("Expected 2 exchanges, got %d", len(info))
	}

	if info["binance"].Name != "binance" || !info["binance"].Healthy {
		t.Errorf("Expected binance to be healthy, got %+v", info["binance"])
	}

	if info["bybit"].Name != "bybit" || info["bybit"].Healthy {
		t.Errorf("Expected bybit to be unhealthy, got %+v", info["bybit"])
	}
}

func TestMultiExchangeUseCase_LogAllFundingRates(t *testing.T) {
	binanceMock := &MockExchangeRepository{
		name:    "binance",
		healthy: true,
		rates: []domain.FundingRate{
			{Symbol: "BTCUSDT", Exchange: "binance", FundingRate: 0.0001},
			{Symbol: "ETHUSDT", Exchange: "binance", FundingRate: 0.0002},
		},
	}

	bybitMock := &MockExchangeRepository{
		name:    "bybit",
		healthy: true,
		rates: []domain.FundingRate{
			{Symbol: "BTCUSDT", Exchange: "bybit", FundingRate: 0.0003},
		},
	}

	exchanges := map[string]domain.ExchangeRepository{
		"binance": binanceMock,
		"bybit":   bybitMock,
	}

	logRepo := &MockLogRepository{}
	useCase := NewMultiExchangeUseCase(exchanges, logRepo)

	// Test logging all funding rates
	err := useCase.LogAllFundingRates()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
} 