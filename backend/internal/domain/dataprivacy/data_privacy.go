/*
Package dataprivacy holds what this application contributes to platform-go's
GDPR/CCPA machinery: the registration keys, the table prefix, and the one lookup
a collector needs that its own domain cannot answer.

The state machine, the request table, the artifact packaging, the expiry sweep,
and the fan-out over domains all live in platform-go's dataprivacy package. What
used to live here was a UserDataCollection struct aggregating eleven domain
types, and it is worth recording why it is gone rather than merely refactored.

Every domain wrote into that one shared value, so adding a domain meant editing a
central type that transitively imported every domain package — the cost paid on
every schema change, by the file most likely to conflict. It also meant one
domain returning an error aborted the whole aggregate: a subject's entire export
failed because one unrelated table was slow.

platform-go inverts it. A Collector returns an opaque, already-encoded JSON
fragment and the library composes fragments by key, so a domain announces itself
in one line and a failure is recorded against its own key while the rest of the
export is still delivered. Each domain's collector now lives beside that domain,
in internal/domain/<domain>/privacy, and the only file that knows about all of
them is the registry wiring in internal/build/dataprivacy.
*/
package dataprivacy

import (
	"context"
)

// TablePrefix namespaces the platform's request table, rendering
// ddb_dataprivacy_requests.
//
// A prefix rather than the platform's empty default, for the same reason
// audit.TablePrefix carries one: the DDL says CREATE TABLE IF NOT EXISTS, so a
// name that collides with something this repository already created is a silent
// no-op followed by code running against the wrong columns. It is referenced by
// the migration that creates the table and by the Store that reads and writes
// it, and a prefix that differs between the two is the misconfiguration that
// stays invisible until somebody asks for their data and gets an empty answer.
const TablePrefix = "ddb"

// Registration keys. These become section names in the export artifact and
// attribute values in telemetry, so they are declared once rather than spelled
// at each registration site — a typo in one would silently rename a section of
// every artifact from then on, and the only symptom is a section missing from a
// file nobody reads until a regulator does.
//
// Adding a domain means adding a constant here and a line in
// internal/build/dataprivacy. It does not mean editing a type.
const (
	// CollectorKeyIdentity covers the user record, their accounts, their
	// memberships, and the invitations they sent or received.
	CollectorKeyIdentity = "identity"
	// CollectorKeyMealPlanning covers recipes, meals, meal plans, ingredient
	// preferences, and ratings.
	CollectorKeyMealPlanning = "meal_planning"
	// CollectorKeyWebhooks covers the webhooks belonging to the subject's
	// accounts.
	CollectorKeyWebhooks = "webhooks"
	// CollectorKeySettings covers user- and account-scoped setting
	// configurations.
	CollectorKeySettings = "settings"
	// CollectorKeyNotifications covers in-app user notifications.
	CollectorKeyNotifications = "notifications"
	// CollectorKeyPayments covers subscriptions, purchases, and payment
	// transactions.
	CollectorKeyPayments = "payments"
	// CollectorKeyAuditLog covers the audit entries recorded about the subject.
	CollectorKeyAuditLog = "audit_log"
	// CollectorKeyIssueReports covers issue reports filed from the subject's
	// accounts.
	CollectorKeyIssueReports = "issue_reports"
	// CollectorKeyUploadedMedia covers media the subject uploaded.
	CollectorKeyUploadedMedia = "uploaded_media"
	// CollectorKeyWaitlists covers the subject's waitlist signups.
	CollectorKeyWaitlists = "waitlists"
	// CollectorKeyComments covers comments the subject authored.
	CollectorKeyComments = "comments"

	// EraserKeyComments is the eraser that destroys the comments a subject wrote.
	//
	// It is registered because comments are the one store the identity cascade no
	// longer reaches. The table this repository used to own carried
	// belongs_to_user REFERENCES users ON DELETE CASCADE; platform-go's carries a
	// plain author column, because that package does not own the directory people
	// live in and has no table to point a foreign key at. Without this eraser a
	// user erasure would leave every comment they wrote in place, live and
	// listable.
	//
	// The erasure is a hard delete rather than an anonymization. A comment's body
	// is free text somebody typed, so what has to go is the words rather than a
	// flag beside them: keeping the text and losing the author would be worse than
	// either.
	//
	// It sorts before EraserKeyIdentity, which costs nothing — there is no foreign
	// key between the two any more — and is the order that reads correctly anyway.
	EraserKeyComments = "comments"

	// EraserKeyIdentity is the eraser that deletes the user row, and with it
	// everything hanging off it by ON DELETE CASCADE. See
	// internal/domain/identity/privacy for what that covers and why it is the
	// only application eraser registered for this schema's cascading tables.
	//
	// The other registered eraser is platform-go's own, under auditerasure.
	// DefaultKey. That key sorts before this one, which is load-bearing: erasers
	// run serially in sorted order inside one transaction, so the audit scopes
	// are resolved and deleted while the accounts that name them still exist.
	EraserKeyIdentity = "identity"
)

// AccountIDResolver answers "which accounts does this user appear in", which is
// the one question an account-scoped collector has to ask and cannot answer from
// its own domain.
//
// It is a function type rather than an identity.Repository parameter so that a
// domain collecting account-scoped data — webhooks, settings, payments — does
// not acquire a dependency on the identity domain to get it. The build layer
// supplies the implementation, and each collector takes exactly what it needs.
type AccountIDResolver func(ctx context.Context, userID string) ([]string, error)
