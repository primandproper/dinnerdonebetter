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

// The data change events a waitlist write emits. They are declared in the
// webhook event catalog (internal/domain/webhooks/catalog), so a subscriber is
// already able to ask for them.
const (
	// WaitlistCreatedServiceEventType indicates a waitlist was opened.
	WaitlistCreatedServiceEventType = "waitlist_created"
	// WaitlistUpdatedServiceEventType indicates a waitlist's name, description
	// or closing time changed.
	WaitlistUpdatedServiceEventType = "waitlist_updated"
	// WaitlistArchivedServiceEventType indicates a waitlist was retired.
	WaitlistArchivedServiceEventType = "waitlist_archived"

	// WaitlistSignupCreatedServiceEventType indicates somebody joined a waitlist.
	WaitlistSignupCreatedServiceEventType = "waitlist_signup_created"
	// WaitlistSignupUpdatedServiceEventType indicates the operator's note against
	// a signup was rewritten. It moves nobody through the queue.
	WaitlistSignupUpdatedServiceEventType = "waitlist_signup_updated"
	// WaitlistSignupTransitionedServiceEventType indicates a signup moved through
	// the lifecycle — invited, or converted.
	//
	// It is distinct from the update event because the two answer different
	// questions and only one of them is a queue. A subscriber that sends the
	// invitation email cannot key off an event that also fires when somebody
	// fixed a typo in a note.
	WaitlistSignupTransitionedServiceEventType = "waitlist_signup_transitioned"
	// WaitlistSignupWithdrawnServiceEventType indicates somebody asked to come
	// off a list.
	//
	// It is its own event rather than a transition, because a withdrawal is the
	// one move a subscriber must not treat as ordinary queue movement: it is a
	// standing instruction to stop writing to that address, and a consumer that
	// learned about it from a generic "transitioned" event would have to know to
	// inspect the status before acting.
	WaitlistSignupWithdrawnServiceEventType = "waitlist_signup_withdrawn"
	// WaitlistSignupArchivedServiceEventType indicates a signup was retired
	// administratively. It is not a withdrawal — see the platform package.
	WaitlistSignupArchivedServiceEventType = "waitlist_signup_archived"
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
