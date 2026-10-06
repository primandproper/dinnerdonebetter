/*
Package waitlists is this application's half of platform-go's waitlist store:
the namespace its tables carry, the tenancy every list is kept under, the
subject a signup belongs to, and the data change events a write emits.

The store itself is platform-go's. It owns the schema, the paging, the tenancy
column, the signup lifecycle and — the half this application did not have — the
withdrawal that keeps somebody off a list after they have asked to come off it.
What is not the platform's is who a signup belongs to and what address the list
writes to, and both of those are decided here.

# Tenancy

Every waitlist and every signup is kept under tenancy.Global(), and that is a
decision rather than a default. A waitlist here is an operator's record of which
of this deployment's users want a feature that does not exist yet — "which of
our users wants to opt into X" — so the catalog is one catalog, administered by
service admins, and the table this replaced carried no ownership column at all.
Filing lists per account would make a list invisible to the operator who opened
it the moment they switched accounts.

It does not follow that a signup is unowned. Who a signup belongs to is the
signup's Subject, not the list's scope; see SubjectFor.
*/
package waitlists

import (
	platformwaitlists "github.com/primandproper/platform-go/v15/waitlists"
)

// SubjectFor is the principal a signup made by a signed-in user belongs to.
//
// A signup made by a signed-in caller names one, and it is the subject that makes
// "which lists am I on" and a subject access request answerable. A visitor's
// signup names none: the signup page is public, and its address is vouched for by
// the double opt-in rather than by a session (see internal/build/waitlists). Such
// a signup is not reached by a subject access request or the subject eraser; it
// is reached by its address, through the unsubscribe link every confirmation
// mail carries, which withdraws it and suppresses the address.
//
// The account is deliberately not part of it. A signup follows the person rather
// than whichever account they had active when they filled the form in, and the
// permission that guards a signup is owner-or-service-admin, on the user.
func SubjectFor(userID string) platformwaitlists.Subject {
	return platformwaitlists.Subject{Type: platformwaitlists.SubjectUser, ID: userID}
}
