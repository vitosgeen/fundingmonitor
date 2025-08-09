package usecase

import (
	"fundingmonitor/internal/domain"
)

// MultiExchangeUseCase handles business logic for multiple exchanges
type MultiExchangeUseCase struct {
	exchanges map[string]domain.ExchangeRepository
	logRepo   domain.LogRepository
}

// NewMultiExchangeUseCase creates a new multi-exchange use case
func NewMultiExchangeUseCase(exchanges map[string]domain.ExchangeRepository, logRepo domain.LogRepository) *MultiExchangeUseCase {
	return &MultiExchangeUseCase{
		exchanges: exchanges,
		logRepo:   logRepo,
	}
}

// GetAllFundingRates retrieves funding rates from all exchanges
func (m *MultiExchangeUseCase) GetAllFundingRates() ([]domain.FundingRate, error) {
	var allRates []domain.FundingRate

	for name, exchange := range m.exchanges {
		rates, err := exchange.GetFundingRates()
		if err != nil {
			// Log error but continue with other exchanges
			continue
		}

		// Add exchange name to each rate
		for i := range rates {
			rates[i].Exchange = name
		}

		allRates = append(allRates, rates...)
	}

	return allRates, nil
}

// GetExchangeFundingRates retrieves funding rates from a specific exchange
func (m *MultiExchangeUseCase) GetExchangeFundingRates(exchangeName string) ([]domain.FundingRate, error) {
	exchange, exists := m.exchanges[exchangeName]
	if !exists {
		return nil, domain.ErrExchangeNotFound
	}

	return exchange.GetFundingRates()
}

// GetExchangeInfo returns information about all exchanges
func (m *MultiExchangeUseCase) GetExchangeInfo() map[string]domain.ExchangeInfo {
	info := make(map[string]domain.ExchangeInfo)

	for name, exchange := range m.exchanges {
		info[name] = domain.ExchangeInfo{
			Name:    exchange.GetName(),
			Healthy: exchange.IsHealthy(),
		}
	}

	return info
}

// LogAllFundingRates logs funding rates from all exchanges grouped by symbol
func (m *MultiExchangeUseCase) LogAllFundingRates() error {
	allRates, err := m.GetAllFundingRates()
	if err != nil {
		return err
	}

	// Group rates by symbol
	symbolRates := make(map[string][]domain.FundingRate)
	for _, rate := range allRates {
		symbolRates[rate.Symbol] = append(symbolRates[rate.Symbol], rate)
	}

	// Log each symbol
	for symbol, rates := range symbolRates {
		if err := m.logRepo.LogFundingRates(symbol, rates); err != nil {
			// Log error but continue with other symbols
			continue
		}
	}

	return nil
}

// GetSymbolLogs retrieves logs for a specific symbol
func (m *MultiExchangeUseCase) GetSymbolLogs(symbol string, date string) ([]byte, error) {
	return m.logRepo.GetSymbolLogs(symbol, date)
}

// GetAllLogs retrieves all available logs
func (m *MultiExchangeUseCase) GetAllLogs() ([]domain.LogFile, error) {
	return m.logRepo.GetAllLogs()
}

// GetHistoricalFundingRates retrieves historical funding rates for a symbol and exchange
func (m *MultiExchangeUseCase) GetHistoricalFundingRates(symbol string, exchange string) ([]domain.FundingRateHistory, error) {
	return m.logRepo.GetHistoricalFundingRates(symbol, exchange)
}
