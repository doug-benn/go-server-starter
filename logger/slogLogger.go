package logger

import (
	"io"
	"log/slog"
	"os"

	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	defaultFilePath        = "logs/slog-logs.log"
	defaultUserLocalTime   = false
	defaultFileMaxSizeInMB = 10
	defaultFileMaxBackups  = 3
	defaultFileAgeInDays   = 30
	defaultLogLevel        = slog.LevelInfo
)

type Config struct {
	FilePath         string
	UserLocalTime    bool
	FileMaxSizeInMB  int
	FileMaxBackups   int
	FileMaxAgeInDays int
	LogLevel         slog.Level
}

var l *slog.Logger

func init() {
	fileWriter := &lumberjack.Logger{
		Filename:   defaultFilePath,
		LocalTime:  defaultUserLocalTime,
		MaxSize:    defaultFileMaxSizeInMB,
		MaxBackups: defaultFileMaxBackups,
		MaxAge:     defaultFileAgeInDays,
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
		Filename:   cfg.FilePath,
		LocalTime:  cfg.UserLocalTime,
		MaxSize:    cfg.FileMaxSizeInMB,
		MaxBackups: cfg.FileMaxBackups,
		MaxAge:     cfg.FileMaxAgeInDays,
	}

	if writeInConsole {
		return slog.New(slog.NewJSONHandler(io.MultiWriter(fileWriter, os.Stdout), opt))
	}

	return slog.New(slog.NewJSONHandler(fileWriter, opt))
}
