package infrastructure

import (
	"fundingmonitor/internal/domain"
	"github.com/sirupsen/logrus"
)

type XTClient struct {
	config domain.ExchangeConfig
	logger *logrus.Logger
}

func NewXTClient(config domain.ExchangeConfig, logger *logrus.Logger) *XTClient {
	return &XTClient{
		config: config,
		logger: logger,
	}
}

func (x *XTClient) GetName() string {
	return "xt"
}

func (x *XTClient) IsHealthy() bool {
	return true // TODO: implement
}

func (x *XTClient) GetFundingRates() ([]domain.FundingRate, error) {
	return []domain.FundingRate{}, nil // TODO: implement
} 