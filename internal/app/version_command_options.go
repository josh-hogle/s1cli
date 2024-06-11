package app

import (
	"encoding/json"
	goerrors "errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.joshhogle.dev/errorx"
	"go.joshhogle.dev/s1cli/internal/build"
	"go.joshhogle.dev/s1cli/internal/errors"
)

// versionCommandOptions holds specific settings for the 'version' subcommand.
type versionCommandOptions struct {
	// Short represents a flag used to determine whether to show just the version or not.
	Short bool `json:"short"`

	// Verbose represents a flag used to determine whether to show verbose version details or not.
	Verbose bool `json:"verbose"`

	// unexported variables
	appState  *State
	parent    *commandOptions
	configKey string
	isLoaded  bool
}

// jsonVersionCommandOptions is just an alias for versionCommandOptions that is used during marshalling and
// unmarshalling to prevent infinite recursion.
type jsonVersionCommandOptions versionCommandOptions

// newVersionCommandOptions returns a new object with defaults set.
func newVersionCommandOptions(state *State, parent *commandOptions) *versionCommandOptions {
	configKey := _ConfigCommandVersionKey
	viper.SetDefault(fmt.Sprintf("%s.short", configKey), false)
	viper.SetDefault(fmt.Sprintf("%s.verbose", configKey), false)

	return &versionCommandOptions{
		appState:  state,
		parent:    parent,
		configKey: configKey,
	}
}

// BindFlags is used to add command-line flags and bind them to viper configuration keys.
func (c *versionCommandOptions) BindFlags(cmd *cobra.Command) {
	flags := cmd.Flags()
	envPrefix := fmt.Sprintf("%s%s_", build.AppEnvPrefix, strings.ReplaceAll(strings.ToUpper(c.configKey), ".", "_"))

	flags.BoolP("short", "s", false, "show version only")
	viper.BindPFlag(fmt.Sprintf("%s.short", c.configKey), flags.Lookup("short"))
	viper.BindEnv(fmt.Sprintf("%s.short", c.configKey), fmt.Sprintf("%sSHORT", envPrefix))

	flags.BoolP("verbose", "v", false, "show detailed version information")
	viper.BindPFlag(fmt.Sprintf("%s.verbose", c.configKey), flags.Lookup("verbose"))
	viper.BindEnv(fmt.Sprintf("%s.verbose", c.configKey), fmt.Sprintf("%sVERBOSE", envPrefix))
}

// ConfigKey returns the base name of the viper configuration key where the options are stored.
func (c *versionCommandOptions) ConfigKey() string {
	return c.configKey
}

// IsLoaded returns whether or not the configuration settings have been loaded.
func (c *versionCommandOptions) IsLoaded() bool {
	return c.isLoaded
}

// Load converts the corresponding viper configuration and loads it into this configuration object, validating
// settings along the way.
//
// If the options have already been loaded, they will not be loaded again.
//
// The following errors are returned by this function:
// ConfigValidateFailure
func (c *versionCommandOptions) Load() errorx.Error {
	if c.isLoaded {
		return nil
	}
	if errx := c.parent.Load(); errx != nil {
		return errx
	}
	logger := c.appState.logger
	viperConfig := c.appState.config.viperConfig.CommandOptions.Version

	// --short and --verbose are mutually exclusive flags
	// NOTE: we need to check these here since viper won't pick up if we just set them exclusive via cobra
	if viperConfig.Short && viperConfig.Verbose {
		errx := errors.NewConfigValidateFailure(c.appState.config.globalOptions.ConfigFile, "short", "true",
			goerrors.New("--short and --verbose are mutually exclusive flags"))
		logger.Error().
			Err(errx).
			Str("option", "short").
			Str("value", "true").
			Msg(errx.Error())
		return errx
	}

	// save options
	c.Short = viperConfig.Short
	c.Verbose = viperConfig.Verbose

	c.isLoaded = true
	return nil
}

// LogSettings simply writes the object settings to the log.
func (c *versionCommandOptions) LogSettings(recurse bool) {
	if recurse {
		c.parent.LogSettings(recurse)
	}
	c.appState.Logger().Debug().Any("options", c.StringMap()).Msg("loaded 'version' subcommand options")
}

// MarshalJSON overrides how the object is marshalled to JSON to alter how field values are presented or to
// add additional fields.
//
// Any errors returned by this function are a result of calling json.Marshal().
func (c *versionCommandOptions) MarshalJSON() ([]byte, error) {
	opt := jsonVersionCommandOptions(*c)
	return json.Marshal(&opt)
}

// StringMap returns a map of strings to any type as a representation of the configuration.
func (c *versionCommandOptions) StringMap() map[string]any {
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
func (c *versionCommandOptions) String() string {
	output, err := json.Marshal(c)
	if err != nil {
		return fmt.Sprintf("error marshalling object to JSON: %s", err.Error())
	}
	return string(output)
}

// viperVersionCommandOptions holds the options for the 'version' subcommand.
type viperVersionCommandOptions struct {
	Short   bool `mapstructure:"short"`
	Verbose bool `mapstructure:"verbose"`
}
