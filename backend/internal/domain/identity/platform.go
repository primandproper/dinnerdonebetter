package identity

import "github.com/primandproper/primitives-go/v2/tenancy"

// TablePrefix namespaces platform's identity tables, rendering ddb_identity_users,
// ddb_identity_accounts, ddb_identity_memberships, ddb_identity_invitations and the three
// role tables beside them.
//
// The same namespace every other adopted store here carries, and for the same reason: it
// says which application's rows these are in a database that may hold another's.
//
// It does not exist to avoid a collision. platform already names its tables identity_*,
// which collides with nothing this application created — and that is what makes the
// adoption buildable one piece at a time, because the two schemas can stand side by side
// until the last of the old readers is gone.
const TablePrefix = "ddb"

// Scope is the tenancy every identity read and write in this application is made in.
//
// Global, because this deployment has one directory. A user is not an account's object and
// neither is an account: the account a request acts on is named by the request and checked
// by the target authorizer, which reads a live membership rather than trusting a scope
// somebody put on a struct.
func Scope() tenancy.Scope { return tenancy.Global() }
