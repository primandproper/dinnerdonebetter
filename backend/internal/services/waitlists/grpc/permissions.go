package grpc

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	waitlistssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/waitlists"
)

// WaitlistsMethodPermissions is a named type for Wire dependency injection.
type WaitlistsMethodPermissions map[string][]authorization.Permission

// ProvideMethodPermissions returns a Wire provider for the waitlists service's method permissions.
func ProvideMethodPermissions() WaitlistsMethodPermissions {
	return WaitlistsMethodPermissions{
		waitlistssvc.WaitlistsService_CreateWaitlist_FullMethodName: {
			authorization.CreateWaitlistsPermission,
		},
		waitlistssvc.WaitlistsService_GetWaitlist_FullMethodName: {
			authorization.ReadWaitlistsPermission,
		},
		waitlistssvc.WaitlistsService_GetWaitlists_FullMethodName: {
			authorization.ReadWaitlistsPermission,
		},
		waitlistssvc.WaitlistsService_GetActiveWaitlists_FullMethodName: {
			authorization.ReadWaitlistsPermission,
		},
		waitlistssvc.WaitlistsService_UpdateWaitlist_FullMethodName: {
			authorization.UpdateWaitlistsPermission,
		},
		waitlistssvc.WaitlistsService_ArchiveWaitlist_FullMethodName: {
			authorization.ArchiveWaitlistsPermission,
		},
		waitlistssvc.WaitlistsService_WaitlistIsNotExpired_FullMethodName: {
			authorization.ReadWaitlistsPermission,
		},
		waitlistssvc.WaitlistsService_CreateWaitlistSignup_FullMethodName: {
			authorization.CreateWaitlistSignupsPermission,
		},
		waitlistssvc.WaitlistsService_GetWaitlistSignup_FullMethodName: {
			authorization.ReadWaitlistSignupsPermission,
		},
		waitlistssvc.WaitlistsService_GetWaitlistSignupsForWaitlist_FullMethodName: {
			authorization.ReadWaitlistSignupsPermission,
		},
		waitlistssvc.WaitlistsService_UpdateWaitlistSignup_FullMethodName: {
			authorization.UpdateWaitlistSignupsPermission,
		},
		waitlistssvc.WaitlistsService_ArchiveWaitlistSignup_FullMethodName: {
			authorization.ArchiveWaitlistSignupsPermission,
		},
	}
}
