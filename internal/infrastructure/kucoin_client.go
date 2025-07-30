package infrastructure

import (
	"encoding/json"
	"fmt"
	"fundingmonitor/internal/domain"
	"io"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

type KuCoinClient struct {
	config domain.ExchangeConfig
	logger *logrus.Logger
	client *http.Client
}

type KuCoinContract struct {
	Symbol                    string  `json:"symbol"`
	MarkPrice                 float64 `json:"markPrice"`
	IndexPrice                float64 `json:"indexPrice"`
	FundingFeeRate            float64 `json:"fundingFeeRate"`
	NextFundingRateDateTime   int64   `json:"nextFundingRateDateTime"`
	Status                    string  `json:"status"`
}

type KuCoinContractsResponse struct {
	Code string            `json:"code"`
	Data []KuCoinContract `json:"data"`
}

func NewKuCoinClient(config domain.ExchangeConfig, logger *logrus.Logger) *KuCoinClient {
	return &KuCoinClient{
		config: config,
		logger: logger,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (k *KuCoinClient) GetName() string {
	return "kucoin"
}

func (k *KuCoinClient) IsHealthy() bool {
	url := fmt.Sprintf("%s/api/v1/contracts/active", k.config.BaseURL)
	resp, err := k.client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func (k *KuCoinClient) GetFundingRates() ([]domain.FundingRate, error) {
	url := fmt.Sprintf("%s/api/v1/contracts/active", k.config.BaseURL)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := k.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var kucoinResponse KuCoinContractsResponse
	if err := json.Unmarshal(body, &kucoinResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if kucoinResponse.Code != "200000" {
		return nil, fmt.Errorf("KuCoin API error: code %s", kucoinResponse.Code)
	}

	var rates []domain.FundingRate
	for _, contract := range kucoinResponse.Data {
		if contract.Status != "Open" {
			continue
		}

		rates = append(rates, domain.FundingRate{
			Symbol:          contract.Symbol,
			Exchange:        k.GetName(),
			FundingRate:     contract.FundingFeeRate,
			NextFundingTime: time.Unix(contract.NextFundingRateDateTime/1000, 0),
			Timestamp:       time.Now(),
			MarkPrice:       contract.MarkPrice,
			IndexPrice:      contract.IndexPrice,
			LastFundingRate: 0,
		})
	}

	k.logger.Infof("Retrieved %d funding rates from KuCoin", len(rates))
	return rates, nil
} 