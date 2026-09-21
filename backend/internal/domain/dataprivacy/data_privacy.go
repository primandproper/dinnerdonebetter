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

	oauth2clientsprivacy "github.com/primandproper/platform-go/v14/authentication/oauth2clients/privacy"
	passkeysprivacy "github.com/primandproper/platform-go/v14/authentication/passkeys/privacy"
	passwordresetprivacy "github.com/primandproper/platform-go/v14/authentication/passwordreset/privacy"
	billingprivacy "github.com/primandproper/platform-go/v14/billing/privacy"
	commentsprivacy "github.com/primandproper/platform-go/v14/comments/privacy"
	identityprivacy "github.com/primandproper/platform-go/v14/identity/privacy"
	issuereportsprivacy "github.com/primandproper/platform-go/v14/issuereports/privacy"
	mediaregistryprivacy "github.com/primandproper/platform-go/v14/mediaregistry/privacy"
	settingsprivacy "github.com/primandproper/platform-go/v14/settings/privacy"
	waitlistsprivacy "github.com/primandproper/platform-go/v14/waitlists/privacy"
	"github.com/primandproper/primitives-go/v2/tenancy"
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
// attribute values in telemetry.
//
// Most of them are platform's now, re-exported rather than declared, for the reason the
// identity permissions give: the package that owns the collector owns the name it is
// normally registered under, and a constant here with a different string would be a
// section this deployment calls one thing and every other reader of platform's artifacts
// calls another. privacyadapters.Register uses each package's DefaultKey and nothing else,
// so a local spelling would not even be reachable.
//
// Two changed value when they moved. uploaded_media is media_registry and payments is
// billing, which renames those sections in every future artifact — free here, since
// nothing is deployed, and worth taking so the export's vocabulary is the module's.
//
// What is still declared below is what this application answers for itself.
const (
	// CollectorKeyIdentity covers the user record, their accounts, their
	// memberships, and the invitations they sent or received.
	//
	// It is this application's to declare even though the collector is platform's,
	// because the eraser filed under the same key is not: see internal/build/dataprivacy
	// for the succession rule that runs before the user row goes.
	CollectorKeyIdentity = identityprivacy.DefaultKey
	// CollectorKeyMealPlanning covers recipes, meals, meal plans, ingredient
	// preferences, and ratings. There is no platform counterpart; this is the domain
	// this application is.
	CollectorKeyMealPlanning = "meal_planning"
	// CollectorKeyNotifications covers in-app user notifications. platform ships a
	// collector for its own inbox and device registry and this application does not use
	// it — ours reads one repository and answers as one section where platform answers
	// as two.
	CollectorKeyNotifications = "notifications"
	// CollectorKeyAuditLog covers the audit entries recorded about the subject.
	CollectorKeyAuditLog = "audit_log"

	// The sections platform's own adapters register, named by the packages that own
	// them so that this application cannot drift from the artifact everybody else reads.
	CollectorKeySettings      = settingsprivacy.DefaultKey
	CollectorKeyIssueReports  = issuereportsprivacy.DefaultKey
	CollectorKeyMediaRegistry = mediaregistryprivacy.DefaultKey
	CollectorKeyWaitlists     = waitlistsprivacy.DefaultKey
	CollectorKeyComments      = commentsprivacy.DefaultKey
	CollectorKeyBilling       = billingprivacy.DefaultKey
	CollectorKeyPasskeys      = passkeysprivacy.DefaultKey
	CollectorKeyPasswordReset = passwordresetprivacy.DefaultKey
	CollectorKeyOAuth2Clients = oauth2clientsprivacy.DefaultKey

	// EraserKeyIdentity is the eraser that deletes the user row, and with it every
	// table in this schema that carries a foreign key to it.
	//
	// It is platform's, under platform's key, with this application's succession rule
	// running ahead of it as the adapter's BeforeErase — see
	// internal/domain/identity/privacy. It is the same string as CollectorKeyIdentity
	// because the two halves of one domain share a key; both are spelled so that a reader
	// looking for either finds it.
	EraserKeyIdentity = identityprivacy.DefaultKey
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

// UnconfinedScope is the tenancy scope this application files privacy requests under: none.
//
// A privacy request is about a person, not about a tenant. The subject is the user, the
// artifact is theirs, and a request submitted while one account was active still covers
// everything held about them — so there is no account to file it under.
//
// "No tenant" is the zero Scope rather than tenancy.Global, and the two are different
// answers here in a way they are not everywhere else. platform refuses a global privacy
// request outright, with ErrGlobalRequestScope: the global scope is the chain
// platform-level events are recorded in, and a request about a person is not one of those.
// The zero Scope is what it reads as unconfined, and it maps that back to Global itself
// when it comes to record the request's own audit entry — see dataprivacy.auditScope.
//
// The reads spell the same thing differently, and that is platform's shape rather than a
// slip here: they take a *tenancy.Scope where nil narrows nothing, and a non-nil pointer at
// the zero Scope is refused as a caller whose own lookup came back empty. So a write passes
// this and a read passes nil.
func UnconfinedScope() tenancy.Scope { return tenancy.Scope{} }
