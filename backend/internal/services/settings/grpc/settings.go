package grpc

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"
	ddbsettings "github.com/primandproper/dinnerdonebetter/backend/internal/domain/settings"
	settingskeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/settings/keys"
	settingssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/settings"
	"github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/types"
	"github.com/primandproper/dinnerdonebetter/backend/internal/services/settings/grpc/converters"

	platformerrors "github.com/primandproper/platform-go/v13/errors"
	errorsgrpc "github.com/primandproper/platform-go/v13/errors/grpc"
	"github.com/primandproper/platform-go/v13/filtering"
	"github.com/primandproper/platform-go/v13/filtering/filteringpb"
	filteringgrpc "github.com/primandproper/platform-go/v13/filtering/grpc"
	"github.com/primandproper/platform-go/v13/observability/logging"
	"github.com/primandproper/platform-go/v13/observability/tracing"
	platformsettings "github.com/primandproper/platform-go/v13/settings"

	"google.golang.org/grpc/codes"
)

// The two refusals this service makes on its own, before the store is reached.
//
// They are here rather than in internal/services/settings/errors because they
// are not the store's: platform deliberately records AdminOnly without enforcing
// it — it has no notion of who is calling — and both of these are that
// enforcement.
var (
	// errNotAuthorizedForSettingDefinition is a non-admin reaching for a setting
	// marked admin-only, or for the administrative read of who has overridden
	// one.
	errNotAuthorizedForSettingDefinition = platformerrors.New("not authorized for this setting")
	// errInputRequired is a request whose input message is absent.
	errInputRequired = platformerrors.New("input is required")
)

// requester is what every method here needs first: the session, whose user is
// the subject of every value this service reads or writes.
//
// The scope is not among them. The whole catalog this application keeps is
// global — see ddbsettings.Scope — so the scope is a constant, and whose answer
// a value is comes from the session instead.
func (s *serviceImpl) requester(ctx context.Context, span tracing.Span) (*sessions.ContextData, error) {
	sessionContextData, err := sessions.RequireFromContext(ctx)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, s.logger.WithSpan(span), span, codes.Unauthenticated, "fetching session context data")
	}

	tracing.AttachToSpan(span, identitykeys.UserIDKey, sessionContextData.GetUserID())
	tracing.AttachToSpan(span, identitykeys.AccountIDKey, sessionContextData.GetActiveAccountID())

	return sessionContextData, nil
}

// pageRequest is the session plus the page a list method was asked for.
func (s *serviceImpl) pageRequest(ctx context.Context, span tracing.Span, protoFilter *filteringpb.QueryFilter) (*sessions.ContextData, *filtering.QueryFilter, error) {
	sessionContextData, err := s.requester(ctx, span)
	if err != nil {
		return nil, nil, err
	}

	filter, err := filteringgrpc.FromProto(protoFilter)
	if err != nil {
		return nil, nil, errorsgrpc.PrepareAndLogGRPCStatus(err, s.logger.WithSpan(span), span, codes.InvalidArgument, "invalid query filter")
	}

	tracing.AttachQueryFilterToSpan(span, filter)

	return sessionContextData, filter, nil
}

// responseDetails is the envelope every response carries.
func responseDetails(span tracing.Span, sessionContextData *sessions.ContextData) *types.ResponseDetails {
	return &types.ResponseDetails{
		TraceId:          span.SpanContext().TraceID().String(),
		CurrentAccountId: sessionContextData.GetActiveAccountID(),
	}
}

// CreateSettingDefinition adds a setting to the catalog.
func (s *serviceImpl) CreateSettingDefinition(ctx context.Context, request *settingssvc.CreateSettingDefinitionRequest) (*settingssvc.CreateSettingDefinitionResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	sessionContextData, err := s.requester(ctx, span)
	if err != nil {
		return nil, err
	}

	logger := s.logger.WithSpan(span)

	if request.GetInput() == nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(errInputRequired, logger, span, codes.InvalidArgument, "missing input")
	}

	definition := converters.ConvertGRPCSettingDefinitionCreationRequestInputToSettingDefinition(request.GetInput())

	created, err := s.settings.CreateDefinition(ctx, ddbsettings.Scope(), definition)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "creating setting definition")
	}

	return &settingssvc.CreateSettingDefinitionResponse{
		ResponseDetails: responseDetails(span, sessionContextData),
		Created:         converters.ConvertSettingDefinitionToGRPCSettingDefinition(created),
	}, nil
}

// GetSettingDefinition reads one setting by id.
func (s *serviceImpl) GetSettingDefinition(ctx context.Context, request *settingssvc.GetSettingDefinitionRequest) (*settingssvc.GetSettingDefinitionResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	sessionContextData, err := s.requester(ctx, span)
	if err != nil {
		return nil, err
	}

	logger := s.logger.WithSpan(span).WithValue(settingskeys.SettingDefinitionIDKey, request.GetSettingDefinitionId())
	tracing.AttachToSpan(span, settingskeys.SettingDefinitionIDKey, request.GetSettingDefinitionId())

	definition, err := s.settings.GetDefinition(ctx, ddbsettings.Scope(), request.GetSettingDefinitionId())
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching setting definition")
	}

	if err = s.readable(definition, sessionContextData, logger, span); err != nil {
		return nil, err
	}

	return &settingssvc.GetSettingDefinitionResponse{
		ResponseDetails: responseDetails(span, sessionContextData),
		Result:          converters.ConvertSettingDefinitionToGRPCSettingDefinition(definition),
	}, nil
}

// GetSettingDefinitionByName reads one setting by the name application code
// spells, which is the handle every value-side call takes.
func (s *serviceImpl) GetSettingDefinitionByName(ctx context.Context, request *settingssvc.GetSettingDefinitionByNameRequest) (*settingssvc.GetSettingDefinitionByNameResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	sessionContextData, err := s.requester(ctx, span)
	if err != nil {
		return nil, err
	}

	logger := s.logger.WithSpan(span).WithValue(settingskeys.SettingNameKey, request.GetSettingName())
	tracing.AttachToSpan(span, settingskeys.SettingNameKey, request.GetSettingName())

	definition, err := s.settings.GetDefinitionByName(ctx, ddbsettings.Scope(), request.GetSettingName())
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching setting definition by name")
	}

	if err = s.readable(definition, sessionContextData, logger, span); err != nil {
		return nil, err
	}

	return &settingssvc.GetSettingDefinitionByNameResponse{
		ResponseDetails: responseDetails(span, sessionContextData),
		Result:          converters.ConvertSettingDefinitionToGRPCSettingDefinition(definition),
	}, nil
}

// GetSettingDefinitions pages the catalog, less the settings the requester may
// not see.
func (s *serviceImpl) GetSettingDefinitions(ctx context.Context, request *settingssvc.GetSettingDefinitionsRequest) (*settingssvc.GetSettingDefinitionsResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	sessionContextData, filter, err := s.pageRequest(ctx, span, request.GetFilter())
	if err != nil {
		return nil, err
	}

	logger := s.logger.WithSpan(span)

	page, err := s.settings.ListDefinitions(ctx, ddbsettings.Scope(), filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching setting definitions")
	}

	admin := sessionContextData.GetServicePermissions().IsServiceAdmin()

	results := make([]*settingssvc.SettingDefinition, 0, len(page.Data))
	for _, definition := range page.Data {
		if definition.AdminOnly && !admin {
			continue
		}

		results = append(results, converters.ConvertSettingDefinitionToGRPCSettingDefinition(definition))
	}

	return &settingssvc.GetSettingDefinitionsResponse{
		ResponseDetails: responseDetails(span, sessionContextData),
		Pagination:      filteringgrpc.PaginationToProto(page.Pagination),
		Results:         results,
	}, nil
}

// UpdateSettingDefinition rewrites a setting in the catalog.
//
// It is a read-modify-write because platform's UpdateDefinition rewrites the
// whole row — enumeration included — and that is what lets it walk the stored
// values first and refuse an edit that would strand one.
func (s *serviceImpl) UpdateSettingDefinition(ctx context.Context, request *settingssvc.UpdateSettingDefinitionRequest) (*settingssvc.UpdateSettingDefinitionResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	sessionContextData, err := s.requester(ctx, span)
	if err != nil {
		return nil, err
	}

	logger := s.logger.WithSpan(span).WithValue(settingskeys.SettingDefinitionIDKey, request.GetSettingDefinitionId())
	tracing.AttachToSpan(span, settingskeys.SettingDefinitionIDKey, request.GetSettingDefinitionId())

	if request.GetInput() == nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(errInputRequired, logger, span, codes.InvalidArgument, "missing input")
	}

	existing, err := s.settings.GetDefinition(ctx, ddbsettings.Scope(), request.GetSettingDefinitionId())
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching setting definition")
	}

	converters.ApplyGRPCSettingDefinitionUpdateRequestInput(existing, request.GetInput())

	if err = s.settings.UpdateDefinition(ctx, ddbsettings.Scope(), existing); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "updating setting definition")
	}

	return &settingssvc.UpdateSettingDefinitionResponse{
		ResponseDetails: responseDetails(span, sessionContextData),
		Updated:         converters.ConvertSettingDefinitionToGRPCSettingDefinition(existing),
	}, nil
}

// ArchiveSettingDefinition retires a setting. The answers stored against it are
// left alone and its name stays claimed; see the platform package.
func (s *serviceImpl) ArchiveSettingDefinition(ctx context.Context, request *settingssvc.ArchiveSettingDefinitionRequest) (*settingssvc.ArchiveSettingDefinitionResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	sessionContextData, err := s.requester(ctx, span)
	if err != nil {
		return nil, err
	}

	logger := s.logger.WithSpan(span).WithValue(settingskeys.SettingDefinitionIDKey, request.GetSettingDefinitionId())
	tracing.AttachToSpan(span, settingskeys.SettingDefinitionIDKey, request.GetSettingDefinitionId())

	if err = s.settings.ArchiveDefinition(ctx, ddbsettings.Scope(), request.GetSettingDefinitionId()); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "archiving setting definition")
	}

	return &settingssvc.ArchiveSettingDefinitionResponse{
		ResponseDetails: responseDetails(span, sessionContextData),
	}, nil
}

// SetSettingValue stores the requester's own answer.
//
// Whose answer it is comes from the session and never from the request, which is
// what makes "may I write this" a question about the setting rather than about
// the subject: there is no way to spell somebody else's id here.
func (s *serviceImpl) SetSettingValue(ctx context.Context, request *settingssvc.SetSettingValueRequest) (*settingssvc.SetSettingValueResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	sessionContextData, definition, err := s.writableSetting(ctx, span, request.GetSettingName())
	if err != nil {
		return nil, err
	}

	logger := s.logger.WithSpan(span).WithValue(settingskeys.SettingNameKey, definition.Name)

	value, err := s.settings.SetValue(ctx, ddbsettings.Scope(), ddbsettings.SubjectFor(sessionContextData.GetUserID()), definition.Name, request.GetValue())
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "setting setting value")
	}

	return &settingssvc.SetSettingValueResponse{
		ResponseDetails: responseDetails(span, sessionContextData),
		Value:           converters.ConvertSettingValueToGRPCSettingValue(value),
	}, nil
}

// GetSettingValue reads the answer the requester stored, or NotFound when they
// have not answered. ResolveSetting is what applies the default.
func (s *serviceImpl) GetSettingValue(ctx context.Context, request *settingssvc.GetSettingValueRequest) (*settingssvc.GetSettingValueResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	sessionContextData, err := s.requester(ctx, span)
	if err != nil {
		return nil, err
	}

	logger := s.logger.WithSpan(span).WithValue(settingskeys.SettingNameKey, request.GetSettingName())
	tracing.AttachToSpan(span, settingskeys.SettingNameKey, request.GetSettingName())

	value, err := s.settings.GetValue(ctx, ddbsettings.Scope(), ddbsettings.SubjectFor(sessionContextData.GetUserID()), request.GetSettingName())
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching setting value")
	}

	return &settingssvc.GetSettingValueResponse{
		ResponseDetails: responseDetails(span, sessionContextData),
		Result:          converters.ConvertSettingValueToGRPCSettingValue(value),
	}, nil
}

// ClearSettingValue takes the requester's answer back, leaving them on the
// setting's default.
func (s *serviceImpl) ClearSettingValue(ctx context.Context, request *settingssvc.ClearSettingValueRequest) (*settingssvc.ClearSettingValueResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	sessionContextData, err := s.requester(ctx, span)
	if err != nil {
		return nil, err
	}

	logger := s.logger.WithSpan(span).WithValue(settingskeys.SettingNameKey, request.GetSettingName())
	tracing.AttachToSpan(span, settingskeys.SettingNameKey, request.GetSettingName())

	if err = s.settings.ClearValue(ctx, ddbsettings.Scope(), ddbsettings.SubjectFor(sessionContextData.GetUserID()), request.GetSettingName()); err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "clearing setting value")
	}

	return &settingssvc.ClearSettingValueResponse{
		ResponseDetails: responseDetails(span, sessionContextData),
	}, nil
}

// GetSettingValues pages everything the requester has answered.
func (s *serviceImpl) GetSettingValues(ctx context.Context, request *settingssvc.GetSettingValuesRequest) (*settingssvc.GetSettingValuesResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	sessionContextData, filter, err := s.pageRequest(ctx, span, request.GetFilter())
	if err != nil {
		return nil, err
	}

	logger := s.logger.WithSpan(span)

	page, err := s.settings.ListValuesForSubject(ctx, ddbsettings.Scope(), ddbsettings.SubjectFor(sessionContextData.GetUserID()), filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching setting values")
	}

	results := make([]*settingssvc.SettingValue, 0, len(page.Data))
	for _, value := range page.Data {
		results = append(results, converters.ConvertSettingValueToGRPCSettingValue(value))
	}

	return &settingssvc.GetSettingValuesResponse{
		ResponseDetails: responseDetails(span, sessionContextData),
		Pagination:      filteringgrpc.PaginationToProto(page.Pagination),
		Results:         results,
	}, nil
}

// GetSettingValuesForDefinition pages everyone who has answered one setting.
//
// Only a service admin may. It is the administrative read behind "who has
// overridden this", and every row it returns is somebody else's preference —
// which is exactly what the table this replaced handed to any account member
// who asked for the account's configurations.
func (s *serviceImpl) GetSettingValuesForDefinition(ctx context.Context, request *settingssvc.GetSettingValuesForDefinitionRequest) (*settingssvc.GetSettingValuesForDefinitionResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	sessionContextData, filter, err := s.pageRequest(ctx, span, request.GetFilter())
	if err != nil {
		return nil, err
	}

	logger := s.logger.WithSpan(span).WithValue(settingskeys.SettingNameKey, request.GetSettingName())
	tracing.AttachToSpan(span, settingskeys.SettingNameKey, request.GetSettingName())

	if !sessionContextData.GetServicePermissions().IsServiceAdmin() {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(errNotAuthorizedForSettingDefinition, logger, span, codes.PermissionDenied, "not authorized to list the answers to a setting")
	}

	page, err := s.settings.ListValuesForDefinition(ctx, ddbsettings.Scope(), request.GetSettingName(), filter)
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching setting values for definition")
	}

	results := make([]*settingssvc.SettingValue, 0, len(page.Data))
	for _, value := range page.Data {
		results = append(results, converters.ConvertSettingValueToGRPCSettingValue(value))
	}

	return &settingssvc.GetSettingValuesForDefinitionResponse{
		ResponseDetails: responseDetails(span, sessionContextData),
		Pagination:      filteringgrpc.PaginationToProto(page.Pagination),
		Results:         results,
	}, nil
}

// ResolveSetting answers one setting for the requester: their value, else the
// setting's default, else neither.
func (s *serviceImpl) ResolveSetting(ctx context.Context, request *settingssvc.ResolveSettingRequest) (*settingssvc.ResolveSettingResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	sessionContextData, err := s.requester(ctx, span)
	if err != nil {
		return nil, err
	}

	logger := s.logger.WithSpan(span).WithValue(settingskeys.SettingNameKey, request.GetSettingName())
	tracing.AttachToSpan(span, settingskeys.SettingNameKey, request.GetSettingName())

	resolution, err := s.settings.Resolve(ctx, ddbsettings.Scope(), ddbsettings.SubjectFor(sessionContextData.GetUserID()), request.GetSettingName())
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "resolving setting")
	}

	if err = s.readable(resolution.Definition, sessionContextData, logger, span); err != nil {
		return nil, err
	}

	tracing.AttachToSpan(span, settingskeys.SettingResolutionSourceKey, resolution.Source.String())

	return &settingssvc.ResolveSettingResponse{
		ResponseDetails: responseDetails(span, sessionContextData),
		Result:          converters.ConvertSettingResolutionToGRPCSettingResolution(resolution),
	}, nil
}

// ResolveSettings answers every setting the requester may see, in one round
// trip.
//
// It is the read a preferences page makes, and it is the reason this service has
// a resolution on the wire at all. The page a client renders is every setting
// paired with the answer that applies — the person's own, or the default they
// have not overridden — and a client assembling that from a catalog and a list
// of values has to reimplement the fallback, which is the part with a tri-state
// in it.
func (s *serviceImpl) ResolveSettings(ctx context.Context, _ *settingssvc.ResolveSettingsRequest) (*settingssvc.ResolveSettingsResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	sessionContextData, err := s.requester(ctx, span)
	if err != nil {
		return nil, err
	}

	logger := s.logger.WithSpan(span)

	resolutions, err := s.settings.ResolveAll(ctx, ddbsettings.Scope(), ddbsettings.SubjectFor(sessionContextData.GetUserID()))
	if err != nil {
		return nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "resolving settings")
	}

	admin := sessionContextData.GetServicePermissions().IsServiceAdmin()

	results := make([]*settingssvc.SettingResolution, 0, len(resolutions))
	for _, resolution := range resolutions {
		if resolution.Definition.AdminOnly && !admin {
			continue
		}

		results = append(results, converters.ConvertSettingResolutionToGRPCSettingResolution(resolution))
	}

	return &settingssvc.ResolveSettingsResponse{
		ResponseDetails: responseDetails(span, sessionContextData),
		Results:         results,
	}, nil
}

// readable refuses a non-admin a setting marked admin-only.
//
// Platform records AdminOnly and does not enforce it, deliberately: a store that
// decided who may read a row would be an authorization check in the wrong layer.
// This is the caller it hands that decision to, and the check is here rather
// than in the permission map because it is per-row — the same method reads a
// setting anybody may see and one only an administrator may.
//
// The refusal is PermissionDenied rather than NotFound because the setting's
// existence is not the secret: its name is in every catalog listing an
// administrator makes, and pretending it is absent would tell a confused caller
// to go and define it.
func (s *serviceImpl) readable(
	definition *platformsettings.Definition,
	sessionContextData *sessions.ContextData,
	logger logging.Logger,
	span tracing.Span,
) error {
	if definition == nil || !definition.AdminOnly {
		return nil
	}

	if sessionContextData.GetServicePermissions().IsServiceAdmin() {
		return nil
	}

	return errorsgrpc.PrepareAndLogGRPCStatus(errNotAuthorizedForSettingDefinition, logger, span, codes.PermissionDenied, "not authorized for this setting")
}

// writableSetting is the guard every value write shares: the session, and the
// setting it is about, read first so that an admin-only one can be refused.
//
// The read is not wasted work. Platform's SetValue reads the definition anyway —
// inside the write's own transaction, which is what checks the value against the
// enumeration — so what this adds is one read, and what it buys is that
// "admin-only" means something on the write path as well as the read one.
func (s *serviceImpl) writableSetting(
	ctx context.Context,
	span tracing.Span,
	name string,
) (*sessions.ContextData, *platformsettings.Definition, error) {
	sessionContextData, err := s.requester(ctx, span)
	if err != nil {
		return nil, nil, err
	}

	logger := s.logger.WithSpan(span).WithValue(settingskeys.SettingNameKey, name)
	tracing.AttachToSpan(span, settingskeys.SettingNameKey, name)

	definition, err := s.settings.GetDefinitionByName(ctx, ddbsettings.Scope(), name)
	if err != nil {
		return nil, nil, errorsgrpc.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching setting definition by name")
	}

	if err = s.readable(definition, sessionContextData, logger, span); err != nil {
		return nil, nil, err
	}

	return sessionContextData, definition, nil
}
