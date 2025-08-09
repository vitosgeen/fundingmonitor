package domain

// MultiExchangeUseCaseInterface defines the contract for multi-exchange use cases
type MultiExchangeUseCaseInterface interface {
	GetAllFundingRates() ([]FundingRate, error)
	GetExchangeFundingRates(exchangeName string) ([]FundingRate, error)
	GetExchangeInfo() map[string]ExchangeInfo
	LogAllFundingRates() error
	GetSymbolLogs(symbol string, date string) ([]byte, error)
	GetAllLogs() ([]LogFile, error)
	GetHistoricalFundingRates(symbol string, exchange string) ([]FundingRateHistory, error)
}
