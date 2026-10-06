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

const (
	// ClientIDSize and ClientSecretSize are how many bytes a credential carries.
	//
	// They survive the adoption because localdev seeds a well-known credential of each
	// length through platform's CredentialGenerator seam. Nothing else mints one — the
	// registry service reads crypto/rand at these same lengths, and does not take them
	// from a caller.
	ClientIDSize     = 16
	ClientSecretSize = 16
)
