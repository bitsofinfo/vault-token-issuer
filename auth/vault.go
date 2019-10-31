package auth

import "net/http"

// Interface for vault authenticators
type VaultAuthenticator interface {
	Auth(credential VaultCredentials) (string, error)
	GetCredentials(req *http.Request) (VaultCredentials, error)
}

// Interface for credentials for vault
type VaultCredentials interface {
	Get() (interface{}, error)
}
