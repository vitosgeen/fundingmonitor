package infrastructure

import (
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
	dateStr := time.Now().Format("02-01-2006")
	filename := filepath.Join(pairDir, fmt.Sprintf("%s.log", dateStr))

	// Append to file
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file for %s: %w", symbol, err)
	}
	defer file.Close()

	// Write timestamp and symbol header
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	header := fmt.Sprintf("[%s] Symbol: %s\n", timestamp, symbol)
	if _, err := file.WriteString(header); err != nil {
		return fmt.Errorf("failed to write header to log file for %s: %w", symbol, err)
	}

	// Write each rate on a separate line
	for _, rate := range rates {
		rateLine := fmt.Sprintf("  Exchange: %s, Funding Rate: %.6f, Mark Price: %.2f, Index Price: %.2f\n",
			rate.Exchange, rate.FundingRate, rate.MarkPrice, rate.IndexPrice)
		if _, err := file.WriteString(rateLine); err != nil {
			return fmt.Errorf("failed to write rate to log file for %s: %w", symbol, err)
		}
	}

	// Write separator
	if _, err := file.WriteString("\n"); err != nil {
		return fmt.Errorf("failed to write separator to log file for %s: %w", symbol, err)
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
