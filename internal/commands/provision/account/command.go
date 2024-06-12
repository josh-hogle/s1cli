package account

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/jszwec/csvutil"
	"github.com/spf13/cobra"
	"go.joshhogle.dev/errorx"
	"go.joshhogle.dev/s1cli/internal/api"
	"go.joshhogle.dev/s1cli/internal/app"
	"go.joshhogle.dev/s1cli/internal/errors"
)

// Command is the object for executing the actual command.
type Command struct {
	cobra.Command

	// unexported variables
	appState *app.State
	s1Client *api.S1Client
}

// accountDetails holds the details for provisioning the account.
type accountDetails struct {
	AccountName  string `csv:"account_name"`
	AccountType  string `csv:"account_type"`
	Expires      string `csv:"expires"`
	ExternalID   string `csv:"external_id"`
	Bundle       string `csv:"bundle"`
	TotalAgents  int    `csv:"total_agents"`
	Modules      string `csv:"modules"`
	FirstName    string `csv:"first_name"`
	LastName     string `csv:"last_name"`
	EmailAddress string `csv:"email_address"`
	Role         string `csv:"role"`
}

// NewCommand creates a new Command object.
func NewCommand(state *app.State) *Command {
	cmd := &Command{
		appState: state,
	}
	cmd.Use = "account"
	cmd.Short = "Provisions accounts."
	cmd.Long = `This command is used to provision accounts on the SentinelOne platform.`
	cmd.RunE = cmd.runE

	// add flags
	state.Config().CommandOptions().Provision().Account().BindFlags(&cmd.Command)

	return cmd
}

// run simply executes the command.
func (c *Command) runE(cmd *cobra.Command, args []string) error {
	if err := c.appState.Initialize(&c.Command); err != nil {
		return err
	}
	cmdOpts := c.appState.Config().CommandOptions().Provision().Account()
	if err := cmdOpts.Load(); err != nil {
		return err
	}
	cmdOpts.LogSettings(true)
	logger := c.appState.Logger()

	// TODO: check API key and tenant URL
	globalOpts := c.appState.Config().GlobalOptions()
	c.s1Client = api.NewS1ClientBuilder(c.appState, globalOpts.TenantURL, globalOpts.APIKey).Build()

	if cmdOpts.CSVSource == "" {
		// TODO: if no CSV has been provided, prompt for the information to provision the account
		fmt.Printf("\n\n-- Only CSV provisioning is supported at this time --\n\n")
		return nil
	}

	// open the CSV
	f, err := os.Open(cmdOpts.CSVSource)
	if err != nil {
		errx := errors.NewGeneralFailure(
			fmt.Sprintf("failed to open CSV file '%s' for reading", cmdOpts.CSVSource), err)
		logger.Error().Err(errx).Str("csv_file", cmdOpts.CSVSource).Msg(errx.Error())
		return errx
	}

	// read the CSV
	csvReader := csv.NewReader(f)
	csvReader.Comma = rune(cmdOpts.CSVSeparator[0])
	dec, err := csvutil.NewDecoder(csvReader)
	if err != nil {
		errx := errors.NewGeneralFailure(
			fmt.Sprintf("failed to parse CSV file '%s'", cmdOpts.CSVSource), err)
		logger.Error().Err(errx).Str("csv_file", cmdOpts.CSVSource).Msg(errx.Error())
		return errx
	}

	// provision the list of accounts
	for {
		var account accountDetails
		if err := dec.Decode(&account); err == io.EOF {
			logger.Info().Msg("all accounts have been provisioned")
			break
		} else if err != nil {
			errx := errors.NewGeneralFailure("failed to decode account record", err)
			logger.Error().Err(errx).Str("csv_file", cmdOpts.CSVSource).Msg(errx.Error())
			return errx
		}

		if err := c.provisionAccount(account, cmdOpts.ReactivateExpiredAccount,
			cmdOpts.ResetFirstUserPassword); err != nil {
			return err
		}
	}
	return nil
}

func (c *Command) provisionAccount(account accountDetails, reactivate, resetFirstUserPass bool) errorx.Error {
	// TODO: add checks for request values

	// create the account
	acct, errx := c.s1Client.CreateAccount(api.S1AccountProvisioningRequest{
		AccountName:       account.AccountName,
		AccountType:       account.AccountType,
		Expires:           account.Expires,
		ExternalID:        account.ExternalID,
		ReactivateAccount: reactivate,
		Bundle:            account.Bundle,
		Modules:           strings.Split(account.Modules, ","),
		TotalAgents:       account.TotalAgents,
	})
	if errx != nil {
		return errx
	}
	logger := c.appState.Logger().With().Str("account_id", acct.ID).Str("account_name", acct.Name).Logger()
	logger.Info().Msg("account has been successfully provisioned")

	// create the user
	user, errx := c.s1Client.CreateUser(&api.S1UserProvisioningRequest{
		FirstName:    account.FirstName,
		LastName:     account.LastName,
		EmailAddress: account.EmailAddress,
		Role:         account.Role,
	}, acct.ID)
	if errx != nil {
		return errx
	}
	logger = logger.With().Str("user_id", user.ID).Str("email_address", user.EmailAddress).Logger()
	logger.Info().Msg("user has been created and enabled for account")

	// reset the user's password
	if resetFirstUserPass {
		if errx := c.s1Client.ResetUserPassword(user.ID); errx != nil {
			return errx
		}
	}

	return nil
}
