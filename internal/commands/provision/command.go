package provision

import (
	"github.com/spf13/cobra"
	"go.joshhogle.dev/s1cli/internal/app"
	"go.joshhogle.dev/s1cli/internal/commands/provision/account"
)

// Command is the object for executing the actual command.
type Command struct {
	cobra.Command

	// unexported variables
	appState *app.State
}

// NewCommand creates a new Command object.
func NewCommand(state *app.State) *Command {
	cmd := &Command{
		appState: state,
	}
	cmd.Use = "provision"
	cmd.Short = "Provisions accounts and users."
	cmd.Long = `This command is used to provision users and accounts on the SentinelOne platform.`

	// add flags
	state.Config().CommandOptions().Provision().BindFlags(&cmd.Command)

	// add commands
	cmd.AddCommand(&account.NewCommand(state).Command)

	return cmd
}
