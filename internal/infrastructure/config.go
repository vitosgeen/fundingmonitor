package infrastructure

import (
	"fundingmonitor/internal/domain"
	"github.com/spf13/viper"
)

func LoadConfig() (*domain.Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// Set defaults
	viper.SetDefault("port", "8080")
	viper.SetDefault("exchanges", map[string]interface{}{
		"binance": map[string]interface{}{
			"enabled":   true,
			"base_url":  "https://api.binance.com",
			"api_key":   "",
			"api_secret": "",
		},
		"bybit": map[string]interface{}{
			"enabled":   true,
			"base_url":  "https://api.bybit.com",
			"api_key":   "",
			"api_secret": "",
		},
		"okx": map[string]interface{}{
			"enabled":   true,
			"base_url":  "https://www.okx.com",
			"api_key":   "",
			"api_secret": "",
		},
		"mexc": map[string]interface{}{
			"enabled":   true,
			"base_url":  "https://api.mexc.com",
			"api_key":   "",
			"api_secret": "",
		},
		"bitget": map[string]interface{}{
			"enabled":   true,
			"base_url":  "https://api.bitget.com",
			"api_key":   "",
			"api_secret": "",
		},
		"gate": map[string]interface{}{
			"enabled":   true,
			"base_url":  "https://api.gateio.ws",
			"api_key":   "",
			"api_secret": "",
		},
		"deribit": map[string]interface{}{
			"enabled":   true,
			"base_url":  "https://www.deribit.com",
			"api_key":   "",
			"api_secret": "",
		},
		"xt": map[string]interface{}{
			"enabled":   true,
			"base_url":  "https://api.xt.com",
			"api_key":   "",
			"api_secret": "",
		},
		"kucoin": map[string]interface{}{
			"enabled":   true,
			"base_url":  "https://api-futures.kucoin.com",
			"api_key":   "",
			"api_secret": "",
		},
	})

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var config domain.Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
} 