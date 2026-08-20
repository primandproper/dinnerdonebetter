package oauth

// Repository reads and writes the OAuth2 client registry.
//
// It no longer covers tokens. Authorization codes, access tokens and refresh tokens are
// the platform authorization server's records, held in its own Store — see
// internal/services/auth/handlers/authentication for how the two are composed.
type Repository interface {
	OAuth2ClientDataManager
}
