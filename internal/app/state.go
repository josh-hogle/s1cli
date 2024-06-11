package app

import (
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.joshhogle.dev/errorx"
	"go.joshhogle.dev/s1cli/internal/build"
	"go.joshhogle.dev/s1cli/internal/errors"
)

// State stores the state of the currently running application.
//
// Be sure to construct the State object using the InitState() function. The configuration stored in the
// state is loaded lazily and only as it is needed.
type State struct {
	// unexported variables
	config      *config
	logger      *zerolog.Logger
	productInfo *build.ProductInfo
	startTime   time.Time
}

// NewState creates and initializes the application state.
func NewState() *State {
	s := &State{
		startTime:   time.Now().UTC(),
		productInfo: build.NewProductInfo(),
	}
	s.config = newConfig(s)
	s.initLogger(zerolog.InfoLevel)
	return s
}

// Cleanup is responsible for cleaning up any open handles, flushing log data and any other general state cleanup
// before the application exits.
func (s *State) Cleanup() {
}

// Config returns the app configuration settings.
//
// To determine if the config has been loaded yet, use the config object's IsLoaded() function.
func (s *State) Config() *config {
	return s.config
}

// DisableLogger disables (or re-enables) the logger.
func (s *State) DisableLogger(disable bool) {
	if disable {
		logger := s.logger.Level(zerolog.Disabled)
		s.logger = &logger
	} else {
		logger := s.logger.Level(s.config.globalOptions.LogLevel)
		s.logger = &logger
	}
}

// Initialize loads the configuration settings from the config file / command-line and then configures the global
// options and logging service.
//
// The following errors are returned by this function:
// ConfigValidateFailure, ProviderConversionFailure, ProviderNotFound, ProviderNotSupported
func (s *State) Initialize(cmd *cobra.Command) errorx.Error {
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	flags := cmd.Flags()

	// adjust log level before proceeding if specified as a flag
	// this is primarily to catch errors while loading the configuration using viper
	visitedFlags := map[string]bool{}
	flags.Visit(func(flag *pflag.Flag) {
		visitedFlags[flag.Name] = true
	})
	if _, ok := visitedFlags[_FlagGlobalOptionsLogLevel]; ok {
		logLevel, err := flags.GetString(_FlagGlobalOptionsLogLevel)
		if err != nil {
			errx := errors.NewConfigValidateFailure("", _FlagGlobalOptionsLogLevel, logLevel, err)
			s.logger.Error().
				Err(errx).
				Str("option", _FlagGlobalOptionsLogLevel).
				Str("value", logLevel).
				Msg(errx.Error())
			return errx
		}

		// --log-level must be a valid string
		level, err := zerolog.ParseLevel(logLevel)
		if err != nil {
			errx := errors.NewConfigValidateFailure("", _FlagGlobalOptionsLogLevel, logLevel, err)
			s.logger.Error().
				Err(errx).
				Str("option", _FlagGlobalOptionsLogLevel).
				Str("value", logLevel).
				Msg(errx.Error())
			return errx
		}
		logger := s.logger.Level(level)
		if level <= zerolog.DebugLevel {
			logger = logger.With().Caller().Logger()
		}
		s.logger = &logger
	}

	// set product environment variables
	for k, v := range s.productInfo.Env {
		if err := os.Setenv(k, v); err != nil {
			errx := errors.NewGeneralFailure(fmt.Sprintf("failed to set environment variable '%s'", k), err)
			s.logger.Error().
				Err(errx).
				Str("env_var", k).
				Str("value", v).
				Msg(errx.Error())
			return errx
		}
	}

	// load the configuration
	configFile := s.config.globalOptions.viperConfigFile()
	if errx := s.config.load(configFile); errx != nil {
		return errx
	}

	// configure global settings
	if errx := s.config.globalOptions.Load(); errx != nil {
		return errx
	}
	return nil
}

// Logger returns the app logger.
func (s *State) Logger() *zerolog.Logger {
	return s.logger
}

// ProductInfo returns build information about the application.
func (s *State) ProductInfo() *build.ProductInfo {
	return s.productInfo
}

// Uptime returns the duration of time that the application has been running.
func (s *State) Uptime() time.Duration {
	return time.Since(s.startTime)
}

// initLogger is responsible for initializing and returning the application logger.
//
// The logger created prints any messages below a LevelWarn level to stdout and any messages at or above LevelWarn
// to stderr.
func (s *State) initLogger(level zerolog.Level) {
	isDebugEnabled := false
	if s.productInfo.IsDeveloperBuild || level <= zerolog.DebugLevel {
		isDebugEnabled = true
	}

	stdoutWriter := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "03:04:05PM",
	}
	stdoutCondition := NewFilteredLevelWriterCondition(func(level zerolog.Level) bool {
		return level < zerolog.WarnLevel
	})
	stderrWriter := zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: "03:04:05PM",
	}
	stderrCondition := NewFilteredLevelWriterCondition(func(level zerolog.Level) bool {
		return level >= zerolog.WarnLevel
	})
	multiWriter := zerolog.MultiLevelWriter(
		NewFilteredLevelWriter(stdoutWriter, []*FilteredLevelWriterCondition{stdoutCondition}),
		NewFilteredLevelWriter(stderrWriter, []*FilteredLevelWriterCondition{stderrCondition}),
	)

	logger := zerolog.New(multiWriter).With().Timestamp().Logger().Level(level)
	if isDebugEnabled {
		logger = logger.With().Caller().Logger()
	}
	s.logger = &logger
}
