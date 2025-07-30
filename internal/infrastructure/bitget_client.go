package infrastructure

import (
	"encoding/json"
	"fmt"
	"fundingmonitor/internal/domain"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
)

type BitgetClient struct {
	config domain.ExchangeConfig
	logger *logrus.Logger
	client *http.Client
}

type BitgetTicker struct {
	Symbol             string  `json:"symbol"`
	Last               string  `json:"last"`
	BestAsk            string  `json:"bestAsk"`
	BestBid            string  `json:"bestBid"`
	BidSz              string  `json:"bidSz"`
	AskSz              string  `json:"askSz"`
	High24h            string  `json:"high24h"`
	Low24h             string  `json:"low24h"`
	Timestamp          string  `json:"timestamp"`
	PriceChangePercent string  `json:"priceChangePercent"`
	BaseVolume         string  `json:"baseVolume"`
	QuoteVolume        string  `json:"quoteVolume"`
	UsdtVolume         string  `json:"usdtVolume"`
	OpenUtc            string  `json:"openUtc"`
	ChgUtc             string  `json:"chgUtc"`
	IndexPrice         string  `json:"indexPrice"`
	FundingRate        string  `json:"fundingRate"`
	HoldingAmount      string  `json:"holdingAmount"`
	DeliveryStartTime  *string `json:"deliveryStartTime"`
	DeliveryTime       *string `json:"deliveryTime"`
	DeliveryStatus     string  `json:"deliveryStatus"`
}

type BitgetTickersResponse struct {
	Code        string         `json:"code"`
	Msg         string         `json:"msg"`
	RequestTime int64          `json:"requestTime"`
	Data        []BitgetTicker `json:"data"`
}

func NewBitgetClient(config domain.ExchangeConfig, logger *logrus.Logger) *BitgetClient {
	return &BitgetClient{
		config: config,
		logger: logger,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (b *BitgetClient) GetName() string {
	return "bitget"
}

func (b *BitgetClient) IsHealthy() bool {
	url := fmt.Sprintf("%s/api/mix/v1/market/contracts?productType=umcbl", b.config.BaseURL)
	resp, err := b.client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func (b *BitgetClient) GetFundingRates() ([]domain.FundingRate, error) {
	// Use the bulk tickers endpoint instead of individual calls
	tickersURL := fmt.Sprintf("%s/api/mix/v1/market/tickers?productType=umcbl", b.config.BaseURL)

	req, err := http.NewRequest("GET", tickersURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create tickers request: %w", err)
	}

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make tickers request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("tickers API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read tickers response body: %w", err)
	}

	var tickersResponse BitgetTickersResponse
	if err := json.Unmarshal(body, &tickersResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tickers response: %w", err)
	}

	if tickersResponse.Code != "00000" {
		return nil, fmt.Errorf("Bitget tickers API error: %s", tickersResponse.Msg)
	}

	var rates []domain.FundingRate

	for _, ticker := range tickersResponse.Data {
		// Skip if funding rate is empty or invalid
		if ticker.FundingRate == "" {
			continue
		}

		fundingRate, err := strconv.ParseFloat(ticker.FundingRate, 64)
		if err != nil {
			b.logger.Warnf("Failed to parse funding rate for %s: %v", ticker.Symbol, err)
			continue
		}

		indexPrice, _ := strconv.ParseFloat(ticker.IndexPrice, 64)
		timestamp, _ := strconv.ParseInt(ticker.Timestamp, 10, 64)

		rates = append(rates, domain.FundingRate{
			Symbol:          ticker.Symbol,
			Exchange:        b.GetName(),
			FundingRate:     fundingRate,
			NextFundingTime: time.Now().Add(8 * time.Hour), // Bitget funding occurs every 8 hours
			Timestamp:       time.Unix(timestamp/1000, 0),
			MarkPrice:       0, // Not provided in ticker endpoint
			IndexPrice:      indexPrice,
			LastFundingRate: 0, // Not provided in ticker endpoint
		})
	}

	b.logger.Infof("Retrieved %d funding rates from Bitget", len(rates))
	return rates, nil
}
