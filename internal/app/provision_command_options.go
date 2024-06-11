package app

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/spf13/cobra"
	"go.joshhogle.dev/errorx"
)

// provisionCommandOptions holds options for the 'provision' subcommand.
type provisionCommandOptions struct {
	// unexported variables
	appState                           *State
	parent                             *commandOptions
	configKey                          string
	isLoaded                           bool
	provisionAccountCommandOptions     *provisionAccountCommandOptions
	provisionAccountCommandOptionsOnce *sync.Once
}

// jsonProvisionCommandOptions is just an alias for provisionCommandOptions that is used during marshalling and
// unmarshalling to prevent infinite recursion.
type jsonProvisionCommandOptions provisionCommandOptions

// newProvisionCommandOptions returns a new object with defaults set.
func newProvisionCommandOptions(state *State, parent *commandOptions) *provisionCommandOptions {
	configKey := _ConfigCommandProvisionKey

	return &provisionCommandOptions{
		appState:                           state,
		parent:                             parent,
		configKey:                          configKey,
		provisionAccountCommandOptionsOnce: &sync.Once{},
	}
}

// BindFlags is used to add command-line flags and bind them to viper configuration keys.
func (c *provisionCommandOptions) BindFlags(cmd *cobra.Command) {
}

// ConfigKey returns the base name of the viper configuration key where the options are stored.
func (c *provisionCommandOptions) ConfigKey() string {
	return c.configKey
}

// IsLoaded returns whether or not the configuration settings have been loaded.
func (c *provisionCommandOptions) IsLoaded() bool {
	return c.isLoaded
}

// Load converts the corresponding viper configuration and loads it into this configuration object, validating
// settings along the way.
//
// If the options have already been loaded, they will not be loaded again.
//
// The following errors are returned by this function:
// ConfigValidateFailure
func (c *provisionCommandOptions) Load() errorx.Error {
	if c.isLoaded {
		return nil
	}
	if errx := c.parent.Load(); errx != nil {
		return errx
	}

	c.isLoaded = true
	return nil
}

// LogSettings simply writes the object settings to the log.
func (c *provisionCommandOptions) LogSettings(recurse bool) {
	if recurse {
		c.parent.LogSettings(recurse)
	}
	c.appState.logger.Debug().Any("options", c.StringMap()).Msg("loaded 'provision' subcommand options")
}

// MarshalJSON overrides how the object is marshalled to JSON to alter how field values are presented or to
// add additional fields.
//
// Any errors returned by this function are a result of calling json.Marshal().
func (c *provisionCommandOptions) MarshalJSON() ([]byte, error) {
	cfg := jsonProvisionCommandOptions(*c)
	//lint:ignore SA9005 this function may change in the future to export fields
	return json.Marshal(&cfg)
}

// StringMap returns a map of strings to any type as a representation of the configuration.
func (c *provisionCommandOptions) StringMap() map[string]any {
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
func (c *provisionCommandOptions) String() string {
	output, err := json.Marshal(c)
	if err != nil {
		return fmt.Sprintf("error marshalling object to JSON: %s", err.Error())
	}
	return string(output)
}

// VersionOptions returns the options for the "version" command.
//
// If the options object has not been initialized, it is automatically initialized. However, the settings
// are *not* automatically loaded when the object is initialized. To determine if the settings have been loaded, use
// the object's IsLoaded() function.
func (c *provisionCommandOptions) Account() *provisionAccountCommandOptions {
	c.provisionAccountCommandOptionsOnce.Do(func() {
		c.provisionAccountCommandOptions = newProvisionAccountCommandOptions(c.appState, c)
	})
	return c.provisionAccountCommandOptions
}

// viperProvisionCommandOptions holds the options for any 'provision' subcommands.
type viperProvisionCommandOptions struct {
	Account viperProvisionAccountCommandOptions `mapstructure:"account"`
}
