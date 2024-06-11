package errors

import (
	"fmt"

	"go.joshhogle.dev/errorx"
)

// configBaseError is the base object used by configuration errors.
type configBaseError struct {
	*errorx.BaseError

	// unexported variables
	configFile string
}

// newConfigBaseError creates a new configBaseError error.
func newConfigBaseError(configFile string, code int, err error) *configBaseError {
	e := &configBaseError{
		BaseError:  errorx.NewBaseError(code, err),
		configFile: configFile,
	}
	e.WithAttrs(map[string]any{
		"config_file": configFile,
	})
	return e
}

// ConfigFile returns the name of the configuration file.
func (e *configBaseError) ConfigFile() string {
	return e.configFile
}

// ConfigLoadFailure occurs when an error is detected while loading the configuration file.
type ConfigLoadFailure struct {
	*configBaseError
}

// NewConfigLoadFailure returns a new ConfigLoadFailure error.
func NewConfigLoadFailure(configFile string, err error) *ConfigLoadFailure {
	return &ConfigLoadFailure{
		configBaseError: newConfigBaseError(configFile, ConfigLoadFailureCode, err),
	}
}

// Error returns the string version of the error.
func (e *ConfigLoadFailure) Error() string {
	return fmt.Sprintf("error while loading configuration file '%s': %s", e.configFile, e.InternalError().Error())
}

// ConfigParseFailure occurs when an error is detected while parsing configuration settings.
type ConfigParseFailure struct {
	*configBaseError
}

// NewConfigParseFailure returns a new ConfigParseFailure error.
func NewConfigParseFailure(configFile string, err error) *ConfigParseFailure {
	return &ConfigParseFailure{
		configBaseError: newConfigBaseError(configFile, ConfigParseFailureCode, err),
	}
}

// Error returns the string version of the error.
func (e *ConfigParseFailure) Error() string {
	return fmt.Sprintf("error while parsing configuration file '%s': %s", e.configFile, e.InternalError().Error())
}

// ConfigValidateFailure occurs when an error is detected while validating configuration settings.
type ConfigValidateFailure struct {
	*configBaseError

	// unexported variables
	setting string
	value   any
}

// NewConfigValidateFailure returns a new ConfigValidateFailure error.
func NewConfigValidateFailure(configFile, setting string, val any, err error) *ConfigValidateFailure {
	e := &ConfigValidateFailure{
		configBaseError: newConfigBaseError(configFile, ConfigValidateFailureCode, err),
		setting:         setting,
		value:           val,
	}
	e.WithAttrs(map[string]any{
		"setting": setting,
		"value":   val,
	})
	return e
}

// Error returns the string version of the error.
func (e *ConfigValidateFailure) Error() string {
	if e.setting != "" {
		return fmt.Sprintf("the configuration setting '%s' is invalid: %s", e.setting, e.InternalError().Error())
	}
	return fmt.Sprintf("one or more configuration settings are invalid: %s", e.InternalError().Error())
}

// Setting returns the name of the setting that was invalid.
func (e *ConfigValidateFailure) Setting() string {
	return e.setting
}

// Value returns the value of the setting that was invalid.
func (e *ConfigValidateFailure) Value() any {
	return e.value
}
