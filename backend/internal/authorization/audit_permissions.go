package authorization

import (
	auditgrpc "github.com/primandproper/platform-go/v14/audit/grpc"
)

// The audit log's permissions are platform's, re-exported under this
// application's names so that the roles below read the way the rest of them do.
//
// The names changed with the adoption: "read.audit_log_entries" became
// "audit.entries.read", and the chain verification that had no permission at all
// acquired one. Nothing was deployed under the old names, so nothing needed
// migrating — see the repository's deployment note.
const (
	// ReadAuditLogEntriesPermission allows reading entries from the audit log.
	ReadAuditLogEntriesPermission = auditgrpc.PermissionReadEntries

	// VerifyAuditChainPermission allows walking a scope's hash chain looking for a
	// break. It is separate from reading because the two answer different
	// questions: a reader asks what happened, and a verifier asks whether what is
	// recorded has been tampered with.
	VerifyAuditChainPermission = auditgrpc.PermissionVerifyChain
)

var (
	// AuditPermissions contains all audit-related permissions.
	AuditPermissions = []Permission{
		ReadAuditLogEntriesPermission,
		VerifyAuditChainPermission,
	}
)
