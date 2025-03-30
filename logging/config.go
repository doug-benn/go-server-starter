package logging

type LoggingConfig struct {
	ConsoleOutput bool
	FileOutput    bool
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
