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
}

// LogFile represents a log file entry
type LogFile struct {
	Symbol   string    `json:"symbol"`
	Date     string    `json:"date"`
	Path     string    `json:"path"`
	Size     int64     `json:"size"`
	Modified time.Time `json:"modified"`
} 