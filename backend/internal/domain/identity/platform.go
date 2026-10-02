package identity

import "github.com/primandproper/primitives-go/v2/tenancy"

// Scope is the tenancy every identity read and write in this application is made in.
//
// Global, because this deployment has one directory. A user is not an account's object and
// neither is an account: the account a request acts on is named by the request and checked
// by the target authorizer, which reads a live membership rather than trusting a scope
// somebody put on a struct.
func Scope() tenancy.Scope { return tenancy.Global() }
