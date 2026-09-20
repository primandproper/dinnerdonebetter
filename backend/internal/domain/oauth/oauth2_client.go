/*
Package oauth is what remains of this application's OAuth2 client registry: the
namespace its tables live in, and the names of the events its writes publish.

The registry itself is platform's — see
internal/repositories/postgres/oauth2clientsstore, and
internal/build/oauth2clients for the surface over it. The types, the validation,
the manager and the repository that used to be here all had counterparts there,
and the two columns platform adds have no equivalent in anything this package
ever declared.
*/
package oauth

// TablePrefix namespaces the platform's oauth2 tables, rendering
// ddb_oauth2_registered_clients for the registry and ddb_oauth2_authorization_codes,
// ddb_oauth2_access_tokens and ddb_oauth2_refresh_tokens for the authorization server.
//
// A prefix rather than the platform's empty default, and it is now a tidiness
// decision rather than a collision one. It was introduced because the platform's
// authorization server names a table oauth2_clients, which was exactly the name
// this application's own registry had — and since that DDL says CREATE TABLE IF
// NOT EXISTS, the platform's schema would have been a silent no-op followed by a
// store reading columns that were not there.
//
// That registry is gone: the administered one is platform's too, and it renders
// oauth2_registered_clients, which collides with nothing. The prefix stays so
// that one application's oauth2 tables sort together in a database that may hold
// another's, which is what the namespace is documented to be for.
const TablePrefix = "ddb"

const (
	// ClientIDSize and ClientSecretSize are how many bytes a credential carries.
	//
	// They survive the adoption because localdev seeds a well-known credential of each
	// length through platform's CredentialGenerator seam. Nothing else mints one — the
	// registry service reads crypto/rand at these same lengths, and does not take them
	// from a caller.
	ClientIDSize     = 16
	ClientSecretSize = 16

	// OAuth2ClientCreatedServiceEventType indicates an OAuth2 client was created.
	OAuth2ClientCreatedServiceEventType = "oauth2_client_created"
	// OAuth2ClientUpdatedServiceEventType indicates a registration's description changed.
	//
	// The four descriptive fields only. A revision cannot rotate a secret — platform's
	// UpdateInput has no field for one, deliberately, because a credential changed by an
	// UPDATE is one nobody was handed a new value for — so this event never means the
	// thing a subscriber would most want to be told about. No RPC reaches it either; it
	// is published by the store, which is the seam a future one would go through.
	OAuth2ClientUpdatedServiceEventType = "oauth2_client_updated"
	// OAuth2ClientArchivedServiceEventType indicates an OAuth2 client was archived.
	OAuth2ClientArchivedServiceEventType = "oauth2_client_archived"
)
