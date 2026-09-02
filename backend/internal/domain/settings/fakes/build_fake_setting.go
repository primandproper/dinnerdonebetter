// Package fakes builds the randomized setting definitions and values this
// application's tests write.
package fakes

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/settings"

	"github.com/primandproper/platform-go/v13/fake"
	"github.com/primandproper/platform-go/v13/filtering"
	"github.com/primandproper/platform-go/v13/pointer"
	platformsettings "github.com/primandproper/platform-go/v13/settings"
)

// BuildFakeSettingDefinition builds a faked Definition: a text setting that
// enumerates its values and defaults to one of them.
//
// Three fields are fixed rather than randomized, and each of them is a value the
// store refuses:
//
//   - The kind, because a random string is not one of the four the store parses,
//     and a definition carrying one could never be written.
//   - The enumeration and the default together, because a default has to be a
//     value the setting admits — a randomized pair would be refused nine times
//     in ten, and the tenth would be a coincidence.
//   - The scope, because a definition and the values against it share one and
//     this application's is global.
func BuildFakeSettingDefinition() *platformsettings.Definition {
	definition := fake.BuildFakeRecord[platformsettings.Definition]()

	definition.Kind = platformsettings.KindString
	definition.Scope = settings.Scope()

	chosen := fake.BuildFakeString()
	definition.Enumeration = []string{chosen, fake.BuildFakeString()}
	definition.Default = pointer.To(chosen)

	return definition
}

// BuildFakeSettingDefinitionList builds a faked page of Definitions.
func BuildFakeSettingDefinitionList() *filtering.QueryFilteredResult[platformsettings.Definition] {
	return fake.BuildFakePage(BuildFakeSettingDefinition)
}

// BuildFakeSettingValue builds a faked Value belonging to nobody in particular.
//
// The subject is left empty rather than randomized: a random one names a user
// that does not exist, and the schema's foreign key refuses it. A caller that
// wants a stored value names the user it belongs to — see
// BuildFakeSettingValueForUser.
func BuildFakeSettingValue() *platformsettings.Value {
	value := fake.BuildFakeRecord[platformsettings.Value]()

	value.Subject = platformsettings.Subject{}
	value.Scope = settings.Scope()

	return value
}

// BuildFakeSettingValueForUser builds a faked Value belonging to one user, which
// is the only shape this application writes.
func BuildFakeSettingValueForUser(userID string) *platformsettings.Value {
	value := BuildFakeSettingValue()
	value.Subject = settings.SubjectFor(userID)

	return value
}

// BuildFakeSettingValueList builds a faked page of Values.
func BuildFakeSettingValueList() *filtering.QueryFilteredResult[platformsettings.Value] {
	return fake.BuildFakePage(BuildFakeSettingValue)
}
