package main

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

// COMPLETE is a log level more verbose than DEBUG for complete request/response dumps
const COMPLETE = slog.LevelDebug - 4
const COMPLETE_LEVEL = "COMPLETE"

type Config struct {
	Listen                    string
	Port                      int
	Target                    string
	LogLevel                  string
	ServedModelName           string
	InstantModelName          string
	ThinkingModelName         string
	PreserveThinkingModelName string
	EnforceSamplingParams     bool
}

func (c Config) Validate() error {
	if c.Listen == "" {
		return errors.New("listen address cannot be empty")
	}
	if c.Port < 1 || c.Port > 65535 {
		return errors.New("port must be between 1 and 65535")
	}
	if c.Target == "" {
		return errors.New("target cannot be empty")
	}
	if c.LogLevel == "" {
		return errors.New("log level cannot be empty")
	}
	if c.ServedModelName == "" {
		return errors.New("served model name cannot be empty")
	}
	if c.InstantModelName == "" {
		return errors.New("instant model name cannot be empty")
	}
	if c.ThinkingModelName == "" {
		return errors.New("thinking model name cannot be empty")
	}
	if c.PreserveThinkingModelName == "" {
		return errors.New("preserve-thinking model name cannot be empty")
	}
	return nil
}

func LoadConfig() (Config, error) {
	listen := flag.String("listen",
		defaultConfigValue("KIMIRP_LISTEN", "0.0.0.0"),
		"IP address to listen on",
	)
	port := flag.Int("port",
		defaultConfigValueInt("KIMIRP_PORT", 9000),
		"Port to listen on",
	)
	target := flag.String("target",
		defaultConfigValue("KIMIRP_TARGET", "http://127.0.0.1:8000"),
		"Backend target, default is for a local vLLM",
	)
	loglevel := flag.String("loglevel",
		defaultConfigValue("KIMIRP_LOGLEVEL", slog.LevelInfo.String()),
		"Log level (COMPLETE, DEBUG, INFO, WARN, ERROR)",
	)
	servedModel := flag.String("served-model",
		defaultConfigValue("KIMIRP_SERVED_MODEL_NAME", ""),
		"Name of the served model",
	)
	instantModel := flag.String("instant-model",
		defaultConfigValue("KIMIRP_INSTANT_MODEL_NAME", ""),
		"Name of the instant model",
	)
	thinkingModel := flag.String("thinking-model",
		defaultConfigValue("KIMIRP_THINKING_MODEL_NAME", "kimi-k26-thinking"),
		"Name of the thinking model",
	)
	preserveThinkingModel := flag.String("preserve-thinking-model",
		defaultConfigValue("KIMIRP_PRESERVE_THINKING_MODEL_NAME", ""),
		"Name of the preserve-thinking model",
	)
	enforceSampling := flag.Bool("enforce-sampling-params",
		defaultConfigValueBool("KIMIRP_ENFORCE_SAMPLING_PARAMS", false),
		"Enforce sampling parameters, overriding client-provided values",
	)
	flag.Parse()

	cfg := Config{
		Listen:                    *listen,
		Port:                      *port,
		Target:                    *target,
		LogLevel:                  *loglevel,
		ServedModelName:           *servedModel,
		InstantModelName:          *instantModel,
		ThinkingModelName:         *thinkingModel,
		PreserveThinkingModelName: *preserveThinkingModel,
		EnforceSamplingParams:     *enforceSampling,
	}
	return cfg, cfg.Validate()
}

func defaultConfigValue(envName, defaultVal string) string {
	if envVal, exists := os.LookupEnv(envName); exists {
		return envVal
	}
	return defaultVal
}

func defaultConfigValueInt(envName string, defaultVal int) int {
	if envVal, exists := os.LookupEnv(envName); exists {
		intVal, err := strconv.Atoi(envVal)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid value for %s=%q: %v, using default %d\n", envName, envVal, err, defaultVal)
			return defaultVal
		}
		return intVal
	}
	return defaultVal
}

func defaultConfigValueBool(envName string, defaultVal bool) bool {
	if envVal, exists := os.LookupEnv(envName); exists {
		boolVal, err := strconv.ParseBool(envVal)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid value for %s=%q: %v, using default %t\n", envName, envVal, err, defaultVal)
			return defaultVal
		}
		return boolVal
	}
	return defaultVal
}

// parseLogLevel parses a log level string, including the COMPLETE level
func parseLogLevel(levelStr string) slog.Level {
	switch strings.ToUpper(levelStr) {
	case COMPLETE_LEVEL:
		return COMPLETE
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
