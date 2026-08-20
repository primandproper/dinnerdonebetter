package keys

const (
	idSuffix = ".id"

	// OAuth2ClientIDKey is the standard key for referring to an OAuth2 client's database ID.
	OAuth2ClientIDKey = "oauth2_clients" + idSuffix
	// OAuth2ClientClientIDKey is the standard key for referring to an OAuth2 client's client ID.
	OAuth2ClientClientIDKey = "oauth2_clients.client_id"
)
