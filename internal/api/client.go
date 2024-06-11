package api

import (
	"encoding/json"
	goerrors "errors"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"go.joshhogle.dev/errorx"
	"go.joshhogle.dev/s1cli/internal/app"
	"go.joshhogle.dev/s1cli/internal/errors"
)

// S1Client is used to interact with the SentinelOne API.
type S1Client struct {
	appState *app.State
	client   *resty.Client
	apiKey   string
	baseURL  string
}

// CreateAccount creates a new Account in SentinelOne if it does not already exist.
func (s *S1Client) CreateAccount(req S1AccountProvisioningRequest) (*S1Account, errorx.Error) {
	logger := s.appState.Logger().With().Str("account_name", req.AccountName).Logger()
	account, errx := s.FindAccount(req.AccountName)
	if errx != nil {
		return nil, errx
	}

	// configure expiration
	var expires time.Time
	dur, err := time.ParseDuration(req.Expires)
	if err == nil {
		expires = time.Now().Add(dur)
	} else {
		expires, err = time.Parse(time.RFC3339, req.Expires)
		if err != nil {
			errx := errors.NewS1ClientError(
				fmt.Sprintf("failed to parse account expiration time and date '%s'", req.Expires), err)
			logger.Error().Err(errx).Str("expiration_date", req.Expires).Msg(errx.Error())
			return nil, errx
		}
	}

	// account exists - if it is expired, either reactivate it or return an error
	if account != nil {
		logger := logger.With().Str("account_id", account.ID).Logger()
		switch account.State {
		case "active":
			logger.Info().Str("expires", account.Expiration.String()).Msg("found existing active account")
			return account, nil
		case "expired":
			if !req.ReactivateAccount {
				errx := errors.NewS1ClientError(
					"failed to create account because it is expired and not set to be reactivated",
					goerrors.New("account already exists"))
				logger.Error().Err(errx).Msg(errx.Error())
				return nil, errx
			}
			if errx := s.ReactivateAccount(account.ID, expires); errx != nil {
				return nil, errx
			}
			return account, nil
		default:
			errx := errors.NewS1ClientError(
				fmt.Sprintf("failed to create account because it exists and is currently '%s'", account.State),
				goerrors.New("account already exists"))
			logger.Error().Err(errx).Msg(errx.Error())
			return nil, errx
		}
	}

	// create the new account, good until configured duration expires
	logger.Info().Msg("creating new account")
	modules := []map[string]any{}
	for _, module := range req.Modules {
		modules = append(modules, map[string]any{"name": module})
	}
	body := map[string]any{
		"data": map[string]any{
			"name":        req.AccountName,
			"accountType": req.AccountType,
			"billingMode": "subscription",
			"expiration":  expires.Format(time.RFC3339),
			"externalId":  req.ExternalID,
			"inherits":    true,
			"licenses": map[string]any{
				"bundles": []map[string]any{
					{
						"name": req.Bundle,
						"surfaces": []map[string]any{
							{
								"count": req.TotalAgents,
								"name":  "Total Agents",
							},
						},
					},
				},
				"modules": modules,
				"settings": []map[string]any{
					{
						"groupName": "dv_retention",
						"setting":   "30 Days",
					},
					{
						"groupName": "malicious_data_retention",
						"setting":   "365 Days",
					},
					{
						"groupName": "remote_shell_availability",
						"setting":   "Enabled",
					},
					{
						"groupName": "marketplace_access_status",
						"setting":   "Available",
					},
					{
						"groupName": "account_level_ranger",
						"setting":   "Account",
					},
				},
			},
			"unlimitedExpiration": false,
			"usageType":           "customer",
		},
	}
	resp, errx := s.exec(http.MethodPost, "/accounts", withRequestBody(body))
	if errx != nil {
		return nil, errx
	}

	// parse the response
	var newAcct S1APIAccountObject
	if err := json.Unmarshal(resp.Data, &newAcct); err != nil {
		errx := errors.NewS1ClientError("failed to unmarshal response from server", err)
		logger.Error().Err(errx).Msg(errx.Error())
		return nil, errx
	}
	return s.fromS1APIAccountObject(newAcct)
}

// CreateUser creates a new User in SentinelOne if it does not already exist.
func (s *S1Client) CreateUser(req *S1UserProvisioningRequest, accountID string) (*S1User, errorx.Error) {
	logger := s.appState.Logger().With().Str("email_address", req.EmailAddress).Logger()
	user, e := s.FindUser(req.EmailAddress)
	if e != nil {
		return nil, e
	}
	adminRole, e := s.FindRole(accountID, "Admin")
	if e != nil {
		return nil, e
	}

	// user exists - add the user as an Admin to the account (if they aren't already)
	if user != nil {
		for _, role := range user.ScopeRoles {
			if role.ScopeID == accountID {
				logger.Info().Str("user_id", user.ID).Msg("found existing user")
				return user, nil
			}
		}

		// add the user as an Admin to the account
		user.ScopeRoles = append(user.ScopeRoles, S1UserScopeRole{
			ScopeID:  accountID,
			RoleID:   adminRole.ID,
			RoleName: adminRole.Name,
		})
		user, err := s.UpdateUserScopeRoles(user.ID, user.ScopeRoles)
		if err != nil {
			return nil, err
		}
		return user, nil
	}

	// generate random password
	//rand.Seed(time.Now().UnixNano()) // not required as of Go 1.20
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")
	length := 32
	var b strings.Builder
	for i := 0; i < length; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	passwd := b.String()

	// create the new user
	logger.Info().Msg("creating new user")
	body := map[string]any{
		"data": map[string]any{
			"email":    req.EmailAddress,
			"password": passwd,
			"fullName": fmt.Sprintf("%s %s", req.FirstName, req.LastName),
			"scope":    "account",
			"scopeRoles": []map[string]any{
				{
					"id":       accountID,
					"roleName": req.Role,
				},
			},
			"twoFaEnabled": true,
		},
	}
	resp, err := s.exec(http.MethodPost, "/users", withRequestBody(body))
	if err != nil {
		return nil, err
	}

	// parse the response
	var newUser S1APIUserObject
	if err := json.Unmarshal(resp.Data, &newUser); err != nil {
		errx := errors.NewS1ClientError("failed to unmarshal response from server", err)
		logger.Error().Err(errx).Msg(errx.Error())
		return nil, errx
	}
	return s.fromS1APIUserObject(newUser)
}

/*
// DeleteUser deletes an S1 user.
func (s *S1ClientService) DeleteUser(userID string) *Error {
	InfoMsg(nil, "deleting user ID '%s'", userID)

	body := map[string]any{
		"filter": map[string]any{
			"ids": []string{userID},
		},
	}
	resp, err := s.exec(http.MethodPost, "/users/delete-users", withRequestBody(body))
	if err != nil {
		return err
	}

	// parse the response
	var data S1APIAffectedResponseData
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		ErrorMsg(nil, "failed to unmarshal response from server: %s", err.Error())
		return NewError(ERR_DELETE_USER_FAILED, err)
	}

	// make sure the user was deleted
	if data.Affected == 0 {
		ErrorMsg(nil, "did not delete any users")
		return NewError(ERR_DELETE_USER_FAILED, errors.New("did not delete any users"))
	}
	return nil
}
*/

// FindAccount searches for the matching account with the given name.
//
// If the account cannot be found, no error will be returned but the account object will be nil.
func (s *S1Client) FindAccount(name string) (*S1Account, errorx.Error) {
	logger := s.appState.Logger()
	logger.Debug().Str("account_name", name).Msgf("searching for account")

	// search for the account
	// -- this should never return more than 1 account as account names must be unique
	resp, err := s.exec(http.MethodGet, "/accounts", withRequestParams(map[string]string{
		"name":  name,
		"limit": "1",
	}))
	if err != nil {
		return nil, err
	}

	// parse the data
	var apiAccounts []S1APIAccountObject
	if err := json.Unmarshal(resp.Data, &apiAccounts); err != nil {
		errx := errors.NewS1ClientError("failed to unmarshal response from server", err)
		logger.Error().Err(errx).Msg(errx.Error())
		return nil, errx
	}

	// convert the response object
	if len(apiAccounts) == 0 {
		return nil, nil
	}
	return s.fromS1APIAccountObject(apiAccounts[0])
}

// FindRole searches for matching roles in the given account with the given name.
//
// If the role cannot be found, no error will be returned but the role object will be nil.
func (s *S1Client) FindRole(accountID, name string) (*S1Role, errorx.Error) {
	logger := s.appState.Logger().With().Str("account_id", accountID).Str("role", name).Logger()
	logger.Debug().Msg("searching for role in account")

	// search for the role
	// -- this should never return more than 1 role as role names must be unique
	resp, err := s.exec(http.MethodGet, "/rbac/roles", withRequestParams(map[string]string{
		"accountIds": accountID,
		"name":       name,
		"limit":      "1",
	}))
	if err != nil {
		return nil, err
	}

	// parse the data
	var apiRoles []S1APIRoleObject
	if err := json.Unmarshal(resp.Data, &apiRoles); err != nil {
		errx := errors.NewS1ClientError("failed to unmarshal response from server", err)
		logger.Error().Err(errx).Msg(errx.Error())
		return nil, errx
	}

	// convert the response object
	if len(apiRoles) == 0 {
		return nil, nil
	}
	return s.fromS1APIRoleObject(apiRoles[0])
}

// FindUser searches for matching users with the given email address.
//
// If the user cannot be found, no error will be returned but the user object will be nil.
func (s *S1Client) FindUser(email string) (*S1User, errorx.Error) {
	logger := s.appState.Logger().With().Str("email_address", email).Logger()
	logger.Debug().Msg("searching for user")

	// search for the user
	// -- this should never return more than 1 user as e-mail addresses must be unique
	resp, err := s.exec(http.MethodGet, "/users", withRequestParams(map[string]string{
		"email": email,
		"limit": "1",
	}))
	if err != nil {
		return nil, err
	}

	// parse the data
	var apiUsers []S1APIUserObject
	if err := json.Unmarshal(resp.Data, &apiUsers); err != nil {
		errx := errors.NewS1ClientError("failed to unmarshal response from server", err)
		logger.Error().Err(errx).Msg(errx.Error())
		return nil, errx
	}

	// convert the response object
	if len(apiUsers) == 0 {
		return nil, nil
	}
	return s.fromS1APIUserObject(apiUsers[0])
}

// ReactivateAccount reactivates an expired account and extends its expiration by the configured duration.
func (s *S1Client) ReactivateAccount(id string, expires time.Time) errorx.Error {
	logger := s.appState.Logger().With().Str("account_id", id).Logger()
	logger.Info().Msg("reactivating account")

	body := map[string]any{
		"data": map[string]any{
			"unlimited":  false,
			"expiration": expires.Format(time.RFC3339),
		},
	}
	resp, err := s.exec(http.MethodPut, fmt.Sprintf("/accounts/%s/reactivate", id),
		withRequestBody(body))
	if err != nil {
		return err
	}

	// parse the response
	var data S1APISuccessResponseData
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		errx := errors.NewS1ClientError("failed to unmarshal response from server", err)
		logger.Error().Err(errx).Msg(errx.Error())
		return errx
	}

	// make sure account was reactivated
	if !data.Success {
		errx := errors.NewS1ClientError("failed to reactivate account", goerrors.New("activation was not successful"))
		logger.Error().Err(errx).Msg(errx.Error())
		return errx
	}
	return nil
}

// ResetUserPassword triggers a password reset email to be sent to the given user.
func (s *S1Client) ResetUserPassword(userID string) errorx.Error {
	logger := s.appState.Logger().With().Str("user_id", userID).Logger()
	logger.Info().Msg("resetting user password")

	body := map[string]any{
		"filter": map[string]any{
			"ids": []string{userID},
		},
	}
	resp, err := s.exec(http.MethodPost, "/users/login/send-reset-password-email", withRequestBody(body))
	if err != nil {
		return err
	}

	// parse the response
	var data S1APIAffectedResponseData
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		errx := errors.NewS1ClientError("failed to unmarshal response from server", err)
		logger.Error().Err(errx).Msg(errx.Error())
		return errx
	}

	// make sure 1 user was affected
	if data.Affected == 0 {
		errx := errors.NewS1ClientError("failed to reset password for user", goerrors.New("user ID was not found"))
		logger.Error().Err(errx).Msg(errx.Error())
		return errx
	}
	return nil
}

// UpdateUserScopeRoles updates the scope roles for the given user.
func (s *S1Client) UpdateUserScopeRoles(userID string, roles []S1UserScopeRole) (*S1User, errorx.Error) {
	logger := s.appState.Logger().With().Str("user_id", userID).Logger()
	logger.Debug().Msg("updating scope roles for user")

	body := map[string]any{
		"data": map[string]any{
			"scope":      "account",
			"scopeRoles": roles,
		},
	}
	resp, err := s.exec(http.MethodPut, fmt.Sprintf("/users/%s", userID), withRequestBody(body))
	if err != nil {
		return nil, err
	}

	// parse the response
	var user S1APIUserObject
	if err := json.Unmarshal(resp.Data, &user); err != nil {
		errx := errors.NewS1ClientError("failed to unmarshal response from server", err)
		logger.Error().Err(errx).Msg(errx.Error())
		return nil, errx
	}

	// convert the response object
	return s.fromS1APIUserObject(user)
}

// exec executes a call to the S1 REST API.
func (s *S1Client) exec(method, endpoint string, optFns ...s1ClientExecOptFn) (*S1APIResponse, errorx.Error) {
	url := fmt.Sprintf("%s/web/api/v2.1%s", s.baseURL, endpoint)
	logger := s.appState.Logger().With().Str("url", url).Str("method", method).Logger()

	req := s.client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetHeader("Authorization", fmt.Sprintf("ApiToken %s", s.apiKey))
	for _, fn := range optFns {
		req = fn(req)
	}
	resp, err := req.Execute(method, url)
	if err != nil {
		errx := errors.NewS1ClientRequestError(method, url, "failed to execute request", err)
		logger.Error().Err(errx).Msg(errx.Error())
		return nil, errx
	}

	// check response status code
	httpCode := resp.StatusCode()
	if httpCode >= http.StatusMethodNotAllowed {
		errx := errors.NewS1ClientRequestError(method, url, "failed to execute request",
			goerrors.New("method is not allowed for endpoint"))
		logger.Error().Err(errx).Msg(errx.Error())
		return nil, errx
	}
	if httpCode >= http.StatusInternalServerError {
		errx := errors.NewS1ClientRequestError(method, url, "failed to execute request",
			fmt.Errorf("request returned server error code %d", httpCode))
		logger.Error().Err(errx).Int("status_code", httpCode).Msg(errx.Error())
		return nil, errx
	}

	// parse the response from the call
	var apiResponse S1APIResponse
	if err := json.Unmarshal(resp.Body(), &apiResponse); err != nil {
		errx := errors.NewS1ClientRequestError(method, url, "failed to unmarshal response from request", err)
		logger.Error().Err(errx).Msg(errx.Error())
		return nil, errx
	}

	// check for errors
	if len(apiResponse.Errors) > 0 {
		for _, e := range apiResponse.Errors {
			if e.Detail != "" {
				logger.Error().Err(fmt.Errorf("%s: %s", e.Title, e.Detail)).Uint64("error_code", e.Code).
					Msgf("%s: %s", e.Title, e.Detail)
			} else {
				logger.Error().Err(goerrors.New(e.Title)).Uint64("error_code", e.Code).Msg(e.Title)
			}
		}
		return nil, errors.NewS1ClientRequestError(method, url, "server returned one or more API errors",
			goerrors.New("server returned one or more API errors"))
	}
	return &apiResponse, nil
}

// fromS1APIAccountObject converts an account object returned by the API to an actual S1 account object.
func (s *S1Client) fromS1APIAccountObject(o S1APIAccountObject) (*S1Account, errorx.Error) {
	logger := s.appState.Logger()
	expires, err := time.Parse(time.RFC3339, o.Expiration)
	if err != nil {
		errx := errors.NewS1ClientError("failed to parse account expiration date", err)
		logger.Error().Err(errx).Str("expires", o.Expiration).Msg(errx.Error())
		return nil, errx
	}
	return &S1Account{
		ID:          o.ID,
		AccountType: o.AccountType,
		BillingMode: o.BillingMode,
		Expiration:  expires,
		ExternalID:  o.ExternalID,
		Name:        o.Name,
		State:       o.State,
	}, nil
}

// fromS1APIRoleObject converts a role object returned by the API to an actual S1 role object.
func (s *S1Client) fromS1APIRoleObject(o S1APIRoleObject) (*S1Role, errorx.Error) {
	role := S1Role(o)
	return &role, nil
}

// fromS1APIUserObject converts a user object returned by the API to an actual S1 user object.
func (s *S1Client) fromS1APIUserObject(o S1APIUserObject) (*S1User, errorx.Error) {
	user := &S1User{
		ID:              o.ID,
		EmailAddress:    o.EmailAddress,
		EmailVerified:   o.EmailVerified,
		TwoFactorStatus: o.TwoFactorStatus,
		Scope:           o.Scope,
		ScopeRoles:      []S1UserScopeRole{},
	}
	for _, role := range o.ScopeRoles {
		user.ScopeRoles = append(user.ScopeRoles, S1UserScopeRole(role))
	}
	return user, nil
}

/*
// formatAccountName replaces all placeholders in the account name format and returns the result.
func (s *S1ClientService) formatAccountName(req *ProvisioningRequest) string {
	// TODO: this can be greatly improved using regexp
	name := strings.ReplaceAll(s.accountNameFormat, "{request_id}", req.RequestID)
	name = strings.ReplaceAll(name, "{workshop_id}", req.WorkshopID)
	name = strings.ReplaceAll(name, "{first_name}", req.FirstName)
	name = strings.ReplaceAll(name, "{last_name}", req.LastName)
	name = strings.ReplaceAll(name, "{company}", req.Company)
	name = strings.ReplaceAll(name, "{title}", req.Title)
	return strings.ReplaceAll(name, "{address}", req.EmailAddress)
}
*/

// s1ClientExecOptFn is used to pass optional settings to the exec() call.
type s1ClientExecOptFn func(*resty.Request) *resty.Request

// withRequestHeaders adds headers to the REST request.
func withRequestHeaders(headers map[string]string) s1ClientExecOptFn {
	return func(r *resty.Request) *resty.Request {
		return r.SetHeaders(headers)
	}
}

// withRequestBody adds a JSON body to the REST request.
func withRequestBody(body map[string]any) s1ClientExecOptFn {
	return func(r *resty.Request) *resty.Request {
		return r.SetBody(body)
	}
}

// withRequestParams adds query parameters to the REST request.
func withRequestParams(params map[string]string) s1ClientExecOptFn {
	return func(r *resty.Request) *resty.Request {
		return r.SetQueryParams(params)
	}
}

// s1ClientBuilder is used to configure the S1 client.
type s1ClientBuilder struct {
	cli *S1Client
}

// NewS1ClientBuilder creates a new s1ClientBuilder object.
func NewS1ClientBuilder(state *app.State, baseURL, apiKey string) *s1ClientBuilder {
	// TODO: check state is not nil
	return &s1ClientBuilder{
		cli: &S1Client{
			appState: state,
			client:   resty.New(),
			baseURL:  baseURL,
			apiKey:   apiKey,
		},
	}
}

// Build finishes the build and returns the configured S1Client object.
func (b *s1ClientBuilder) Build() *S1Client {
	return b.cli
}

// WithHTTPClient sets the customized HTTP client to use when accessing the S1 API.
func (b *s1ClientBuilder) WithHTTPClient(client *resty.Client) *s1ClientBuilder {
	if client != nil {
		b.cli.client = client
	}
	return b
}
