package build

import (
	"fmt"
	"runtime"
	"strconv"

	"github.com/Masterminds/semver"
)

// ProductInfo holds information about the product.
type ProductInfo struct {
	// Build is the product build number
	Build string

	// CodeName is the product's internal code name.
	CodeName string

	// Command is the base command executed at the CLI.
	Command string

	// Copyright is the product copyright info.
	Copyright string

	// Env holds environment variables which can be used to store product information.
	Env map[string]string

	// IsDeveloperBuild indicates whether or not this build is a non-optimized, debug build for development purposes.
	IsDeveloperBuild bool

	// ShortTitle is the short title of the product.
	ShortTitle string

	// Title is the full title of the product.
	Title string

	// Version contains the current semantic version of the product.
	Version *semver.Version
}

// NewProductInfo constructs a new object based on the product's version and build information.
func NewProductInfo() *ProductInfo {
	// construct build from the Git commit hash
	productBuild := Commit
	if len(Commit) >= 8 {
		productBuild = Commit[0:8]
	}

	// determine if this is a development build
	if IsDevelopment == "" {
		IsDevelopment = "false"
	}
	isDeveloperBuild, err := strconv.ParseBool(IsDevelopment)
	if err != nil {
		// should never happen as this is controlled directly through the build process
		// this is a safety net to disable developer mode should something go wrong
		isDeveloperBuild = false
	}

	// convert app version to a semantic version
	productVersion, err := semver.NewVersion(Version)
	if err != nil {
		// should never happen as we control this through the build process
		// this is a safety net to let the user know that something went wrong
		productVersion, _ = semver.NewVersion("0.0.0-UNKNOWN")
	}

	// construct the environment variables
	env := map[string]string{
		fmt.Sprintf("%sVERSION", AppEnvPrefix):  productVersion.String(),
		fmt.Sprintf("%sBUILD", AppEnvPrefix):    productBuild,
		fmt.Sprintf("%sCODENAME", AppEnvPrefix): CodeName,
		fmt.Sprintf("%sGOOS", AppEnvPrefix):     runtime.GOOS,
		fmt.Sprintf("%sGOARCH", AppEnvPrefix):   runtime.GOARCH,
	}

	return &ProductInfo{
		Build:            productBuild,
		CodeName:         CodeName,
		Command:          AppCommand,
		Copyright:        AppCopyright,
		Env:              env,
		IsDeveloperBuild: isDeveloperBuild,
		ShortTitle:       AppShortTitle,
		Title:            AppTitle,
		Version:          productVersion,
	}
}
