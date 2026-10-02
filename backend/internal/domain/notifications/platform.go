package notifications

import (
	"github.com/primandproper/primitives-go/v2/tenancy"
)

// Scope is the tenancy every notification and device in this deployment is filed
// under.
//
// Global, and that is a decision rather than a default. A notification is
// addressed to a person, not to an account: somebody told that their meal plan
// finalized is told once, whichever household it was for, and scoping the inbox
// per account would split one person's notifications across as many inboxes as
// they have memberships. The account a notification is about travels in its body
// where it matters.
func Scope() tenancy.Scope { return tenancy.Global() }
