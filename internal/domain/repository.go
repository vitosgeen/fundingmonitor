package domain

import "time"

// ExchangeRepository defines the contract for exchange data access
type ExchangeRepository interface {
	GetFundingRates() ([]FundingRate, error)
	GetName() string
	IsHealthy() bool
}

// LogRepository defines the contract for logging operations
type LogRepository interface {
	LogFundingRates(symbol string, rates []FundingRate) error
	GetSymbolLogs(symbol string, date string) ([]byte, error)
	GetAllLogs() ([]LogFile, error)
	GetHistoricalFundingRates(symbol string, exchange string) ([]FundingRateHistory, error)
}

// LogFile represents a log file entry
type LogFile struct {
	Symbol   string    `json:"symbol"`
	Date     string    `json:"date"`
	Path     string    `json:"path"`
	Size     int64     `json:"size"`
	Modified time.Time `json:"modified"`
}

// FundingRateHistory represents a funding rate at a specific time for a symbol and exchange
type FundingRateHistory struct {
	Timestamp   int64   `json:"timestamp"`
	FundingRate float64 `json:"funding_rate"`
}
