package logger

import (
	"io"
	"log/slog"
	"os"

	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	defaultFilePath        = "logs/logs.json"
	defaultUserLocalTime   = false
	defaultFileMaxSizeInMB = 10
	defaultFileAgeInDays   = 30
	defaultLogLevel        = slog.LevelInfo
)

// log.SetOutput(&lumberjack.Logger{
//     Filename:   "/var/log/myapp/foo.log",
//     MaxSize:    500, // megabytes
//     MaxBackups: 3,
//     MaxAge:     28, //days
//     Compress:   true, // disabled by default
// })

type Config struct {
	FilePath         string     `koanf:"file_path"`
	UserLocalTime    bool       `koanf:"use_local_time"`
	FileMaxSizeInMB  int        `koanf:"file_max_size_in_mb"`
	FileMaxAgeInDays int        `koanf:"file_max_age_in_days"`
	LogLevel         slog.Level `koanf:"log_level"`
}

var l *slog.Logger

func init() {
	fileWriter := &lumberjack.Logger{
		Filename:  defaultFilePath,
		LocalTime: defaultUserLocalTime,
		MaxSize:   defaultFileMaxSizeInMB,
		MaxAge:    defaultFileAgeInDays,
	}
	l = slog.New(slog.NewJSONHandler(io.MultiWriter(fileWriter, os.Stdout), &slog.HandlerOptions{
		Level: defaultLogLevel,
	}))
}

func L() *slog.Logger {
	return l
}

func New(cfg Config, opt *slog.HandlerOptions, writeInConsole bool) *slog.Logger {
	fileWriter := &lumberjack.Logger{
		Filename:  cfg.FilePath,
		LocalTime: cfg.UserLocalTime,
		MaxSize:   cfg.FileMaxSizeInMB,
		MaxAge:    cfg.FileMaxAgeInDays,
	}

	if writeInConsole {
		return slog.New(slog.NewJSONHandler(io.MultiWriter(fileWriter, os.Stdout), opt))
	}

	return slog.New(slog.NewJSONHandler(fileWriter, opt))
}
