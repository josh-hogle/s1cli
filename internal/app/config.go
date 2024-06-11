package app

import (
	goerrors "errors"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"go.joshhogle.dev/errorx"
	"go.joshhogle.dev/s1cli/internal/errors"
)

// config is an internal structure to hold the application configuration.
type config struct {
	// unexported variables
	appState       *State
	globalOptions  *globalOptions
	commandOptions *commandOptions
	isLoaded       bool
	viperConfig    viperConfig
}

// newConfig returns a new object with defaults set.
func newConfig(state *State) *config {
	config := &config{
		appState: state,
	}
	config.globalOptions = newGlobalOptions(state, config)
	config.commandOptions = newCommandOptions(state, config)
	return config
}

// CommandOptions returns the configuration settings for all commands.
//
// To determine if the settings have been loaded, use the object's IsLoaded() function.
func (c *config) CommandOptions() *commandOptions {
	return c.commandOptions
}

// GlobalOptions returns the global configuration settings.
//
// To determine if the settings have been loaded, use the object's IsLoaded() function.
func (c *config) GlobalOptions() *globalOptions {
	return c.globalOptions
}

// IsLoaded returns whether or not the configuration settings have been loaded.
func (c *config) IsLoaded() bool {
	return c.isLoaded
}

// load simply loads the configuration settings into memory.
//
// It is the caller's responsibility to validate the configuration settings once they have been loaded.
//
// The config file is determined as follows:
//
//	◽ If the --config-file option is specified on the command-line, use that file.
//	◽ If the appropriate <PREFIX>CONFIG_FILE environment variable is set, use that file.
//	◽ Use the config.yaml file in the current working directory if it exists.
//
// The following errors are returned by this function:
// ConfigLoadFailure, ConfigParseFailure
func (c *config) load(file string) errorx.Error {
	// config file was specified on the command-line
	if file != "" {
		if errx := c.loadFile(file); errx != nil {
			return errx
		}
		c.isLoaded = true
		return nil
	}

	// use the default config file
	if errx := c.loadDefaultFile(); errx != nil {
		return errx
	}
	c.isLoaded = true
	return nil
}

// loadDefaultFile attempts to load a default configuration file from the user's configuration folder.
//
// The following errors are returned by this function:
// ConfigLoadFailure, ConfigParseFailure
func (c *config) loadDefaultFile() errorx.Error {
	logger := c.appState.Logger()

	// no specific config file was specified so we'll check for a default config file
	viper.AddConfigPath(_DefaultConfigDir)
	viper.SupportedExts = []string{"yaml", "yml"}
	viper.SetConfigType("yaml")
	viper.SetConfigName(_DefaultConfigFileBaseName)

	// read the configuration file
	if err := viper.ReadInConfig(); err != nil {
		// since the default config file is being used but was not found, do not return an error
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return c.unmarshal()
		}
		configFile := viper.ConfigFileUsed()
		errx := errors.NewConfigLoadFailure(configFile, err)
		logger.Error().
			Err(errx).
			Str("config_file", configFile).
			Msg(errx.Error())
		return errx
	}

	c.globalOptions.ConfigFile = viper.ConfigFileUsed()
	return c.unmarshal()
}

// loadFile loads the specified configuration file.
//
// The following errors are returned by this function:
// ConfigLoadFailure, ConfigParseFailure
func (c *config) loadFile(file string) errorx.Error {
	logger := c.appState.Logger()

	file, err := filepath.Abs(os.ExpandEnv(file))
	if err != nil {
		errx := errors.NewConfigLoadFailure(viper.ConfigFileUsed(), err)
		logger.Error().
			Err(errx).
			Str("config_file", viper.ConfigFileUsed()).
			Msg(errx.Error())
		return errx
	}

	// read the configuration file
	viper.SetConfigFile(file)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			errx := errors.NewConfigLoadFailure(viper.ConfigFileUsed(), goerrors.New("configuration file not found"))
			logger.Error().
				Err(errx).
				Str("config_file", viper.ConfigFileUsed()).
				Msg(errx.Error())
			return errx
		}
		errx := errors.NewConfigLoadFailure(viper.ConfigFileUsed(), err)
		logger.Error().
			Err(errx).
			Str("config_file", viper.ConfigFileUsed()).
			Msg(errx.Error())
		return errx
	}
	c.globalOptions.ConfigFile = viper.ConfigFileUsed()
	return c.unmarshal()
}

// unmarshal simply unmarshals the data from the config file into the object.
//
// The following errors are returned by this function:
// ConfigParseFailure
func (c *config) unmarshal() errorx.Error {
	logger := c.appState.Logger()

	var viperCfg viperConfig
	if err := viper.Unmarshal(&viperCfg); err != nil {
		errx := errors.NewConfigParseFailure(viper.ConfigFileUsed(), err)
		logger.Error().
			Err(errx).
			Str("config_file", viper.ConfigFileUsed()).
			Msg(errx.Error())
		return errx
	}
	c.viperConfig = viperCfg
	return nil
}

// viperConfig is used for unmarshaling the configuration file, environment variables and CLI flags using viper.
type viperConfig struct {
	GlobalOptions  viperGlobalOptions  `mapstructure:"global"`
	CommandOptions viperCommandOptions `mapstructure:"command"`
}
