package version

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"go.joshhogle.dev/s1cli/internal/app"
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
	cmd.Use = "version"
	cmd.Short = "Display application version information"
	cmd.Long = "This command allows you to see simple or detailed version information."
	cmd.RunE = cmd.runE

	// add flags
	state.Config().CommandOptions().Version().BindFlags(&cmd.Command)
	return cmd
}

// run simply executes the command.
func (c *Command) runE(cmd *cobra.Command, args []string) error {
	if err := c.appState.Initialize(&c.Command); err != nil {
		return err
	}
	cmdOpts := c.appState.Config().CommandOptions().Version()
	if err := cmdOpts.Load(); err != nil {
		return err
	}
	productInfo := c.appState.ProductInfo()
	c.appState.DisableLogger(true)

	// show just the version
	if cmdOpts.Short {
		fmt.Printf("%s\n", productInfo.Version.String())
		return nil
	}

	// show version and build
	if !cmdOpts.Verbose {
		fmt.Printf("%s build %s (%s)", productInfo.Version.String(), productInfo.Build, productInfo.CodeName)
		if productInfo.IsDeveloperBuild {
			fmt.Printf(" [Developer Build]")
		}
		fmt.Printf("\n")
		return nil
	}

	// get main version info
	buf := bytes.NewBufferString("\n")
	title := fmt.Sprintf("%s Version Information", productInfo.Title)
	titleLen := len(title)
	titleCenterIndent := ((75 - titleLen) / 2) + titleLen
	rightBarIndent := 75 - titleCenterIndent - 1
	fmt.Fprintf(buf, "===========================================================================\n"+
		"|%*s%*s\n"+
		"===========================================================================\n", titleCenterIndent, title,
		rightBarIndent, "|")
	fmt.Fprintf(buf, "%30s : %s build %s\n", fmt.Sprintf("%s CLI version", productInfo.ShortTitle),
		productInfo.Version.String(), productInfo.Build)
	fmt.Fprintf(buf, "%30s : %s\n", "Code Name", productInfo.CodeName)
	fmt.Fprintf(buf, "%30s : %s\n", "Developer Build", strconv.FormatBool(productInfo.IsDeveloperBuild))
	fmt.Fprintf(buf, "\n")

	// show the output
	fmt.Printf("%s\n", buf.String())
	return nil
}
