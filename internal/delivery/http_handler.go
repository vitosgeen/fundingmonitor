package delivery

import (
	"encoding/json"
	"fmt"
	"fundingmonitor/internal/domain"
	"net/http"
	"strconv"
	"strings"
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
		"status":        "healthy",
		"timestamp":     time.Now().Unix(),
		"exchanges":     len(exchangeInfo),
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

	// Parse the log content into structured data
	logEntries := parseLogContent(string(content))

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	response := map[string]interface{}{
		"symbol":     symbol,
		"date":       date,
		"timestamp":  time.Now().Unix(),
		"entries":    logEntries,
		"count":      len(logEntries),
	}

	json.NewEncoder(w).Encode(response)
}

// parseLogContent parses the log content and returns structured data
func parseLogContent(content string) []map[string]interface{} {
	var entries []map[string]interface{}
	lines := strings.Split(content, "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		// Parse log line format: [timestamp] Symbol: symbol, Exchange: exchange, Funding Rate: rate, Mark Price: price, Index Price: price
		if strings.HasPrefix(line, "[") && strings.Contains(line, "] Symbol: ") {
			entry := parseLogLine(line)
			if entry != nil {
				entries = append(entries, entry)
			}
		}
	}
	
	return entries
}

// parseLogLine parses a single log line and returns structured data
func parseLogLine(line string) map[string]interface{} {
	// Extract timestamp
	timestampEnd := strings.Index(line, "]")
	if timestampEnd == -1 {
		return nil
	}
	
	timestampStr := line[1:timestampEnd]
	
	// Extract the rest of the data after the timestamp
	dataPart := line[timestampEnd+2:] // Skip "] "
	
	// Parse the comma-separated fields
	fields := strings.Split(dataPart, ", ")
	entry := map[string]interface{}{
		"timestamp": timestampStr,
	}
	
	for _, field := range fields {
		field = strings.TrimSpace(field)
		if strings.Contains(field, ": ") {
			parts := strings.SplitN(field, ": ", 2)
			if len(parts) == 2 {
				key := parts[0]
				value := parts[1]
				
				// Convert numeric values
				if key == "Funding Rate" || key == "Mark Price" || key == "Index Price" {
					if num, err := strconv.ParseFloat(value, 64); err == nil {
						entry[key] = num
					} else {
						entry[key] = value
					}
				} else {
					entry[key] = value
				}
			}
		}
	}
	
	return entry
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

func (h *FundingHandler) GetHistoricalFundingRates(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	symbol := vars["symbol"]
	exchange := r.URL.Query().Get("exchange")
	if exchange == "" {
		http.Error(w, "Missing exchange parameter", http.StatusBadRequest)
		return
	}
	history, err := h.multiExchangeUseCase.GetHistoricalFundingRates(symbol, exchange)
	if err != nil {
		http.Error(w, "Failed to get historical funding rates", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(history)
}
