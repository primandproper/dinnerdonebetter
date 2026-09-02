// Package converters carries setting definitions, values and resolutions between
// the wire and the platform store's shapes.
package converters

import (
	grpcconverters "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/converters"
	settingssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/settings"

	platformsettings "github.com/primandproper/platform-go/v13/settings"
)

// ConvertSettingDefinitionToGRPCSettingDefinition converts a stored definition
// to proto.
//
// The scope is deliberately not on the wire. Every definition this application
// keeps is in the global scope, so a field carrying it would say the same thing
// on every row.
func ConvertSettingDefinitionToGRPCSettingDefinition(input *platformsettings.Definition) *settingssvc.SettingDefinition {
	if input == nil {
		return nil
	}

	return &settingssvc.SettingDefinition{
		Id:            input.ID,
		Name:          input.Name,
		Description:   input.Description,
		Kind:          input.Kind.String(),
		DefaultValue:  input.Default,
		Enumeration:   input.Enumeration,
		AdminOnly:     input.AdminOnly,
		CreatedAt:     grpcconverters.ConvertTimeToPBTimestamp(input.CreatedAt),
		LastUpdatedAt: grpcconverters.ConvertTimePointerToPBTimestamp(input.LastUpdatedAt),
		ArchivedAt:    grpcconverters.ConvertTimePointerToPBTimestamp(input.ArchivedAt),
	}
}

// ConvertGRPCSettingDefinitionToSettingDefinition converts a proto definition
// back to the platform's, for a client asserting against what it was handed.
func ConvertGRPCSettingDefinitionToSettingDefinition(input *settingssvc.SettingDefinition) *platformsettings.Definition {
	if input == nil {
		return nil
	}

	return &platformsettings.Definition{
		ID:            input.GetId(),
		Name:          input.GetName(),
		Description:   input.GetDescription(),
		Kind:          platformsettings.Kind(input.GetKind()),
		Default:       input.DefaultValue,
		Enumeration:   input.GetEnumeration(),
		AdminOnly:     input.GetAdminOnly(),
		CreatedAt:     grpcconverters.ConvertPBTimestampToTime(input.GetCreatedAt()),
		LastUpdatedAt: grpcconverters.ConvertPBTimestampToTimePointer(input.GetLastUpdatedAt()),
		ArchivedAt:    grpcconverters.ConvertPBTimestampToTimePointer(input.GetArchivedAt()),
	}
}

// ConvertGRPCSettingDefinitionCreationRequestInputToSettingDefinition builds the
// definition the store writes from what the client sent.
//
// The id and the creation time are left unset: the store mints one and reads the
// other back from the database's clock. The scope is the caller's to supply,
// because it is the application's decision rather than the request's.
func ConvertGRPCSettingDefinitionCreationRequestInputToSettingDefinition(input *settingssvc.SettingDefinitionCreationRequestInput) *platformsettings.Definition {
	if input == nil {
		return nil
	}

	return &platformsettings.Definition{
		Name:        input.GetName(),
		Description: input.GetDescription(),
		Kind:        platformsettings.Kind(input.GetKind()),
		Default:     input.DefaultValue,
		Enumeration: input.GetEnumeration(),
		AdminOnly:   input.GetAdminOnly(),
	}
}

// ApplyGRPCSettingDefinitionUpdateRequestInput folds an update onto the
// definition as it stands.
//
// Platform's UpdateDefinition rewrites the whole row, so an edit is a
// read-modify-write and this is the modify. Every scalar is optional and an
// absent one is left alone; the enumeration is a wrapper message for the same
// reason, since a bare repeated field could not tell "leave it" from "empty it".
func ApplyGRPCSettingDefinitionUpdateRequestInput(definition *platformsettings.Definition, input *settingssvc.SettingDefinitionUpdateRequestInput) {
	if definition == nil || input == nil {
		return
	}

	if input.Name != nil {
		definition.Name = input.GetName()
	}

	if input.Description != nil {
		definition.Description = input.GetDescription()
	}

	if input.Kind != nil {
		definition.Kind = platformsettings.Kind(input.GetKind())
	}

	// A present default replaces the stored one. There is deliberately no way to
	// spell "this setting no longer has a default" here: proto3 gives an optional
	// scalar presence but no way to distinguish a null from an absence, and
	// clearing a default is an edit that changes what every subject who has not
	// chosen resolves to. It belongs in a method that says so rather than in the
	// field that also means "leave it alone".
	if input.DefaultValue != nil {
		definition.Default = input.DefaultValue
	}

	if input.Enumeration != nil {
		definition.Enumeration = input.GetEnumeration().GetValues()
	}

	if input.AdminOnly != nil {
		definition.AdminOnly = input.GetAdminOnly()
	}
}

// ConvertSettingValueToGRPCSettingValue converts a stored value to proto.
//
// The subject becomes a bare user id, because every value this application
// stores belongs to a user — see internal/domain/settings — so a type beside it
// would say "user" on every row.
func ConvertSettingValueToGRPCSettingValue(input *platformsettings.Value) *settingssvc.SettingValue {
	if input == nil {
		return nil
	}

	return &settingssvc.SettingValue{
		Id:            input.ID,
		DefinitionId:  input.DefinitionID,
		BelongsToUser: input.Subject.ID,
		Value:         input.Raw,
		CreatedAt:     grpcconverters.ConvertTimeToPBTimestamp(input.CreatedAt),
		LastUpdatedAt: grpcconverters.ConvertTimePointerToPBTimestamp(input.LastUpdatedAt),
		ArchivedAt:    grpcconverters.ConvertTimePointerToPBTimestamp(input.ArchivedAt),
	}
}

// ConvertGRPCSettingValueToSettingValue converts a proto value back to the
// platform's, for a client asserting against what it was handed.
func ConvertGRPCSettingValueToSettingValue(input *settingssvc.SettingValue) *platformsettings.Value {
	if input == nil {
		return nil
	}

	return &platformsettings.Value{
		ID:            input.GetId(),
		DefinitionID:  input.GetDefinitionId(),
		Subject:       platformsettings.Subject{Type: platformsettings.SubjectUser, ID: input.GetBelongsToUser()},
		Raw:           input.GetValue(),
		CreatedAt:     grpcconverters.ConvertPBTimestampToTime(input.GetCreatedAt()),
		LastUpdatedAt: grpcconverters.ConvertPBTimestampToTimePointer(input.GetLastUpdatedAt()),
		ArchivedAt:    grpcconverters.ConvertPBTimestampToTimePointer(input.GetArchivedAt()),
	}
}

// ConvertSettingResolutionToGRPCSettingResolution converts a resolution to
// proto, source and all.
func ConvertSettingResolutionToGRPCSettingResolution(input *platformsettings.Resolution) *settingssvc.SettingResolution {
	if input == nil {
		return nil
	}

	return &settingssvc.SettingResolution{
		Definition: ConvertSettingDefinitionToGRPCSettingDefinition(input.Definition),
		Value:      ConvertSettingValueToGRPCSettingValue(input.Value),
		Raw:        input.Raw,
		Source:     input.Source.String(),
	}
}
