package app

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.joshhogle.dev/errorx"
	"go.joshhogle.dev/s1cli/internal/build"
	"go.joshhogle.dev/s1cli/internal/errors"
)

// globalOptions holds global configuration settings.
type globalOptions struct {
	// APIKey is the API key to use for authentication with the SentinelOne API.
	APIKey string `json:"api_key"`

	// ConfigDir is the directory in which the configuration file is located.
	ConfigDir string `json:"config_dir"`

	// ConfigFile is the configuration file from which the configuration was read.
	ConfigFile string `json:"config_file"`

	// LogLevel identifies the minimum level of messages to log.
	LogLevel zerolog.Level `json:"log_level"`

	// TenantURL is the URL for the customer's SentinelOne SaaS tenant.
	TenantURL string `json:"tenant_url"`

	// unexported variables
	appState  *State
	parent    *config
	configKey string
	isLoaded  bool
}

// jsonGlobalOptions is just an alias for globalOptions that is used during marshalling and unmarshalling to
// prevent infinite recursion.
type jsonGlobalOptions globalOptions

// newGlobalOptions returns a new object with defaults set.
func newGlobalOptions(state *State, parent *config) *globalOptions {
	configKey := _ConfigGlobalKey
	viper.SetDefault(fmt.Sprintf("%s.api_key", configKey), "")
	if state.productInfo.IsDeveloperBuild {
		viper.SetDefault(fmt.Sprintf("%s.log_level", configKey), zerolog.DebugLevel)
	} else {
		viper.SetDefault(fmt.Sprintf("%s.log_level", configKey), zerolog.InfoLevel)
	}
	viper.SetDefault(fmt.Sprintf("%s.tenant_url", configKey), "")

	return &globalOptions{
		appState:  state,
		parent:    parent,
		configKey: configKey,
	}
}

// BindFlags is used to add command-line flags and bind them to viper configuration keys.
func (c *globalOptions) BindFlags(cmd *cobra.Command) {
	persistentFlags := cmd.PersistentFlags()
	envPrefix := fmt.Sprintf("%s%s_", build.AppEnvPrefix, strings.ReplaceAll(strings.ToUpper(c.configKey), ".", "_"))

	// API key
	persistentFlags.StringP("api-key", "k", "", "SentinelOne API key")
	viper.BindPFlag(fmt.Sprintf("%s.api_key", c.configKey), persistentFlags.Lookup("api-key"))
	viper.BindEnv(fmt.Sprintf("%s.api_key", c.configKey), fmt.Sprintf("%sAPI_KEY", envPrefix))

	// config file
	persistentFlags.StringP("config-file", "f", "", "path to configuration file")
	viper.BindPFlag(fmt.Sprintf("%s.config_file", c.configKey), persistentFlags.Lookup("config-file"))
	viper.BindEnv(fmt.Sprintf("%s.config_file", c.configKey), fmt.Sprintf("%sCONFIG_FILE", envPrefix))

	// log level
	usage := "set logging level to trace, debug, info, notice, warn, error, fatal or panic"
	if c.appState.productInfo.IsDeveloperBuild {
		persistentFlags.StringP(_FlagGlobalOptionsLogLevel, "l", zerolog.DebugLevel.String(), usage)
	} else {
		persistentFlags.StringP(_FlagGlobalOptionsLogLevel, "l", zerolog.InfoLevel.String(), usage)
	}
	viper.BindPFlag(fmt.Sprintf("%s.log_level", c.configKey), persistentFlags.Lookup(_FlagGlobalOptionsLogLevel))
	viper.BindEnv(fmt.Sprintf("%s.log_level", c.configKey), fmt.Sprintf("%s_LOG_LEVEL", envPrefix))

	// tenant URL
	persistentFlags.StringP("tenant-url", "t", "", "SentinelOne tenant URL")
	viper.BindPFlag(fmt.Sprintf("%s.tenant_url", c.configKey), persistentFlags.Lookup("tenant-url"))
	viper.BindEnv(fmt.Sprintf("%s.tenant_url", c.configKey), fmt.Sprintf("%sTENANT_URL", envPrefix))
}

// ConfigKey returns the base name of the viper configuration key where the options are stored.
func (c *globalOptions) ConfigKey() string {
	return c.configKey
}

// IsLoaded returns whether or not the configuration settings have been loaded.
func (c *globalOptions) IsLoaded() bool {
	return c.isLoaded
}

// Load converts the corresponding viper configuration and loads it into this configuration object, validating
// settings along the way.
//
// If the options have already been loaded, they will not be loaded again.
//
// The following errors are returned by this function:
// ConfigLoadFailure, ConfigValidateFailure
func (c *globalOptions) Load() errorx.Error {
	if c.isLoaded {
		return nil
	}
	logger := c.appState.logger
	viperConfig := c.appState.config.viperConfig.GlobalOptions

	// NOTE: API key and tenant URL are almost always required, however, we don't check that here because
	// commands like 'configure' and 'version' do not require them and this could cause a failure here if
	// they were required
	c.APIKey = viperConfig.APIKey
	c.TenantURL = viperConfig.TenantURL

	// check log level
	level, err := zerolog.ParseLevel(viperConfig.LogLevel)
	if err != nil {
		errx := errors.NewConfigValidateFailure(c.ConfigFile, _FlagGlobalOptionsLogLevel, viperConfig.LogLevel, err)
		logger.Error().
			Err(errx).
			Str("option", _FlagGlobalOptionsLogLevel).
			Str("value", viperConfig.LogLevel).
			Msg(errx.Error())
		return errx
	}
	newLogger := logger.Level(level)
	if level <= zerolog.DebugLevel {
		newLogger = newLogger.With().Caller().Logger()
	}
	c.appState.logger = &newLogger
	c.LogLevel = level

	// save the absolute path to the directory in which the config file is located
	absPath, err := filepath.Abs(c.ConfigFile)
	if err != nil {
		errx := errors.NewConfigLoadFailure(c.ConfigFile, err)
		logger.Error().
			Err(errx).
			Str("config_file", c.ConfigFile).
			Msg(errx.Error())
		return errx
	}
	if c.ConfigFile == "" {
		c.ConfigDir = absPath
	} else {
		c.ConfigDir = filepath.Dir(absPath)
	}

	c.isLoaded = true
	return nil
}

// LogSettings simply writes the object settings to the log.
func (c *globalOptions) LogSettings() {
	c.appState.logger.Debug().Any("options", c.StringMap()).Msg("loaded global options")
}

// MarshalJSON overrides how the object is marshalled to JSON to alter how field values are presented or to
// add additional fields.
//
// Any errors returned by this function are a result of calling json.Marshal().
func (c *globalOptions) MarshalJSON() ([]byte, error) {
	cfg := jsonGlobalOptions(*c)
	return json.Marshal(&cfg)
}

// StringMap returns a map of strings to any type as a representation of the configuration.
func (c *globalOptions) StringMap() map[string]any {
	asString := c.String()
	var stringMap map[string]any
	if err := json.Unmarshal([]byte(asString), &stringMap); err != nil {
		return map[string]any{
			"error": fmt.Sprintf("error marshalling object to JSON: %s", err.Error()),
		}
	}
	return stringMap
}

// Returns a string representation of the configuration as JSON.
func (c *globalOptions) String() string {
	output, err := json.Marshal(c)
	if err != nil {
		return fmt.Sprintf("error marshalling object to JSON: %s", err.Error())
	}
	return string(output)
}

// viperConfigFile returns the config file that should be parsed by Viper.
func (c *globalOptions) viperConfigFile() string {
	return viper.GetString(fmt.Sprintf("%s.config_file", c.configKey))
}

// viperGlobalOptions holds the global options for the root command.
type viperGlobalOptions struct {
	APIKey    string `mapstructure:"api_key"`
	LogLevel  string `mapstructure:"log_level"`
	TenantURL string `mapstructure:"tenant_url"`
}
