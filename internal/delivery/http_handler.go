package delivery

import (
	"encoding/json"
	"fmt"
	"fundingmonitor/internal/domain"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type FundingHandler struct {
	multiExchangeUseCase domain.MultiExchangeUseCaseInterface
}

func NewFundingHandler(multiExchangeUseCase domain.MultiExchangeUseCaseInterface) *FundingHandler {
	return &FundingHandler{
		multiExchangeUseCase: multiExchangeUseCase,
	}
}

func (h *FundingHandler) GetFundingRates(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	rates, err := h.multiExchangeUseCase.GetAllFundingRates()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get funding rates: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"timestamp": time.Now().Unix(),
		"rates":     rates,
	}

	json.NewEncoder(w).Encode(response)
}

func (h *FundingHandler) GetExchangeFunding(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	exchangeName := vars["exchange"]

	rates, err := h.multiExchangeUseCase.GetExchangeFundingRates(exchangeName)
	if err != nil {
		if err == domain.ErrExchangeNotFound {
			http.Error(w, "Exchange not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to get funding rates: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	
	response := map[string]interface{}{
		"exchange":  exchangeName,
		"timestamp": time.Now().Unix(),
		"rates":     rates,
	}

	json.NewEncoder(w).Encode(response)
}

func (h *FundingHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	exchangeInfo := h.multiExchangeUseCase.GetExchangeInfo()
	
	w.Header().Set("Content-Type", "application/json")
	
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"exchanges": len(exchangeInfo),
		"exchange_info": exchangeInfo,
	}

	json.NewEncoder(w).Encode(response)
}

func (h *FundingHandler) GetSymbolLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	symbol := vars["symbol"]
	
	// Get date from query parameter, default to today
	date := r.URL.Query().Get("date")
	if date == "" {
		date = time.Now().Format("02-01-2006")
	}
	
	content, err := h.multiExchangeUseCase.GetSymbolLogs(symbol, date)
	if err != nil {
		if err == domain.ErrLogFileNotFound {
			http.Error(w, "Log file not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to read log file", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(content)
}

func (h *FundingHandler) GetAllLogs(w http.ResponseWriter, r *http.Request) {
	logFiles, err := h.multiExchangeUseCase.GetAllLogs()
	if err != nil {
		http.Error(w, "Failed to read log directory", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	
	response := map[string]interface{}{
		"log_files": logFiles,
		"count":     len(logFiles),
	}
	
	json.NewEncoder(w).Encode(response)
}

func (h *FundingHandler) FundingWebSocket(w http.ResponseWriter, r *http.Request) {
	// WebSocket implementation for real-time funding rate updates
	// This would require additional implementation
	http.Error(w, "WebSocket not implemented yet", http.StatusNotImplemented)
} 