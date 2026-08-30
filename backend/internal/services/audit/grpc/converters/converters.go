package converters

import (
	"encoding/json"
	"fmt"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	grpcconverters "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/converters"
	auditsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/audit"

	"github.com/primandproper/platform-go/v13/pointer"
)

func ConvertAuditLogEntryToGRPCAuditLogEntry(entry *audit.AuditLogEntry) *auditsvc.AuditLogEntry {
	changes := make(map[string]*auditsvc.ChangeLog, len(entry.Changes))
	for k := range entry.Changes {
		change := entry.Changes[k]
		changes[k] = &auditsvc.ChangeLog{
			OldValue: renderChangeValue(change.Old),
			NewValue: renderChangeValue(change.New),
		}
	}

	return &auditsvc.AuditLogEntry{
		CreatedAt:        grpcconverters.ConvertTimeToPBTimestamp(entry.CreatedAt),
		Changes:          changes,
		BelongsToAccount: pointer.Dereference(entry.BelongsToAccount),
		Id:               entry.ID,
		ResourceType:     entry.ResourceType,
		RelevantId:       entry.RelevantID,
		EventType:        entry.EventType,
		BelongsToUser:    entry.BelongsToUser,
		ActorType:        entry.ActorType,
		ActorIp:          entry.ActorIP,
		Scope:            entry.Scope,
		PrevHash:         entry.PrevHash,
		Hash:             entry.Hash,
		Seq:              entry.Seq,
	}
}

// renderChangeValue renders one half of a change for the wire.
//
// Values are stored typed — a numeric field stays numeric through the log — but
// this wire format has always carried strings, and a client showing an audit trail
// shows text. An absent half stays empty rather than becoming "null": a creation
// has no old value and a deletion has no new one, and rendering those as a literal
// is the kind of small lie that gets read back as data.
//
// A value that will not marshal falls back to Go's own rendering rather than
// failing the read. The entry is the evidence; refusing to display a page of
// history because one field held something exotic would be the wrong trade.
func renderChangeValue(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return v
	}

	encoded, err := json.Marshal(value)
	if err != nil {
		return fmt.Sprintf("%v", value)
	}

	return string(encoded)
}
