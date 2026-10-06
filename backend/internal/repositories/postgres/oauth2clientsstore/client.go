package oauth2clientsstore

import (
	platformoauth2clients "github.com/primandproper/platform-go/v15/authentication/oauth2clients"
	platformrecording "github.com/primandproper/platform-go/v15/recording"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
)

// ProvideHooks builds what the registry Service records each write with: platform's own
// RecordingHooks over the recording.Recorder this application registers. A client's entries
// name its owner as their subject, so they are filed on the owner's chain.
func ProvideHooks(recorder *platformrecording.Recorder) (platformoauth2clients.Hooks, error) {
	hooks, err := platformoauth2clients.NewRecordingHooks(recorder)
	if err != nil {
		return nil, platformerrors.Wrap(err, "building the OAuth2 clients recording hooks")
	}

	return hooks, nil
}
