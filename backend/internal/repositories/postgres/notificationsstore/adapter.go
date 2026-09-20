package notificationsstore

import (
	"context"

	ddbnotifications "github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications"

	platformnotifications "github.com/primandproper/platform-go/v14/notifications"
	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/filtering"
	"github.com/primandproper/primitives-go/v2/identifiers"
)

// Adapter presents platform's inbox and device registry as this application's
// notifications.Repository.
//
// It exists so that the manager, the push fanout and the privacy collector do
// not change when the store underneath them does. Those three speak this
// application's vocabulary — UserNotification, UserDeviceToken — and that
// vocabulary is the API's, so translating here rather than at each of them is
// one conversion instead of three.
//
// It is the seam that will go last. Once nothing outside this package speaks
// UserNotification, the manager can take platform's types and this file can be
// deleted; until then it is what keeps the swap to one repository.
type Adapter struct {
	_ struct{} `json:"-"`

	inbox    platformnotifications.Inbox
	registry platformnotifications.Registry
	db       database.Client
}

var _ ddbnotifications.Repository = (*Adapter)(nil)

// NewAdapter builds the adapter over a decorated inbox and registry.
func NewAdapter(inbox platformnotifications.Inbox, registry platformnotifications.Registry, db database.Client) (*Adapter, error) {
	if inbox == nil || registry == nil || db == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	return &Adapter{inbox: inbox, registry: registry, db: db}, nil
}

// ---- notifications ----

func notificationToDomain(n *platformnotifications.Notification) *ddbnotifications.UserNotification {
	if n == nil {
		return nil
	}

	// Status is derived rather than stored. platform records when a notification
	// was read, which is strictly more than whether, and this application's
	// vocabulary only ever had the two words.
	status := ddbnotifications.UserNotificationStatusTypeUnread
	if n.ReadAt != nil {
		status = ddbnotifications.UserNotificationStatusTypeRead
	}

	return &ddbnotifications.UserNotification{
		CreatedAt:     n.CreatedAt,
		LastUpdatedAt: n.LastUpdatedAt,
		ID:            n.ID,
		Content:       n.Body,
		Status:        status,
		BelongsToUser: n.Principal,
	}
}

// GetUserNotification reads one notification.
func (a *Adapter) GetUserNotification(ctx context.Context, userID, userNotificationID string) (*ddbnotifications.UserNotification, error) {
	found, err := a.inbox.GetNotification(ctx, a.db.Reader(), ddbnotifications.Scope(), userID, userNotificationID)
	if err != nil {
		return nil, err
	}

	return notificationToDomain(found), nil
}

// UserNotificationExists reports whether the notification is there.
func (a *Adapter) UserNotificationExists(ctx context.Context, userID, userNotificationID string) (bool, error) {
	_, err := a.inbox.GetNotification(ctx, a.db.Reader(), ddbnotifications.Scope(), userID, userNotificationID)
	if err != nil {
		if platformerrors.Is(err, platformnotifications.ErrNotificationNotFound) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

// GetUserNotifications pages a person's inbox.
func (a *Adapter) GetUserNotifications(
	ctx context.Context,
	userID string,
	filter *filtering.QueryFilter,
) (*filtering.QueryFilteredResult[ddbnotifications.UserNotification], error) {
	page, err := a.inbox.ListNotifications(ctx, a.db.Reader(), ddbnotifications.Scope(), userID, filter)
	if err != nil {
		return nil, err
	}

	out := make([]*ddbnotifications.UserNotification, 0, len(page.Data))
	for _, n := range page.Data {
		out = append(out, notificationToDomain(n))
	}

	return filtering.NewQueryFilteredResult(
		out, page.FilteredCount, page.TotalCount,
		func(n *ddbnotifications.UserNotification) string { return n.ID },
		filter,
	), nil
}

// CreateUserNotification writes one, in a transaction of its own.
func (a *Adapter) CreateUserNotification(
	ctx context.Context,
	input *ddbnotifications.UserNotificationDatabaseCreationInput,
) (*ddbnotifications.UserNotification, error) {
	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	id := input.ID
	if id == "" {
		id = identifiers.New()
	}

	var created *platformnotifications.Notification

	if err := a.db.WithTransaction(ctx, func(tx database.Tx) error {
		var writeErr error
		created, writeErr = a.inbox.CreateNotification(ctx, tx, ddbnotifications.Scope(), &platformnotifications.Notification{
			ID:        id,
			Principal: input.BelongsToUser,
			Body:      input.Content,
			Topic:     ddbnotifications.DefaultTopic,
			// Title is platform's and this application has never had it: every
			// notification it writes is a line of text. It is left empty rather
			// than invented, and the day a notification wants a heading the input
			// grows a field rather than this guessing one.
		})

		return writeErr
	}); err != nil {
		return nil, err
	}

	return notificationToDomain(created), nil
}

// UpdateUserNotification marks the notification read.
//
// The local vocabulary's only update was the status, and its only transition was
// unread to read — platform's MarkNotificationRead is that, and there is nothing
// else an update could have meant.
func (a *Adapter) UpdateUserNotification(ctx context.Context, updated *ddbnotifications.UserNotification) error {
	if updated == nil {
		return platformerrors.ErrNilInputParameter
	}

	if updated.Status != ddbnotifications.UserNotificationStatusTypeRead {
		return nil
	}

	return a.db.WithTransaction(ctx, func(tx database.Tx) error {
		_, err := a.inbox.MarkNotificationRead(ctx, tx, ddbnotifications.Scope(), updated.BelongsToUser, updated.ID)

		return err
	})
}

// ---- device tokens ----

func deviceToDomain(d *platformnotifications.Device) *ddbnotifications.UserDeviceToken {
	if d == nil {
		return nil
	}

	return &ddbnotifications.UserDeviceToken{
		CreatedAt:     d.CreatedAt,
		ID:            d.ID,
		DeviceToken:   d.Token,
		Platform:      string(d.Platform),
		BelongsToUser: d.Principal,
	}
}

// GetUserDeviceTokens pages a person's handsets, optionally one platform's.
func (a *Adapter) GetUserDeviceTokens(
	ctx context.Context,
	userID string,
	filter *filtering.QueryFilter,
	platformFilter *string,
) (*filtering.QueryFilteredResult[ddbnotifications.UserDeviceToken], error) {
	page, err := a.registry.ListDevices(ctx, a.db.Reader(), ddbnotifications.Scope(), userID, filter)
	if err != nil {
		return nil, err
	}

	out := make([]*ddbnotifications.UserDeviceToken, 0, len(page.Data))
	for _, d := range page.Data {
		// The platform filter is applied here rather than in the query, because
		// platform's registry does not take one. A person's handsets are a handful
		// of rows, so the page is already small; a deployment where that stops
		// being true wants the filter pushed into the store rather than a bigger
		// page read here.
		if platformFilter != nil && *platformFilter != "" && string(d.Platform) != *platformFilter {
			continue
		}

		out = append(out, deviceToDomain(d))
	}

	return filtering.NewQueryFilteredResult(
		out, page.FilteredCount, page.TotalCount,
		func(d *ddbnotifications.UserDeviceToken) string { return d.ID },
		filter,
	), nil
}

// GetUserDeviceToken reads one handset.
//
// Through the list, because platform's registry has no keyed read: a device is
// found by its owner and its token rather than by an identifier a client holds.
func (a *Adapter) GetUserDeviceToken(ctx context.Context, userID, tokenID string) (*ddbnotifications.UserDeviceToken, error) {
	page, err := a.registry.ListDevices(ctx, a.db.Reader(), ddbnotifications.Scope(), userID, nil)
	if err != nil {
		return nil, err
	}

	for _, d := range page.Data {
		if d != nil && d.ID == tokenID {
			return deviceToDomain(d), nil
		}
	}

	return nil, platformnotifications.ErrDeviceNotFound
}

// UserDeviceTokenExists reports whether the handset is registered.
func (a *Adapter) UserDeviceTokenExists(ctx context.Context, userID, tokenID string) (bool, error) {
	found, err := a.GetUserDeviceToken(ctx, userID, tokenID)
	if err != nil {
		if platformerrors.Is(err, platformnotifications.ErrDeviceNotFound) {
			return false, nil
		}

		return false, err
	}

	return found != nil, nil
}

// CreateUserDeviceToken registers a handset, in a transaction of its own.
func (a *Adapter) CreateUserDeviceToken(
	ctx context.Context,
	input *ddbnotifications.UserDeviceTokenDatabaseCreationInput,
) (*ddbnotifications.UserDeviceToken, error) {
	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	id := input.ID
	if id == "" {
		id = identifiers.New()
	}

	var registered *platformnotifications.Device

	if err := a.db.WithTransaction(ctx, func(tx database.Tx) error {
		var writeErr error
		registered, writeErr = a.registry.RegisterDevice(ctx, tx, ddbnotifications.Scope(), &platformnotifications.Device{
			ID:        id,
			Principal: input.BelongsToUser,
			Token:     input.DeviceToken,
			Platform:  platformnotifications.Platform(input.Platform),
		})

		return writeErr
	}); err != nil {
		return nil, err
	}

	return deviceToDomain(registered), nil
}

// UpdateUserDeviceToken has no counterpart and no caller.
//
// platform's registry offers no device update, and the reason is the model: a
// handset is identified by its token, so changing the token is registering a
// different device and changing its platform is describing the same one wrongly.
// Nothing in this application has ever called it — the RPC that would have is
// absent from the deleted service too — so it refuses rather than pretending.
func (a *Adapter) UpdateUserDeviceToken(context.Context, *ddbnotifications.UserDeviceToken) error {
	return platformerrors.New("a registered device is not updated; revoke it and register the new token")
}

// ArchiveUserDeviceToken revokes a handset, in a transaction of its own.
//
// Revoking removes the row rather than flagging it, which is platform's model
// and the reason its devices carry no archived_at. The name is kept because the
// callers are this application's and it is what they already say.
func (a *Adapter) ArchiveUserDeviceToken(ctx context.Context, userID, tokenID string) error {
	return a.db.WithTransaction(ctx, func(tx database.Tx) error {
		_, err := a.registry.RevokeDevice(ctx, tx, ddbnotifications.Scope(), userID, tokenID)

		return err
	})
}
