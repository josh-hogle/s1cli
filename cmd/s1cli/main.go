package main

import (
	"io"
	"log"
	"os"

	"go.joshhogle.dev/errorx"
	"go.joshhogle.dev/s1cli/internal/app"
	"go.joshhogle.dev/s1cli/internal/errors"
)

func main() {
	os.Exit(run())
}

// run is responsible for actually running the main application thread.
//
// We use this function in order to defer state cleanup. os.Exit() will bypass the deferred function
// so simply moving the functionality inside of its own function easily solves the issue.
func run() int {
	// initialize application state
	log.SetOutput(io.Discard) // discard standard logger output
	appState := app.NewState()
	defer appState.Cleanup()

	// execute the command
	var exitCode int
	err := NewRootCommand(appState).Execute()
	if e, ok := err.(errorx.Error); ok {
		// the extended error message should already have been logged during execution
		exitCode = e.Code()
	} else if err != nil {
		// error returned was not an "extended" error so treat it as a usage error
		errx := errors.NewUsageError(err)
		appState.Logger().Error().Err(errx).Msg(errx.Error())
		exitCode = errx.Code()
	}
	if exitCode != 0 && exitCode != errors.UsageErrorCode {
		appState.Logger().Warn().Int("exit_code", exitCode).Msgf("exiting with non-zero exit code: %d", exitCode)
	}
	return exitCode
}
