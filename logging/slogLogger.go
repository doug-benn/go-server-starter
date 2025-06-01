package logging

import (
	"io"
	"log/slog"
	"os"

	"gopkg.in/natefinch/lumberjack.v2"
)

func NewSlogLogger(config loggingConfig) *slog.Logger {
	outputs := []io.Writer{}

	if config.ConsoleOutput {
		consoleOutput := os.Stdout
		outputs = append(outputs, consoleOutput)
	}

	if config.FileOutput {
		fileOutput := &lumberjack.Logger{
			Filename:   config.FilePath,
			MaxSize:    config.FileMaxSizeInMB,
			MaxBackups: config.FileMaxBackups,
			MaxAge:     config.FileMaxAgeInDays,
			LocalTime:  config.UserLocalTime,
			Compress:   config.CompressLogs,
		}
		outputs = append(outputs, fileOutput)
	}

	logger := slog.New(slog.NewJSONHandler(io.MultiWriter(outputs...), &slog.HandlerOptions{
		Level: slog.Level(config.LogLevel),
	}))

	return logger
}
