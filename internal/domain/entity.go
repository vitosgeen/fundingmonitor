package domain

import (
	"time"
)

// FundingRate represents a funding rate for a specific trading pair
type FundingRate struct {
	Symbol           string    `json:"symbol"`
	Exchange         string    `json:"exchange"`
	FundingRate      float64   `json:"funding_rate"`
	NextFundingTime  time.Time `json:"next_funding_time"`
	Timestamp        time.Time `json:"timestamp"`
	MarkPrice        float64   `json:"mark_price,omitempty"`
	IndexPrice       float64   `json:"index_price,omitempty"`
	LastFundingRate  float64   `json:"last_funding_rate,omitempty"`
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

// ExchangeInfo represents exchange status information
type ExchangeInfo struct {
	Name    string `json:"name"`
	Healthy bool   `json:"healthy"`
} 