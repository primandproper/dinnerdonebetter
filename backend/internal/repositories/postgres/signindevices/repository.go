package signindevices

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/devices"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/signindevices/generated"

	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/observability"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
)

const (
	o11yName = "sign_in_devices_db_client"
)

var _ devices.Store = (*Repository)(nil)

// Repository keeps the device behind each sign-in. See internal/authentication/devices.
type Repository struct {
	client  database.Client
	tracer  tracing.Tracer
	querier generated.Querier
}

// ProvideSignInDevicesRepository builds a Repository.
func ProvideSignInDevicesRepository(tracerProvider tracing.Provider, client database.Client) *Repository {
	return &Repository{
		client:  client,
		tracer:  tracing.NewNamedTracer(tracerProvider, o11yName),
		querier: generated.New(),
	}
}

// RecordSignInDevice implements devices.Store.
func (r *Repository) RecordSignInDevice(ctx context.Context, q database.SQLQueryExecutor, device *devices.Device) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if device == nil {
		return platformerrors.ErrNilInputParameter
	}

	if device.FamilyID == "" || device.UserID == "" {
		return platformerrors.ErrInvalidIDProvided
	}

	if err := r.querier.UpsertSignInDevice(ctx, q, &generated.UpsertSignInDeviceParams{
		FamilyID:      device.FamilyID,
		BelongsToUser: device.UserID,
		IpAddress:     device.IPAddress,
		UserAgent:     device.UserAgent,
		DeviceName:    device.DeviceName,
		ExpiresAt:     device.ExpiresAt,
	}); err != nil {
		return observability.PrepareError(err, span, "recording a sign-in device")
	}

	return nil
}

// GetSignInDevicesForFamilies implements devices.Store.
func (r *Repository) GetSignInDevicesForFamilies(ctx context.Context, userID string, familyIDs []string) ([]*devices.Device, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}

	rows, err := r.querier.GetSignInDevicesForFamilies(ctx, r.client.Reader(), &generated.GetSignInDevicesForFamiliesParams{
		BelongsToUser: userID,
		FamilyIds:     familyIDs,
	})
	if err != nil {
		return nil, observability.PrepareError(err, span, "reading sign-in devices")
	}

	out := make([]*devices.Device, 0, len(rows))
	for _, row := range rows {
		out = append(out, &devices.Device{
			FamilyID:   row.FamilyID,
			UserID:     userID,
			IPAddress:  row.IpAddress,
			UserAgent:  row.UserAgent,
			DeviceName: row.DeviceName,
			CreatedAt:  row.CreatedAt,
			LastSeenAt: row.LastSeenAt,
		})
	}

	return out, nil
}

// GetSignInDevicesForUser implements devices.Store.
func (r *Repository) GetSignInDevicesForUser(ctx context.Context, userID string) ([]*devices.Device, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}

	rows, err := r.querier.GetSignInDevicesForUser(ctx, r.client.Reader(), userID)
	if err != nil {
		return nil, observability.PrepareError(err, span, "reading a user's sign-in devices")
	}

	out := make([]*devices.Device, 0, len(rows))
	for _, row := range rows {
		out = append(out, &devices.Device{
			FamilyID:   row.FamilyID,
			UserID:     userID,
			IPAddress:  row.IpAddress,
			UserAgent:  row.UserAgent,
			DeviceName: row.DeviceName,
			CreatedAt:  row.CreatedAt,
			LastSeenAt: row.LastSeenAt,
			ExpiresAt:  row.ExpiresAt,
		})
	}

	return out, nil
}

// Sweep deletes the devices of every login that can no longer be alive, and reports how many.
func (r *Repository) Sweep(ctx context.Context) (int64, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	deleted, err := r.querier.DeleteExpiredSignInDevices(ctx, r.client.Writer())
	if err != nil {
		return 0, observability.PrepareError(err, span, "sweeping expired sign-in devices")
	}

	return deleted, nil
}
