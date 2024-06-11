package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.joshhogle.dev/s1cli/internal/app"
	"go.joshhogle.dev/s1cli/internal/build"
	"go.joshhogle.dev/s1cli/internal/commands/provision"
	"go.joshhogle.dev/s1cli/internal/commands/version"
)

// RootCommand is the base command for the application.
type RootCommand struct {
	cobra.Command

	// unexported variables
	appState *app.State
}

// NewRootCommand creates a new Command object.
func NewRootCommand(state *app.State) *RootCommand {
	cmd := &RootCommand{
		appState: state,
	}
	cmd.Use = build.AppCommand
	cmd.Short = fmt.Sprintf("Runs %s", build.AppShortTitle)
	cmd.Long = fmt.Sprintf("Runs the %s application", build.AppTitle)

	// add flags
	state.Config().GlobalOptions().BindFlags(&cmd.Command)

	// add commands
	cmd.AddCommand(&provision.NewCommand(state).Command)
	cmd.AddCommand(&version.NewCommand(state).Command)

	return cmd
}
