package op

import "time"

const (
	ApplicationTypeWeb ApplicationType = iota
	ApplicationTypeUserAgent
	ApplicationTypeNative

	AccessTokenTypeBearer AccessTokenType = iota
	AccessTokenTypeJWT
)

type Client interface {
	GetID() string
	RedirectURIs() []string
	PostLogoutRedirectURIs() []string
	ApplicationType() ApplicationType
	GetAuthMethod() AuthMethod
	LoginURL(string) string
	AccessTokenType() AccessTokenType
	IDTokenLifetime() time.Duration
}

func IsConfidentialType(c Client) bool {
	return c.ApplicationType() == ApplicationTypeWeb
}

type ApplicationType int

type AuthMethod string

type AccessTokenType int
