package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"

	"gateway-service/pkg/config"
)

// New builds a logrus logger based on gateway configuration.
func New(cfg config.LogConfig) (*logrus.Logger, error) {
	log := logrus.New()

	level, err := logrus.ParseLevel(strings.ToLower(cfg.Level))
	if err != nil {
		level = logrus.InfoLevel
	}
	log.SetLevel(level)

	switch strings.ToLower(cfg.Format) {
	case "text":
		log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		})
	default:
		log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
		})
	}

	writer, err := resolveOutput(cfg.Output)
	if err != nil {
		return nil, err
	}
	log.SetOutput(writer)

	return log, nil
}

func resolveOutput(output string) (io.Writer, error) {
	switch strings.ToLower(output) {
	case "", "stdout":
		return os.Stdout, nil
	case "stderr":
		return os.Stderr, nil
	default:
		if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
			return nil, fmt.Errorf("prepare log directory: %w", err)
		}
		file, err := os.OpenFile(output, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			return nil, fmt.Errorf("open log file: %w", err)
		}
		return file, nil
	}
}

