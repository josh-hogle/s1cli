package build

// Application information.
const (
	AppCommand    = "s1cli"
	AppCopyright  = "Copyright (c) 2024 Josh Hogle. All rights reserved."
	AppTitle      = "SentinelOne API Client"
	AppShortTitle = "S1 API Client"
	AppEnvPrefix  = "S1CLI_"
)

var (
	// Commit is the git commit hash.
	Commit string

	// CodeName is the "code name" for the major release of the product.
	CodeName string

	// IsDevelopment is either "true" or "false" to indicate whether or not this is a development build.
	IsDevelopment string

	// Version is the current semver-compatible version of the product.
	Version string
)
