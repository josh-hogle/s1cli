package errors

// Object error codes (0-1000)
const (
	// general errors (0-20)
	NoneCode           = 0
	UsageErrorCode     = 1
	GeneralFailureCode = 2

	// configuration errors (21-40)
	ConfigLoadFailureCode     = 21
	ConfigParseFailureCode    = 22
	ConfigValidateFailureCode = 23

	// S1 client errors (101-120)
	S1ClientErrorCode        = 101
	S1ClientRequestErrorCode = 102

	/*
		// HTTP service errors (41-60)
		HTTPServiceFailureCode        = 41
		HTTPServiceShutdownForcedCode = 42

		// TLS errors (61-70)
		TLSCertificateLoadFailureCode = 61
		TLSPrivateKeyLoadFailureCode  = 62

		// context errors (71-75)
		ContextKeyNotFoundCode            = 71
		ContextValueConversionFailureCode = 72

		// API errors (101-200)
		APIBadRequestCode       = 101
		APITokenExpiredCode     = 102
		APITokenInvalidCode     = 103
		APIGeneralFailureCode   = 104
		APIResourceNotFoundCode = 105

		// provider errors (201-220)
		ProviderConversionFailureCode = 201
		ProviderNotSupportedCode      = 202
		ProviderNotFoundCode          = 203
		LogProviderWriteErrorCode     = 204
	*/
)
