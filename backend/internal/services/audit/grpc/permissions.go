package grpc

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	auditsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/audit"
)

// AuditMethodPermissions is a named type for Wire dependency injection.
type AuditMethodPermissions map[string][]authorization.Permission

// ProvideMethodPermissions returns a Wire provider for the audit service's method permissions.
func ProvideMethodPermissions() AuditMethodPermissions {
	return AuditMethodPermissions{
		auditsvc.AuditService_GetAuditLogEntriesForAccount_FullMethodName: {
			authorization.ReadAuditLogEntriesPermission,
		},
		auditsvc.AuditService_GetAuditLogEntriesForUser_FullMethodName: {
			authorization.ReadAuditLogEntriesPermission,
		},
		auditsvc.AuditService_GetAuditLogEntryByID_FullMethodName: {
			authorization.ReadAuditLogEntriesPermission,
		},
	}
}
