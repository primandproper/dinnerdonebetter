package authorization

import (
	waitlistsgrpc "github.com/primandproper/platform-go/v15/waitlists/grpc"
)

// The waitlist permissions are platform's, re-exported under the names this
// application's policy already spells. See comments_permissions.go.
//
// EraseWaitlistSignupsPermission has no local predecessor: it gates the bulk
// withdrawal a data privacy erasure performs, which this application reaches
// through the eraser rather than the wire. It is granted to nobody.
const (
	// CreateWaitlistsPermission is a permission.
	CreateWaitlistsPermission = waitlistsgrpc.PermissionCreateLists
	// ReadWaitlistsPermission is a permission.
	ReadWaitlistsPermission = waitlistsgrpc.PermissionReadLists
	// UpdateWaitlistsPermission is a permission.
	UpdateWaitlistsPermission = waitlistsgrpc.PermissionUpdateLists
	// ArchiveWaitlistsPermission is a permission.
	ArchiveWaitlistsPermission = waitlistsgrpc.PermissionArchiveLists
	// ReadWaitlistSignupsPermission is a permission.
	ReadWaitlistSignupsPermission = waitlistsgrpc.PermissionReadSignups
	// UpdateWaitlistSignupsPermission is a permission.
	UpdateWaitlistSignupsPermission = waitlistsgrpc.PermissionUpdateSignups
	// InviteWaitlistSignupsPermission is a permission.
	InviteWaitlistSignupsPermission = waitlistsgrpc.PermissionInviteSignups
	// ConvertWaitlistSignupsPermission is a permission.
	ConvertWaitlistSignupsPermission = waitlistsgrpc.PermissionConvertSignups
	// ArchiveWaitlistSignupsPermission is a permission.
	ArchiveWaitlistSignupsPermission = waitlistsgrpc.PermissionArchiveSignups
	// EraseWaitlistSignupsPermission gates the bulk withdrawal an erasure performs.
	EraseWaitlistSignupsPermission = waitlistsgrpc.PermissionEraseSignups
)

// ReadOwnWaitlistSignupsPermission gates asking where one's own signups are.
//
// platform puts four reads behind one grant — a signup by id, one by the address
// it was made with, a list's page, and one subject's signups — and the sharpest
// of them is an oracle over every address in the deployment. That grant is a
// service admin's and stays one. But the last of the four is the only question a
// member has any reason to ask, and it is a question about themselves, so it is
// split off here and pointed at the same RPC under a name a member can hold.
//
// Holding it is not the whole answer. The subject comes off the request, so the
// grant alone would let a member name somebody else; whose signups these are is
// decided inside the handler by platform's AuthorizeSubjectRead, which this
// deployment answers with own-subject-or-admin. See internal/build/waitlists.
const ReadOwnWaitlistSignupsPermission Permission = "waitlists.signups.read_own"
