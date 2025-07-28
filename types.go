package main

import (
	"time"
)

// Exchange interface defines the contract for all exchange implementations
type Exchange interface {
	GetFundingRates() ([]FundingRate, error)
	GetName() string
	IsHealthy() bool
}

// FundingRate represents a funding rate for a specific trading pair
type FundingRate struct {
	Symbol        string    `json:"symbol"`
	Exchange      string    `json:"exchange"`
	FundingRate   float64   `json:"funding_rate"`
	NextFundingTime time.Time `json:"next_funding_time"`
	Timestamp     time.Time `json:"timestamp"`
	MarkPrice     float64   `json:"mark_price,omitempty"`
	IndexPrice    float64   `json:"index_price,omitempty"`
	LastFundingRate float64 `json:"last_funding_rate,omitempty"`
}

// ExchangeConfig holds configuration for each exchange
type ExchangeConfig struct {
	APIKey    string `mapstructure:"api_key"`
	APISecret string `mapstructure:"api_secret"`
	BaseURL   string `mapstructure:"base_url"`
	Enabled   bool   `mapstructure:"enabled"`
}

// Config represents the main application configuration
type Config struct {
	Port            string                    `mapstructure:"port"`
	Exchanges       map[string]ExchangeConfig `mapstructure:"exchanges"`
	LoggingInterval int                       `mapstructure:"logging_interval"` // in minutes
	LogDirectory    string                    `mapstructure:"log_directory"`
}

// FundingRateResponse represents the API response structure
type FundingRateResponse struct {
	Timestamp int64          `json:"timestamp"`
	Rates     []FundingRate  `json:"rates"`
	Total     int            `json:"total"`
}

// ExchangeResponse represents exchange-specific API response
type ExchangeResponse struct {
	Exchange  string         `json:"exchange"`
	Timestamp int64          `json:"timestamp"`
	Rates     []FundingRate  `json:"rates"`
	Status    string         `json:"status"`
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp int64  `json:"timestamp"`
	Exchanges int    `json:"exchanges"`
}

// WebSocketMessage represents real-time funding rate updates
type WebSocketMessage struct {
	Type      string         `json:"type"`
	Exchange  string         `json:"exchange"`
	Timestamp int64          `json:"timestamp"`
	Data      []FundingRate  `json:"data,omitempty"`
	Error     string         `json:"error,omitempty"`
} 