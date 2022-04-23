package config

import (
	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap/zapcore"
)

// Config define application config object
type Config struct {
	DatabaseDsn      string `envconfig:"DATABASE_DSN" required:"true" default:"127.0.0.1:2379"`
	LogFilePath      string `envconfig:"LOG_FILE_PATH" required:"false" default:"./logs/log.txt"`
	LogToFileEnabled bool   `envconfig:"LOG_TO_FILE_ENABLED" required:"false" default:"false"`
	LogLevel         string `envconfig:"LOG_LEVEL" required:"false" default:"error"`
}

// NewConfig returns actual config instance
func NewConfig() (*Config, error) {
	cfg := &Config{}
	err := envconfig.Process("", cfg)

	return cfg, err
}

func GetZapLevel(level string) zapcore.Level {
	switch level {
	case LogInfo:
		return zapcore.InfoLevel
	case LogWarn:
		return zapcore.WarnLevel
	case LogDebug:
		return zapcore.DebugLevel
	case LogError:
		return zapcore.ErrorLevel
	case LogFatal:
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}
