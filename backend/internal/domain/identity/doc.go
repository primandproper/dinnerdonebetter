/*
Package identity is this application's half of platform-go's identity: the
events its operations emit, the vocabulary that travels with them, and the walk
over an account's roster. The household succession rule platform leaves to the
consumer is in the succession subpackage.

The directory itself — users, accounts, memberships, invitations — is
platform's; see internal/repositories/postgres/identitystore.

# Tenancy

Every identity read and write is made under tenancy.Global(), because this
deployment has one directory. A user is not an account's object and neither is
an account: the account a request acts on is named by the request and checked by
the target authorizer, which reads a live membership rather than trusting a
scope somebody put on a struct.
*/
package identity
