package infrastructure

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"fundingmonitor/internal/domain"

	"github.com/sirupsen/logrus"
)

type ElasticsearchLogger struct {
	client    *http.Client
	baseURL   string
	logger    *logrus.Logger
	indexName string
}

type FundingRateDocument struct {
	Symbol      string    `json:"symbol"`
	Exchange    string    `json:"exchange"`
	FundingRate float64   `json:"funding_rate"`
	MarkPrice   float64   `json:"mark_price"`
	IndexPrice  float64   `json:"index_price"`
	Timestamp   time.Time `json:"timestamp"`
	DataType    string    `json:"data_type"`
}

func NewElasticsearchLogger(baseURL string, logger *logrus.Logger) *ElasticsearchLogger {
	return &ElasticsearchLogger{
		client:    &http.Client{Timeout: 10 * time.Second},
		baseURL:   baseURL,
		logger:    logger,
		indexName: "funding-monitor",
	}
}

func (e *ElasticsearchLogger) LogFundingRates(symbol string, rates []domain.FundingRate) error {
	if len(rates) == 0 {
		return nil
	}

	// Create bulk request
	var bulkBody bytes.Buffer

	for _, rate := range rates {
		// Index action
		indexAction := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": fmt.Sprintf("%s-%s", e.indexName, time.Now().Format("2006.01.02")),
			},
		}
		indexJSON, _ := json.Marshal(indexAction)
		bulkBody.Write(indexJSON)
		bulkBody.WriteString("\n")

		// Document
		doc := FundingRateDocument{
			Symbol:      symbol,
			Exchange:    rate.Exchange,
			FundingRate: rate.FundingRate,
			MarkPrice:   rate.MarkPrice,
			IndexPrice:  rate.IndexPrice,
			Timestamp:   rate.Timestamp,
			DataType:    "funding_rate",
		}
		docJSON, _ := json.Marshal(doc)
		bulkBody.Write(docJSON)
		bulkBody.WriteString("\n")
	}

	// Send bulk request
	url := fmt.Sprintf("%s/_bulk", e.baseURL)
	resp, err := e.client.Post(url, "application/json", &bulkBody)
	if err != nil {
		return fmt.Errorf("failed to send bulk request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("elasticsearch bulk request failed with status: %d", resp.StatusCode)
	}

	e.logger.Infof("Successfully logged %d funding rates for %s to Elasticsearch", len(rates), symbol)
	return nil
}

func (e *ElasticsearchLogger) GetSymbolLogs(symbol string, date string) ([]byte, error) {
	// Query Elasticsearch for symbol logs
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					{"term": map[string]interface{}{"symbol": symbol}},
					{"term": map[string]interface{}{"data_type": "funding_rate"}},
				},
			},
		},
		"sort": []map[string]interface{}{
			{"timestamp": map[string]interface{}{"order": "desc"}},
		},
		"size": 1000,
	}

	queryJSON, _ := json.Marshal(query)
	url := fmt.Sprintf("%s/%s-%s/_search", e.baseURL, e.indexName, date)

	resp, err := e.client.Post(url, "application/json", bytes.NewBuffer(queryJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to query elasticsearch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("elasticsearch query failed with status: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode elasticsearch response: %w", err)
	}

	return json.MarshalIndent(result, "", "  ")
}

func (e *ElasticsearchLogger) GetAllLogs() ([]domain.LogFile, error) {
	// For Elasticsearch, we return a single log file representing the index
	today := time.Now().Format("2006.01.02")
	return []domain.LogFile{
		{
			Symbol:   "elasticsearch",
			Date:     today,
			Path:     fmt.Sprintf("%s-%s", e.indexName, today),
			Size:     0, // Size not applicable for Elasticsearch
			Modified: time.Now(),
		},
	}, nil
}

func (e *ElasticsearchLogger) GetHistoricalFundingRates(symbol string, exchange string) ([]domain.FundingRateHistory, error) {
	// Not implemented for ElasticsearchLogger
	return []domain.FundingRateHistory{}, nil
}
