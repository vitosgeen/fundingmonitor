package usecase

import (
	"fundingmonitor/internal/domain"
)

// FundingUseCase handles business logic for funding rate operations
type FundingUseCase struct {
	exchangeRepo domain.ExchangeRepository
	logRepo      domain.LogRepository
}

// NewFundingUseCase creates a new funding use case
func NewFundingUseCase(exchangeRepo domain.ExchangeRepository, logRepo domain.LogRepository) *FundingUseCase {
	return &FundingUseCase{
		exchangeRepo: exchangeRepo,
		logRepo:      logRepo,
	}
}

// GetFundingRates retrieves funding rates from the exchange
func (f *FundingUseCase) GetFundingRates() ([]domain.FundingRate, error) {
	return f.exchangeRepo.GetFundingRates()
}

// GetExchangeInfo returns exchange information
func (f *FundingUseCase) GetExchangeInfo() domain.ExchangeInfo {
	return domain.ExchangeInfo{
		Name:    f.exchangeRepo.GetName(),
		Healthy: f.exchangeRepo.IsHealthy(),
	}
}

// LogFundingRates logs funding rates for a symbol
func (f *FundingUseCase) LogFundingRates(symbol string, rates []domain.FundingRate) error {
	return f.logRepo.LogFundingRates(symbol, rates)
}

// GetSymbolLogs retrieves logs for a specific symbol
func (f *FundingUseCase) GetSymbolLogs(symbol string, date string) ([]byte, error) {
	return f.logRepo.GetSymbolLogs(symbol, date)
}

// GetAllLogs retrieves all available logs
func (f *FundingUseCase) GetAllLogs() ([]domain.LogFile, error) {
	return f.logRepo.GetAllLogs()
} 