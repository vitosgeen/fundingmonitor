package infrastructure

import (
	"encoding/json"
	"fmt"
	"fundingmonitor/internal/domain"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type FileLogger struct {
	logDir string
	logger *logrus.Logger
}

func NewFileLogger(logDir string, logger *logrus.Logger) *FileLogger {
	return &FileLogger{
		logDir: logDir,
		logger: logger,
	}
}

func (f *FileLogger) LogFundingRates(symbol string, rates []domain.FundingRate) error {
	// Create directory structure: funding_logs/symbol/date.log
	pairDir := filepath.Join(f.logDir, symbol)
	if err := os.MkdirAll(pairDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory for %s: %w", symbol, err)
	}
	
	// Create filename with date format DD-MM-YYYY
	timestamp := time.Now().Format("02-01-2006")
	filename := filepath.Join(pairDir, fmt.Sprintf("%s.log", timestamp))
	
	// Create log entry
	logEntry := struct {
		Timestamp time.Time           `json:"timestamp"`
		Symbol    string              `json:"symbol"`
		Rates     []domain.FundingRate `json:"rates"`
	}{
		Timestamp: time.Now(),
		Symbol:    symbol,
		Rates:     rates,
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(logEntry, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal log entry for %s: %w", symbol, err)
	}

	// Append to file
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file for %s: %w", symbol, err)
	}
	defer file.Close()

	// Write with newline
	if _, err := file.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to write to log file for %s: %w", symbol, err)
	}

	return nil
}

func (f *FileLogger) GetSymbolLogs(symbol string, date string) ([]byte, error) {
	// Convert from YYYY-MM-DD to DD-MM-YYYY if needed
	if len(date) == 10 && date[4] == '-' && date[7] == '-' {
		parsedDate, err := time.Parse("2006-01-02", date)
		if err == nil {
			date = parsedDate.Format("02-01-2006")
		}
	}
	
	filename := filepath.Join(f.logDir, symbol, fmt.Sprintf("%s.log", date))
	
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, domain.ErrLogFileNotFound
	}
	
	return content, nil
}

func (f *FileLogger) GetAllLogs() ([]domain.LogFile, error) {
	var logFiles []domain.LogFile
	
	// Walk through all subdirectories
	err := filepath.Walk(f.logDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Skip the root directory
		if path == f.logDir {
			return nil
		}
		
		// Only process .log files
		if !info.IsDir() && filepath.Ext(path) == ".log" {
			// Extract symbol and date from path
			relPath, err := filepath.Rel(f.logDir, path)
			if err != nil {
				return err
			}
			
			// Path format: symbol/date.log
			parts := strings.Split(relPath, string(filepath.Separator))
			if len(parts) == 2 {
				symbol := parts[0]
				date := strings.TrimSuffix(parts[1], ".log")
				
				logFiles = append(logFiles, domain.LogFile{
					Symbol:   symbol,
					Date:     date,
					Path:     relPath,
					Size:     info.Size(),
					Modified: info.ModTime(),
				})
			}
		}
		
		return nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to read log directory: %w", err)
	}
	
	return logFiles, nil
} 