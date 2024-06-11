package app

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/spf13/cobra"
	"go.joshhogle.dev/errorx"
)

// commandOptions holds options for all of the subcommands.
type commandOptions struct {
	// unexported variables
	appState  *State
	parent    *config
	configKey string
	isLoaded  bool
	/*
		configureOptions *configureCommandOptions
		configureOptionsOnce *sync.Once
	*/
	provisionOptions     *provisionCommandOptions
	provisionOptionsOnce *sync.Once
	versionOptions       *versionCommandOptions
	versionOptionsOnce   *sync.Once
}

// jsonCommandOptions is just an alias for commandOptions that is used during marshalling and unmarshalling to
// prevent infinite recursion.
type jsonCommandOptions commandOptions

// newCommandOptions returns a new object with defaults set.
func newCommandOptions(state *State, parent *config) *commandOptions {
	configKey := _ConfigCommandKey

	return &commandOptions{
		appState:             state,
		parent:               parent,
		configKey:            configKey,
		provisionOptionsOnce: &sync.Once{},
		versionOptionsOnce:   &sync.Once{},
	}
}

// BindFlags is used to add command-line flags and bind them to viper configuration keys.
func (c *commandOptions) BindFlags(cmd *cobra.Command) {
}

// ConfigKey returns the base name of the viper configuration key where the options are stored.
func (c *commandOptions) ConfigKey() string {
	return c.configKey
}

// IsLoaded returns whether or not the configuration settings have been loaded.
func (c *commandOptions) IsLoaded() bool {
	return c.isLoaded
}

// Load converts the corresponding viper configuration and loads it into this configuration object, validating
// settings along the way.
//
// If the options have already been loaded, they will not be loaded again.
//
// The following errors are returned by this function:
// ConfigValidateFailure
func (c *commandOptions) Load() errorx.Error {
	if c.isLoaded {
		return nil
	}

	c.isLoaded = true
	return nil
}

// LogSettings simply writes the object settings to the log.
//
// If recurse is true, the global options are logged as well.
func (c *commandOptions) LogSettings(recurse bool) {
	if recurse {
		c.appState.config.globalOptions.LogSettings()
	}
	c.appState.logger.Debug().Any("options", c.StringMap()).Msg("loaded root command options")
}

// MarshalJSON overrides how the object is marshalled to JSON to alter how field values are presented or to
// add additional fields.
//
// Any errors returned by this function are a result of calling json.Marshal().
func (c *commandOptions) MarshalJSON() ([]byte, error) {
	cfg := jsonCommandOptions(*c)
	//lint:ignore SA9005 this function may change in the future to export fields
	return json.Marshal(&cfg)
}

// ProvisionOptions returns the options for the "provision" subcommand.
//
// If the options object has not been initialized, it is automatically initialized. However, the settings
// are *not* automatically loaded when the object is initialized. To determine if the settings have been loaded, use
// the object's IsLoaded() function.
func (c *commandOptions) Provision() *provisionCommandOptions {
	c.provisionOptionsOnce.Do(func() {
		c.provisionOptions = newProvisionCommandOptions(c.appState, c)
	})
	return c.provisionOptions
}

// StringMap returns a map of strings to any type as a representation of the configuration.
func (c *commandOptions) StringMap() map[string]any {
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
func (c *commandOptions) String() string {
	output, err := json.Marshal(c)
	if err != nil {
		return fmt.Sprintf("error marshalling object to JSON: %s", err.Error())
	}
	return string(output)
}

// VersionOptions returns the options for the "version" subcommand.
//
// If the options object has not been initialized, it is automatically initialized. However, the settings
// are *not* automatically loaded when the object is initialized. To determine if the settings have been loaded, use
// the object's IsLoaded() function.
func (c *commandOptions) Version() *versionCommandOptions {
	c.versionOptionsOnce.Do(func() {
		c.versionOptions = newVersionCommandOptions(c.appState, c)
	})
	return c.versionOptions
}

// viperCommandOptions holds the options for all subcommands.
type viperCommandOptions struct {
	Provision viperProvisionCommandOptions `mapstructure:"provision"`
	Version   viperVersionCommandOptions   `mapstructure:"version"`
}
