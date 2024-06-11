package api

import (
	"encoding/json"
	"time"
)

// S1APIResponse represents a response returned from an S1 API call.
type S1APIResponse struct {
	S1APIErrors
	S1APIPageInfo
	Data json.RawMessage `json:"data"`
}

// S1APIErrors holds details on errors returned by the S1 API.
type S1APIErrors struct {
	Errors []struct {
		Code   uint64 `json:"code"`
		Detail string `json:"detail"`
		Title  string `json:"title"`
	} `json:"errors"`
}

// S1APIPage holds details about the current page of results in a response.
type S1APIPageInfo struct {
	Pagination struct {
		NextCursor string `json:"nextCursor"`
		TotalItems uint   `json:"totalItems"`
	} `json:"pagination"`
}

// S1AccountProvisioningRequest holds the body of an account provisioning request.
type S1AccountProvisioningRequest struct {
	AccountName       string   `json:"account_name"`
	AccountType       string   `json:"account_type"`
	Expires           string   `json:"expires"`
	ExternalID        string   `json:"external_id"`
	ReactivateAccount bool     `json:"reactivate_account"`
	Bundle            string   `json:"bundle"`
	TotalAgents       int      `json:"total_agents"`
	Modules           []string `json:"modules"`
}

// S1APIAccountObject represents an account object returned by the S1 API.
type S1APIAccountObject struct {
	ID          string `json:"id"`
	AccountType string `json:"accountType"`
	BillingMode string `json:"billingMode"`
	Expiration  string `json:"expiration"`
	ExternalID  string `json:"externalId"`
	Name        string `json:"name"`
	State       string `json:"state"`
}

// S1Account represents the actual S1 account object.
type S1Account struct {
	ID          string
	AccountType string
	BillingMode string
	Expiration  time.Time
	ExternalID  string
	Name        string
	State       string
}

// S1UserProvisioningRequest holds the body of a user provisioning request.
type S1UserProvisioningRequest struct {
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	EmailAddress string `json:"email_address"`
	Role         string `json:"role"`
}

// S1APIUserObject represents a user object returned by the S1 API.
type S1APIUserObject struct {
	ID              string                     `json:"id"`
	EmailAddress    string                     `json:"email"`
	EmailVerified   bool                       `json:"emailVerified"`
	TwoFactorStatus string                     `json:"twoFaStatus"`
	Scope           string                     `json:"scope"`
	ScopeRoles      []S1APIUserScopeRoleObject `json:"scopeRoles"`
}

// S1APIUserScopeRoleObject represents a single scope role for a user returned by the S1 API.
type S1APIUserScopeRoleObject struct {
	ScopeID  string `json:"id"`
	RoleID   string `json:"roleId"`
	RoleName string `json:"roleName"`
}

// S1User represents the actual S1 user object.
type S1User struct {
	ID              string
	EmailAddress    string
	EmailVerified   bool
	TwoFactorStatus string
	Scope           string
	ScopeRoles      []S1UserScopeRole
}

// S1UserScopeRole represents a single scope role for an S1 user.
type S1UserScopeRole struct {
	ScopeID  string `json:"id"`
	RoleID   string `json:"roleId"`
	RoleName string `json:"roleName"`
}

type S1APIRoleObject struct {
	AccountName    string `json:"accountName"`
	ID             string `json:"id"`
	Name           string `json:"name"`
	PredefinedRole bool   `json:"predefinedRole"`
	Scope          string `json:"scope"`
	ScopeID        string `json:"scopeId"`
	UsersInRole    uint64 `json:"usersInRoles"`
}

type S1Role struct {
	AccountName    string `json:"accountName"`
	ID             string `json:"id"`
	Name           string `json:"name"`
	PredefinedRole bool   `json:"predefinedRole"`
	Scope          string `json:"scope"`
	ScopeID        string `json:"scopeId"`
	UsersInRole    uint64 `json:"usersInRoles"`
}

// S1APISuccessResponseData represents the response to an API call that only indicates success or not.
type S1APISuccessResponseData struct {
	Success bool `json:"success"`
}

// S1APIAffectedResponseData represents the response to an API call that only indicates the number of
// affected records.
type S1APIAffectedResponseData struct {
	Affected uint64 `json:"affected"`
}
