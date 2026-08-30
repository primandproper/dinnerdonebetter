package audit

import (
	platformaudit "github.com/primandproper/platform-go/v13/audit"
)

// Redaction declares what happens to a resource type's named fields on the way
// into the log.
type Redaction = platformaudit.Redaction

// Redactions is this application's policy for what never becomes durable in the
// audit log, keyed by resource type. The empty key applies to every resource
// type.
//
// It is declared here, in code, rather than in a config file, because it is a
// property of the domain rather than of a deployment: a bearer token is a bearer
// token in every environment, and a policy that could be relaxed by an
// environment variable is a policy one bad rollout away from writing secrets into
// the one table designed to be immutable and kept for years. Filtering at query
// time is not the same thing and does not help — by then the value is written.
//
// The catch-all below is the important half. It is a rule about the word, not
// about one table: a field named "password" is a password wherever it shows up,
// including in a Diff of a struct nobody thought about when they added it. A
// resource type's own rules can only add to the catch-all; there is no way to opt
// back out of it, which is deliberate, since that is the shape of every redaction
// bug worth having.
//
// Hash rather than Drop where the audit question is "did this change, and is it
// the same value as that one" — rotating a credential is a real event worth
// recording, and the new credential is not a thing to write down.
var Redactions = map[string]Redaction{
	"": {
		Drop: []string{
			"password",
			"hashed_password",
			"two_factor_secret",
			"twoFactorSecret",
			"totp_token",
			"totpToken",
			"client_secret",
			"clientSecret",
			"access_token",
			"accessToken",
			"refresh_token",
			"refreshToken",
			"secret",
		},
		Hash: []string{
			"token",
			"email_verification_token",
			"emailVerificationToken",
			"api_key",
			"apiKey",
		},
	},
	// A password reset token's whole value is that it is unguessable, and this table
	// outlives the token by years. The reset store itself no longer has a plaintext
	// token to hand anybody — it keeps a digest — so this is a guard rather than a
	// filter that fires: it is what happens if a "token" field ever reappears in a
	// Diff of a struct nobody thought about. The digest still answers the question an
	// investigation asks — whether the token presented was the token issued.
	"password_reset_tokens": {
		Hash: []string{"token"},
	},
	// The address and phone number a user reports an issue with are theirs, not the
	// report's, and re-recording them here would put a copy outside every erasure
	// path that knows about the user record.
	"issue_reports": {
		Drop: []string{"contact_phone", "contactPhone"},
	},
}
