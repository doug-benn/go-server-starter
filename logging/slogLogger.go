package logging

import (
	"io"
	"log/slog"
	"os"

	"gopkg.in/natefinch/lumberjack.v2"
)

func NewSlogLogger(options ...func(l *LoggingConfig)) *slog.Logger {

	config := LoggingConfig{
		ConsoleOutput:    true,
		FileOutput:       true,
		LogLevel:         1,
		FilePath:         "logs/slog-logs.log",
		FileMaxSizeInMB:  5,
		FileMaxBackups:   10,
		FileMaxAgeInDays: 14,
		UserLocalTime:    true,
		CompressLogs:     false,
	}

	for _, opt := range options {
		opt(&config)
	}

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
