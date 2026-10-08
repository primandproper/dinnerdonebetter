/*
Package notifications is this application's half of platform-go's notifications:
the push fanout's registration, in push.

The inbox and the device registry are platform's; see
internal/repositories/postgres/notificationsstore.

# Tenancy

Every notification and device is filed under tenancy.Global(), and that is a
decision rather than a default. A notification is addressed to a person, not to
an account: somebody told that their meal plan finalized is told once, whichever
household it was for, and scoping the inbox per account would split one
person's notifications across as many inboxes as they have memberships. The
account a notification is about travels in its body where it matters.
*/
package notifications
