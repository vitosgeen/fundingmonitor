package infrastructure

import (
	"fundingmonitor/internal/domain"
	"fundingmonitor/internal/usecase"
	"github.com/sirupsen/logrus"
)

// ExchangeFactory creates exchange clients
type ExchangeFactory struct {
	logger *logrus.Logger
}

func NewExchangeFactory(logger *logrus.Logger) *ExchangeFactory {
	return &ExchangeFactory{
		logger: logger,
	}
}

// CreateExchanges creates all enabled exchanges
func (f *ExchangeFactory) CreateExchanges(config *domain.Config) (map[string]domain.ExchangeRepository, error) {
	exchanges := make(map[string]domain.ExchangeRepository)

	for name, exchangeConfig := range config.Exchanges {
		if !exchangeConfig.Enabled {
			continue
		}

		var exchange domain.ExchangeRepository
		switch name {
		case "binance":
			exchange = NewBinanceClient(exchangeConfig, f.logger)
		case "bybit":
			exchange = NewBybitClient(exchangeConfig, f.logger)
		case "okx":
			exchange = NewOKXClient(exchangeConfig, f.logger)
		case "mexc":
			exchange = NewMEXCClient(exchangeConfig, f.logger)
		case "bitget":
			exchange = NewBitgetClient(exchangeConfig, f.logger)
		case "gate":
			exchange = NewGateClient(exchangeConfig, f.logger)
		case "deribit":
			exchange = NewDeribitClient(exchangeConfig, f.logger)
		case "xt":
			exchange = NewXTClient(exchangeConfig, f.logger)
		case "kucoin":
			exchange = NewKuCoinClient(exchangeConfig, f.logger)
		default:
			f.logger.Warnf("Unknown exchange: %s", name)
			continue
		}

		exchanges[name] = exchange
		f.logger.Infof("Initialized exchange: %s", name)
	}

	return exchanges, nil
}

// CreateUseCases creates all use cases
func (f *ExchangeFactory) CreateUseCases(exchanges map[string]domain.ExchangeRepository, logRepo domain.LogRepository) *usecase.MultiExchangeUseCase {
	return usecase.NewMultiExchangeUseCase(exchanges, logRepo)
} 