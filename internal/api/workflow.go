package api

/*
// ProvisioningWorkflow is responsible for the overall workflow when provisioning a user.
type ProvisioningWorkflow struct {
	s1Client *S1ClientService
}

// NewProvisioningWorkflow creates a new ProvisioningWorkflow object.
//
// Always defer a call to Cleanup() after creating the object.
func NewProvisioningWorkflow(svc *S1ClientService) *ProvisioningWorkflow {
	return &ProvisioningWorkflow{
		s1Client: svc,
	}
}

// Cleanup is responsible for cleaning up open DB handles, etc.
func (w *ProvisioningWorkflow) Cleanup() {
}

/*
// ProvisionUser is responsible for actually creating an S1 account and user.
func (w *ProvisioningWorkflow) ProvisionUser(req *ProvisioningRequest) (*S1Account, *S1User, *Error) {
	// create the S1 account
	s1Acct, err := w.s1Client.CreateAccount(req)
	if err != nil {
		return nil, nil, err
	}

	// create the S1 user and send a password reset email
	s1User, err := w.s1Client.CreateUser(req, s1Acct.ID)
	if err != nil {
		return nil, nil, err
	}
	if err := w.s1Client.ResetUserPassword(s1User.ID); err != nil {
		return nil, nil, err
	}

	// update the database with the S1 account and user ID (for deletion later); set to provisioned
	return s1Acct, s1User, nil
}

// ResetUserPassword resets the user's password associated with the given registration.
func (w *ProvisioningWorkflow) ResetUserPassword(req *ProvisioningRequest) *Error {
	user, err := w.s1Client.FindUser(req.EmailAddress)
	if err != nil {
		return err
	}
	if user == nil {
		return NewError(404, fmt.Errorf("unable to find user with e-mail address '%s'", req.EmailAddress))
	}
	return w.s1Client.ResetUserPassword(user.ID)
}

// ValidateRequest ensures the request body passed in is valid.
func (w *ProvisioningWorkflow) ValidateRequest(req *ProvisioningRequest) *Error {
	if _, err := mail.ParseAddress(req.EmailAddress); err != nil {
		ErrorMsg(req, "email address '%s' is invalid: %s", req.EmailAddress, err.Error())
		return NewError(ERR_VALIDATE_REQUEST, err)
	}
	if req.FirstName == "" || req.LastName == "" || req.Company == "" || req.Title == "" {
		ErrorMsg(req, "one or more required request parameters is empty")
		return NewError(ERR_VALIDATE_REQUEST, errors.New("missing required request parameters"))
	}
	return nil
}
*/
