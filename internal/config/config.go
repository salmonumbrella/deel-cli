package config

const (
	// AppName is the application name used for keychain storage
	AppName = "deel-cli"

	// BaseURL is the Deel API base URL
	BaseURL = "https://api.letsdeel.com"

	// EnvToken is the environment variable for direct token auth (CI/scripts)
	EnvToken = "DEEL_TOKEN"

	// EnvAccount is the environment variable for default account name
	EnvAccount = "DEEL_ACCOUNT"

	// EnvOutput is the environment variable for output format
	EnvOutput = "DEEL_OUTPUT"

	// EnvColor is the environment variable for color mode
	EnvColor = "DEEL_COLOR"

	// EnvIdempotencyKey is the environment variable for idempotency key header
	EnvIdempotencyKey = "DEEL_IDEMPOTENCY_KEY"
)
