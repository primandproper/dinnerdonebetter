/*
Package identitystore wires platform-go's identity service into this application's
recording spine and its users search index.

The recording is platform's. identity.RecordingHooks writes the audit entries and
publishes the events every identity operation owes, on the operation's own
database.Tx, so they commit with the rows they describe or not at all. Nothing here
decides what an operation records; see platform-go's identity/recording.go.

Platform's events feed this application's users search index on their own: the envelope
the Emitter wraps each one in answers searchsync.Change, and the rules in
internal/indexevents match platform's event names and read the user's ID by the
payload's own field name. So this package is the constructor and nothing more.
*/
package identitystore

import (
	platformidentity "github.com/primandproper/platform-go/v15/identity"
	platformrecording "github.com/primandproper/platform-go/v15/recording"
)

// Hooks is platform's identity.RecordingHooks.
type Hooks = platformidentity.RecordingHooks

// ProvideHooks builds platform's recording hooks over the spine's Recorder. A nil Recorder is
// refused by platform.
func ProvideHooks(recorder *platformrecording.Recorder) (*Hooks, error) {
	return platformidentity.NewRecordingHooks(recorder)
}
