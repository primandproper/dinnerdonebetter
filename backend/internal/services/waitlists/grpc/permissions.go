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
		waitlistssvc.WaitlistsService_GetOpenWaitlists_FullMethodName: {
			authorization.ReadWaitlistsPermission,
		},
		waitlistssvc.WaitlistsService_UpdateWaitlist_FullMethodName: {
			authorization.UpdateWaitlistsPermission,
		},
		waitlistssvc.WaitlistsService_ArchiveWaitlist_FullMethodName: {
			authorization.ArchiveWaitlistsPermission,
		},
		waitlistssvc.WaitlistsService_WaitlistIsOpen_FullMethodName: {
			authorization.ReadWaitlistsPermission,
		},
		waitlistssvc.WaitlistsService_JoinWaitlist_FullMethodName: {
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
		// The three lifecycle moves are writes to a signup and are gated by the same
		// capability as a note. None of them is a permission of its own, because this
		// application has no role that may invite somebody but not amend their signup —
		// a permission nothing distinguishes is a line in the role grid that grants what
		// the one beside it already granted.
		//
		// What separates them is not the capability but who may aim it. Inviting and
		// converting are service-admin acts and are refused to anybody else by the
		// handler; withdrawing is the signup owner's own act, and is refused to anybody
		// but them and a service admin. See waitlists.go.
		waitlistssvc.WaitlistsService_InviteWaitlistSignup_FullMethodName: {
			authorization.UpdateWaitlistSignupsPermission,
		},
		waitlistssvc.WaitlistsService_ConvertWaitlistSignup_FullMethodName: {
			authorization.UpdateWaitlistSignupsPermission,
		},
		waitlistssvc.WaitlistsService_WithdrawFromWaitlist_FullMethodName: {
			authorization.UpdateWaitlistSignupsPermission,
		},
		waitlistssvc.WaitlistsService_ArchiveWaitlistSignup_FullMethodName: {
			authorization.ArchiveWaitlistSignupsPermission,
		},
	}
}
