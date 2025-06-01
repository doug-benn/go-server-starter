package logging

import (
	"time"

	"github.com/rs/zerolog"
)

type loggingConfig struct {
	ConsoleOutput bool
	FileOutput    bool
	PrettyConsole bool
	LogLevel      int
	TimeFormat    string
	PartsOrder    []string
	//File Logging Options
	FilePath         string
	FileMaxSizeInMB  int
	FileMaxBackups   int
	FileMaxAgeInDays int
	UserLocalTime    bool
	CompressLogs     bool
}

func NewLoggingConfig() loggingConfig {
	return loggingConfig{
		ConsoleOutput: true,
		PrettyConsole: false,
		FileOutput:    false,
		LogLevel:      1,
		TimeFormat:    time.RFC3339,
		PartsOrder: []string{
			zerolog.LevelFieldName,
			zerolog.TimestampFieldName,
			zerolog.MessageFieldName,
		},
		//File Logging Options
		FilePath:         "logs/zero-logs.log",
		FileMaxSizeInMB:  5,
		FileMaxBackups:   10,
		FileMaxAgeInDays: 14,
		UserLocalTime:    true,
		CompressLogs:     false,
	}
}
