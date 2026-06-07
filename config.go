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
	Listen                string
	Port                  int
	Target                string
	LogLevel              string
	ServedModelName       string
	ThinkingModelName     string
	NoThinkingModelName   string
	EnforceSamplingParams bool
	PreserveThinking      bool
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
	if c.ThinkingModelName == "" {
		return errors.New("thinking model name cannot be empty")
	}
	if c.NoThinkingModelName == "" {
		return errors.New("no-thinking model name cannot be empty")
	}
	return nil
}

func LoadConfig() (Config, error) {
	var cfg Config

	listen := flag.String("listen", "0.0.0.0", "IP address to listen on")
	port := flag.Int("port", 9000, "Port to listen on")
	target := flag.String("target", "http://127.0.0.1:8000", "Backend target, default is for a local vLLM")
	loglevel := flag.String("loglevel", slog.LevelInfo.String(), "Log level (COMPLETE, DEBUG, INFO, WARN, ERROR)")
	servedModel := flag.String("served-model", "", "Name of the served model")
	thinkingModel := flag.String("thinking-model", "", "Name of the thinking model")
	noThinkingModel := flag.String("no-thinking-model", "", "Name of the no-thinking model")
	enforceSampling := flag.Bool("enforce-sampling-params", false, "Enforce sampling parameters, overriding client-provided values")
	preserveThinking := flag.Bool("preserve-thinking", false, "Automatically enable preserve_thinking in chat_template_kwargs for thinking mode")

	flag.Parse()

	cfg.Listen = getEnvOrFlag(*listen, "KIMIRP_LISTEN")
	cfg.Target = getEnvOrFlag(*target, "KIMIRP_TARGET")
	cfg.LogLevel = getEnvOrFlag(*loglevel, "KIMIRP_LOGLEVEL")
	cfg.ServedModelName = getEnvOrFlag(*servedModel, "KIMIRP_SERVED_MODEL_NAME")
	cfg.ThinkingModelName = getEnvOrFlag(*thinkingModel, "KIMIRP_THINKING_MODEL_NAME")
	cfg.NoThinkingModelName = getEnvOrFlag(*noThinkingModel, "KIMIRP_NO_THINKING_MODEL_NAME")

	var err error
	cfg.Port, err = getEnvOrFlagInt(*port, "KIMIRP_PORT")
	if err != nil {
		return cfg, err
	}
	cfg.EnforceSamplingParams, err = getEnvOrFlagBool(*enforceSampling, "KIMIRP_ENFORCE_SAMPLING_PARAMS")
	if err != nil {
		return cfg, err
	}
	cfg.PreserveThinking, err = getEnvOrFlagBool(*preserveThinking, "KIMIRP_PRESERVE_THINKING")
	if err != nil {
		return cfg, err
	}

	return cfg, cfg.Validate()
}

func getEnvOrFlag(flagVal string, envName string) string {
	if envVal, exists := os.LookupEnv(envName); exists {
		return envVal
	}
	return flagVal
}

func getEnvOrFlagInt(flagVal int, envName string) (int, error) {
	if envVal, exists := os.LookupEnv(envName); exists {
		intVal, err := strconv.Atoi(envVal)
		if err != nil {
			return 0, fmt.Errorf("invalid value for %s=%q: %w", envName, envVal, err)
		}
		return intVal, nil
	}
	return flagVal, nil
}

func getEnvOrFlagBool(flagVal bool, envName string) (bool, error) {
	if envVal, exists := os.LookupEnv(envName); exists {
		boolVal, err := strconv.ParseBool(envVal)
		if err != nil {
			return false, fmt.Errorf("invalid value for %s=%q: %w", envName, envVal, err)
		}
		return boolVal, nil
	}
	return flagVal, nil
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
