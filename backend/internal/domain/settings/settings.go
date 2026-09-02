/*
Package settings is this application's half of platform-go's settings store: the
namespace its tables carry, the tenancy the catalog is kept under, whose answer a
stored value is, and the data change events a write emits.

The store itself is platform-go's. It owns the schema, the paging, the tenancy
column, the enumeration a value is checked against, the tri-state resolution, and
the guard that refuses an edit to a definition some stored value no longer
satisfies. What is not the platform's is who a value belongs to, which is decided
here.

# One subject type, and the foreign key that buys

Platform's schema files a value against a subject — a (type, id) pair — and
leaves the set of types open, so an application whose settings hang off a device
or a workspace can say so. This one uses exactly one: [SubjectUser]. Every
setting value in this deployment is a fact a person chose about themselves.

That is a narrowing of what the table it replaced allowed, and it is the point
rather than a casualty. A service_setting_configurations row carried
belongs_to_user *and* belongs_to_account, and the pair made two things true that
should not have been:

  - The account read returned other people's answers. Listing an account's
    configurations filtered on belongs_to_account alone, so any member holding
    read.service_setting_configurations saw every other member's personal
    preferences.
  - Nothing was ever account-owned. Every write set both columns from the
    session, so a row a user wrote in one account was invisible to them in
    another — a per-person preference silently filed per membership.

One subject type ends both, and it is what lets the schema keep the property the
old table had: subject_id names a user in every row, so it carries a foreign key
to users with ON DELETE CASCADE, and the single identity eraser goes on covering
settings. A domain whose subject column is mixed cannot have that key — see
internal/domain/waitlists/privacy for what a domain has to build instead.

An account-owned setting is therefore a schema change away rather than a line of
code away: dropping that foreign key, and deciding what erases the rows it was
holding. Nothing in this application wants one today.
*/
package settings

import (
	platformsettings "github.com/primandproper/platform-go/v13/settings"
	"github.com/primandproper/platform-go/v13/tenancy"
)

// TablePrefix namespaces the platform-go settings tables, rendering
// ddb_settings_definitions, ddb_settings_definition_options and
// ddb_settings_values.
//
// The platform's own default is the empty prefix. Nothing this replaced carried
// those names, so the prefix is not avoiding a collision the way waitlists' is —
// it says which application created the tables in a database that may hold more
// than one.
const TablePrefix = "ddb"

// The data change events a settings write emits. They are declared in the
// webhook event catalog (internal/domain/webhooks/catalog), so a subscriber is
// already able to ask for them.
//
// There is one value event for writing and one for clearing, and no
// created/updated pair. Platform's SetValue converges on the row, so a first
// answer, a changed answer and an answer revived after being cleared are the
// same statement — a subscriber that cared about the difference would be
// reading it off an event that cannot tell them. Clearing is its own event
// because it leads somewhere different: "they no longer have an opinion" is not
// a new preference to act on.
const (
	// SettingDefinitionCreatedServiceEventType indicates a setting was added to
	// the catalog.
	SettingDefinitionCreatedServiceEventType = "setting_definition_created"
	// SettingDefinitionUpdatedServiceEventType indicates a setting's kind,
	// default, enumeration or description changed.
	SettingDefinitionUpdatedServiceEventType = "setting_definition_updated"
	// SettingDefinitionArchivedServiceEventType indicates a setting was retired.
	SettingDefinitionArchivedServiceEventType = "setting_definition_archived"

	// SettingValueSetServiceEventType indicates somebody answered a setting.
	SettingValueSetServiceEventType = "setting_value_set"
	// SettingValueClearedServiceEventType indicates somebody took their answer back.
	SettingValueClearedServiceEventType = "setting_value_cleared"
)

// Scope is the tenancy this application keeps the settings catalog under, which
// is the global one.
//
// It is a decision rather than a default. A setting here is a service-wide
// definition — what "user_temperature_unit" means, and which units it admits —
// administered by service admins, and the table this replaced had no ownership
// column at all. Filing definitions per account would make a setting invisible to
// the operator who defined it the moment they switched accounts, and would put
// two scopes into every resolution besides: platform requires a definition and
// the values against it to share one.
func Scope() tenancy.Scope { return tenancy.Global() }

// SubjectFor is the principal a stored setting value belongs to.
//
// Every value this application writes names a user; see the package
// documentation for why there is no second subject type.
func SubjectFor(userID string) platformsettings.Subject {
	return platformsettings.Subject{Type: platformsettings.SubjectUser, ID: userID}
}
