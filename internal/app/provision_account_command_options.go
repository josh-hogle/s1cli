package app

import (
	"encoding/json"
	goerrors "errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.joshhogle.dev/errorx"
	"go.joshhogle.dev/s1cli/internal/build"
	"go.joshhogle.dev/s1cli/internal/errors"
)

// provisionAccountCommandOptions holds options for the 'provision account' subcommand.
type provisionAccountCommandOptions struct {
	CSVSeparator             string `json:"csv_separator"`
	CSVSource                string `json:"csv_source"`
	ReactivateExpiredAccount bool   `json:"reactivate_expired_account"`
	ResetFirstUserPassword   bool   `json:"reset_first_user_password"`

	// unexported variables
	appState  *State
	parent    *provisionCommandOptions
	configKey string
	isLoaded  bool
}

// jsonProvisionAccountCommandOptions is just an alias for provisionAccountCommandOptions that is used during
// marshalling and unmarshalling to prevent infinite recursion.
type jsonProvisionAccountCommandOptions provisionAccountCommandOptions

// newProvisionAccountCommandOptions returns a new object with defaults set.
func newProvisionAccountCommandOptions(state *State,
	parent *provisionCommandOptions) *provisionAccountCommandOptions {

	configKey := _ConfigCommandProvisionAccountKey
	viper.SetDefault(fmt.Sprintf("%s.csv_separator", configKey), _DefaultCSVSeparator)
	viper.SetDefault(fmt.Sprintf("%s.csv_source", configKey), "")
	viper.SetDefault(fmt.Sprintf("%s.reactivate_expired_account", configKey), false)
	viper.SetDefault(fmt.Sprintf("%s.reset_first_user_password", configKey), false)

	return &provisionAccountCommandOptions{
		CSVSeparator: _DefaultCSVSeparator,
		appState:     state,
		parent:       parent,
		configKey:    configKey,
	}
}

// BindFlags is used to add command-line flags and bind them to viper configuration keys.
func (c *provisionAccountCommandOptions) BindFlags(cmd *cobra.Command) {
	flags := cmd.Flags()
	envPrefix := fmt.Sprintf("%s%s_", build.AppEnvPrefix, strings.ReplaceAll(strings.ToUpper(c.configKey), ".", "_"))

	// --csv-separator
	flags.String("csv-separator", _DefaultCSVSeparator, "when using a CSV, this is the separator token")
	viper.BindPFlag(fmt.Sprintf("%s.csv_separator", c.configKey), flags.Lookup("csv-separator"))
	viper.BindEnv(fmt.Sprintf("%s.csv_separator", c.configKey), fmt.Sprintf("%sCSV_SEPARATOR", envPrefix))

	// --csv-source
	flags.String("csv-source", "", "provision accounts from the given CSV file")
	viper.BindPFlag(fmt.Sprintf("%s.csv_source", c.configKey), flags.Lookup("csv-source"))
	viper.BindEnv(fmt.Sprintf("%s.csv_source", c.configKey), fmt.Sprintf("%sCSV_SOURCE", envPrefix))

	// --reactivate-expired-account
	flags.Bool("reactivate-expired-account", false, "if an account exists and is expired, reactivate it")
	viper.BindPFlag(fmt.Sprintf("%s.reactivate_expired_account", c.configKey),
		flags.Lookup("reactivate-expired-account"))
	viper.BindEnv(fmt.Sprintf("%s.reactivate_expired_account", c.configKey),
		fmt.Sprintf("%sREACTIVATE_EXPIRED_ACCOUNT", envPrefix))

	// --reset-first-user-password
	flags.Bool("reset-first-user-password", false, "send the first user a password reset email")
	viper.BindPFlag(fmt.Sprintf("%s.reset_first_user_password", c.configKey),
		flags.Lookup("reset-first-user-password"))
	viper.BindEnv(fmt.Sprintf("%s.reset_first_user_password", c.configKey),
		fmt.Sprintf("%sRESET_FIRST_USER_PASSWORD", envPrefix))
}

// ConfigKey returns the base name of the viper configuration key where the options are stored.
func (c *provisionAccountCommandOptions) ConfigKey() string {
	return c.configKey
}

// IsLoaded returns whether or not the configuration settings have been loaded.
func (c *provisionAccountCommandOptions) IsLoaded() bool {
	return c.isLoaded
}

// Load converts the corresponding viper configuration and loads it into this configuration object, validating
// settings along the way.
//
// If the options have already been loaded, they will not be loaded again.
//
// The following errors are returned by this function:
// ConfigValidateFailure
func (c *provisionAccountCommandOptions) Load() errorx.Error {
	if c.isLoaded {
		return nil
	}
	if errx := c.parent.Load(); errx != nil {
		return errx
	}
	viperConfig := c.appState.config.viperConfig.CommandOptions.Provision.Account
	logger := c.appState.logger

	// using a CSV file
	if viperConfig.CSVSource != "" {
		// CSV separator cannot be empty
		if viperConfig.CSVSeparator == "" {
			viperConfig.CSVSeparator = _DefaultCSVSeparator
			logger.Warn().Msgf("an empty CSV separator is not allowed ; defaulting to %s for separator",
				_DefaultCSVSeparator)
		}

		// CSV separator should be a single character
		if len(viperConfig.CSVSeparator) != 1 {
			errx := errors.NewConfigValidateFailure(c.appState.config.globalOptions.ConfigFile, "csv_separator",
				viperConfig.CSVSeparator, goerrors.New("CSV separator must be a single character"))
			logger.Error().
				Err(errx).
				Str("option", "csv_source").
				Str("value", viperConfig.CSVSource).
				Msg(errx.Error())
			return errx
		}

		// make sure CSV file exists
		_, err := os.Stat(viperConfig.CSVSource)
		if err != nil {
			errx := errors.NewConfigValidateFailure(c.appState.config.globalOptions.ConfigFile, "csv_source",
				viperConfig.CSVSource, err)
			logger.Error().
				Err(errx).
				Str("option", "csv_source").
				Str("value", viperConfig.CSVSource).
				Msg(errx.Error())
			return errx
		}
	}

	// save options
	c.CSVSeparator = viperConfig.CSVSeparator
	c.CSVSource = viperConfig.CSVSource
	c.ReactivateExpiredAccount = viperConfig.ReactivateExpiredAccount
	c.ResetFirstUserPassword = viperConfig.ResetFirstUserPassword

	c.isLoaded = true
	return nil
}

// LogSettings simply writes the object settings to the log.
func (c *provisionAccountCommandOptions) LogSettings(recurse bool) {
	if recurse {
		c.parent.LogSettings(recurse)
	}
	c.appState.Logger().Debug().Any("options", c.StringMap()).Msg("loaded 'provision account' subcommand options")
}

// MarshalJSON overrides how the object is marshalled to JSON to alter how field values are presented or to
// add additional fields.
//
// Any errors returned by this function are a result of calling json.Marshal().
func (c *provisionAccountCommandOptions) MarshalJSON() ([]byte, error) {
	opt := jsonProvisionAccountCommandOptions(*c)
	return json.Marshal(&opt)
}

// StringMap returns a map of strings to any type as a representation of the configuration.
func (c *provisionAccountCommandOptions) StringMap() map[string]any {
	asString := c.String()
	var stringMap map[string]any
	if err := json.Unmarshal([]byte(asString), &stringMap); err != nil {
		return map[string]any{
			"error": fmt.Sprintf("error marshalling object to JSON: %s", err.Error()),
		}
	}
	return stringMap
}

// String returns a string representation of the configuration as JSON.
func (c *provisionAccountCommandOptions) String() string {
	output, err := json.Marshal(c)
	if err != nil {
		return fmt.Sprintf("error marshalling object to JSON: %s", err.Error())
	}
	return string(output)
}

// viperProvisionAccouintCommandOptions holds the options for the 'provision account' subcommand.
type viperProvisionAccountCommandOptions struct {
	CSVSeparator             string `mapstructure:"csv_separator"`
	CSVSource                string `mapstructure:"csv_source"`
	ReactivateExpiredAccount bool   `mapstructure:"reactivate_expired_account"`
	ResetFirstUserPassword   bool   `mapstructure:"reset_first_user_password"`
}
