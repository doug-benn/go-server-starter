package logging

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

func NewZeroLogger(config loggingConfig) zerolog.Logger {

	outputs := []io.Writer{}

	if config.ConsoleOutput {
		if config.PrettyConsole {
			var prettyOutput io.Writer = zerolog.ConsoleWriter{
				Out:        os.Stdout,
				TimeFormat: time.RFC3339,
				PartsOrder: config.PartsOrder,
			}
			outputs = append(outputs, prettyOutput)
		} else {
			var consoleOutput io.Writer = os.Stdout
			zerolog.TimeFieldFormat = config.TimeFormat
			outputs = append(outputs, consoleOutput)
		}
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

	logger := zerolog.New(zerolog.MultiLevelWriter(outputs...)).
		Level(zerolog.Level(config.LogLevel)).
		With().
		Timestamp().
		Logger()

	return logger
}
