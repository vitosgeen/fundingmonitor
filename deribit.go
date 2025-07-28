package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

type DeribitExchange struct {
	config ExchangeConfig
	logger *logrus.Logger
	client *http.Client
}

type DeribitInstrument struct {
	InstrumentName string `json:"instrument_name"`
	BaseCurrency   string `json:"base_currency"`
	QuoteCurrency  string `json:"quote_currency"`
	IsActive       bool   `json:"is_active"`
}

type DeribitInstrumentsResponse struct {
	JsonRPC string              `json:"jsonrpc"`
	Result  []DeribitInstrument `json:"result"`
}

type DeribitTicker struct {
	InstrumentName string  `json:"instrument_name"`
	CurrentFunding float64 `json:"current_funding"`
	Funding8h      float64 `json:"funding_8h"`
	MarkPrice      float64 `json:"mark_price"`
	IndexPrice     float64 `json:"index_price"`
	Timestamp      int64   `json:"timestamp"`
	State          string  `json:"state"`
}

type DeribitTickerResponse struct {
	JsonRPC string        `json:"jsonrpc"`
	Result  DeribitTicker `json:"result"`
}

func NewDeribitExchange(config ExchangeConfig, logger *logrus.Logger) *DeribitExchange {
	return &DeribitExchange{
		config: config,
		logger: logger,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (d *DeribitExchange) GetName() string {
	return "deribit"
}

func (d *DeribitExchange) IsHealthy() bool {
	url := fmt.Sprintf("%s/api/v2/public/get_instruments?currency=USDC&kind=future&expired=false", d.config.BaseURL)
	resp, err := d.client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func (d *DeribitExchange) GetFundingRates() ([]FundingRate, error) {
	// Get all perpetual instruments for both USDC and BTC
	var allInstruments []DeribitInstrument
	
	// Get USDC perpetual instruments
	usdcURL := fmt.Sprintf("%s/api/v2/public/get_instruments?currency=USDC&kind=future&expired=false", d.config.BaseURL)
	usdcResp, err := d.client.Get(usdcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get USDC instruments: %w", err)
	}
	defer usdcResp.Body.Close()

	if usdcResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(usdcResp.Body)
		return nil, fmt.Errorf("USDC instruments API request failed with status %d: %s", usdcResp.StatusCode, string(body))
	}

	usdcBody, err := io.ReadAll(usdcResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read USDC instruments response: %w", err)
	}

	var usdcResponse DeribitInstrumentsResponse
	if err := json.Unmarshal(usdcBody, &usdcResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal USDC instruments response: %w", err)
	}

	// Get BTC perpetual instruments
	btcURL := fmt.Sprintf("%s/api/v2/public/get_instruments?currency=BTC&kind=future&expired=false", d.config.BaseURL)
	btcResp, err := d.client.Get(btcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get BTC instruments: %w", err)
	}
	defer btcResp.Body.Close()

	if btcResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(btcResp.Body)
		return nil, fmt.Errorf("BTC instruments API request failed with status %d: %s", btcResp.StatusCode, string(body))
	}

	btcBody, err := io.ReadAll(btcResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read BTC instruments response: %w", err)
	}

	var btcResponse DeribitInstrumentsResponse
	if err := json.Unmarshal(btcBody, &btcResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal BTC instruments response: %w", err)
	}

	// Combine all instruments and filter for perpetual contracts
	allInstruments = append(allInstruments, usdcResponse.Result...)
	allInstruments = append(allInstruments, btcResponse.Result...)

	var rates []FundingRate
	
	// Get ticker data for each perpetual instrument
	for _, instrument := range allInstruments {
		// Only process perpetual contracts
		if !instrument.IsActive || !contains(instrument.InstrumentName, "PERPETUAL") {
			continue
		}

		// Get ticker data for this instrument
		tickerURL := fmt.Sprintf("%s/api/v2/public/ticker?instrument_name=%s", d.config.BaseURL, instrument.InstrumentName)
		
		tickerResp, err := d.client.Get(tickerURL)
		if err != nil {
			d.logger.Warnf("Failed to get ticker for %s: %v", instrument.InstrumentName, err)
			continue
		}

		if tickerResp.StatusCode != http.StatusOK {
			tickerResp.Body.Close()
			d.logger.Warnf("Ticker request failed for %s with status %d", instrument.InstrumentName, tickerResp.StatusCode)
			continue
		}

		tickerBody, err := io.ReadAll(tickerResp.Body)
		tickerResp.Body.Close()
		if err != nil {
			d.logger.Warnf("Failed to read ticker response for %s: %v", instrument.InstrumentName, err)
			continue
		}

		var tickerResponse DeribitTickerResponse
		if err := json.Unmarshal(tickerBody, &tickerResponse); err != nil {
			d.logger.Warnf("Failed to unmarshal ticker response for %s: %v", instrument.InstrumentName, err)
			continue
		}

		// Skip if instrument is not active
		if tickerResponse.Result.State != "open" {
			continue
		}

		rates = append(rates, FundingRate{
			Symbol:          tickerResponse.Result.InstrumentName,
			Exchange:        d.GetName(),
			FundingRate:     tickerResponse.Result.CurrentFunding,
			NextFundingTime: time.Now().Add(8 * time.Hour), // Deribit funding occurs every 8 hours
			Timestamp:       time.Unix(tickerResponse.Result.Timestamp/1000, 0),
			MarkPrice:       tickerResponse.Result.MarkPrice,
			IndexPrice:      tickerResponse.Result.IndexPrice,
			LastFundingRate: tickerResponse.Result.Funding8h,
		})
	}

	d.logger.Infof("Retrieved %d funding rates from Deribit", len(rates))
	return rates, nil
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(s) > len(substr) && (s[:len(substr)] == substr || 
		s[len(s)-len(substr):] == substr || 
		containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
} 