package devices

import (
	"context"
	"encoding/json"

	"github.com/primandproper/platform-go/v15/authentication/signin"
	signingrpc "github.com/primandproper/platform-go/v15/authentication/signin/grpc"
	platformdataprivacy "github.com/primandproper/platform-go/v15/dataprivacy"
	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/tenancy"
)

// recordingHooks are a sign-in service's hooks with the device recorded beside every token issued.
type recordingHooks struct {
	signin.Hooks

	store Store
}

// NewHooks wraps a sign-in service's hooks so that every token issued for a login records where
// it was issued to.
//
// The record is written on the token's own transaction, after the hooks it wraps have run, so a
// login whose token was not issued records no device, and a device that could not be recorded
// fails the sign-in rather than leaving a login the screen cannot place.
//
// An impersonation records nothing. The request behind one is the operator's, and the login it
// begins is listed to the subject — who would be shown their operator's address.
func NewHooks(inner signin.Hooks, store Store) signin.Hooks {
	return &recordingHooks{Hooks: inner, store: store}
}

// AfterIssueToken implements signin.Hooks.
func (h *recordingHooks) AfterIssueToken(ctx context.Context, tx database.Tx, scope tenancy.Scope, signIn *signin.SignIn) error {
	if err := h.Hooks.AfterIssueToken(ctx, tx, scope, signIn); err != nil {
		return err
	}

	if signIn == nil || signIn.ActorID != "" || signIn.FamilyID == "" {
		return nil
	}

	device := ForSignIn(ctx, signIn)
	if device.UserID == "" {
		return nil
	}

	if err := h.store.RecordSignInDevice(ctx, tx, device); err != nil {
		return platformerrors.Wrap(err, "recording a sign-in's device")
	}

	return nil
}

// NewAnnotator answers what was recorded about the devices behind a person's logins, in the shape
// signingrpc.WithSignInAnnotator takes. A login with nothing recorded gets no attributes.
func NewAnnotator(store Store) signingrpc.SignInAnnotator {
	return func(ctx context.Context, _ tenancy.Scope, userID string, familyIDs []string) (map[string]map[string]string, error) {
		if len(familyIDs) == 0 {
			return map[string]map[string]string{}, nil
		}

		devices, err := store.GetSignInDevicesForFamilies(ctx, userID, familyIDs)
		if err != nil {
			return nil, platformerrors.Wrap(err, "reading the devices behind a person's sign-ins")
		}

		annotations := make(map[string]map[string]string, len(devices))
		for _, device := range devices {
			annotations[device.FamilyID] = device.Attributes()
		}

		return annotations, nil
	}
}

// NewCollector is the sign-in devices' section of a subject access request: every address, user
// agent and device name recorded against the person's logins.
//
// There is no eraser beside it. The rows carry a foreign key to the user with ON DELETE CASCADE,
// so the identity eraser takes them with the person.
func NewCollector(store Store) platformdataprivacy.Collector {
	return platformdataprivacy.CollectorFunc(func(ctx context.Context, _ tenancy.Scope, subject platformdataprivacy.Subject) (json.RawMessage, error) {
		devices, err := store.GetSignInDevicesForUser(ctx, subject.ID)
		if err != nil {
			return nil, platformerrors.Wrap(err, "reading a person's sign-in devices")
		}

		if len(devices) == 0 {
			return nil, nil
		}

		return json.Marshal(devices)
	})
}
