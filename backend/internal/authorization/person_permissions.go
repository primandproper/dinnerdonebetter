package authorization

import (
	dataprivacyhttp "github.com/primandproper/platform-go/v15/dataprivacy/http"
	mediaregistrygrpc "github.com/primandproper/platform-go/v15/mediaregistry/grpc"
	operationshttp "github.com/primandproper/platform-go/v15/operations/http"
)

// The permissions of platform's surfaces that are about a person rather than an account: a
// privacy request is about its subject, the operation fulfilling it is filed under them, and an
// uploaded object is its uploader's, filed under the global scope. Every route and method is
// narrowed to the caller its resolver answers, so holding one of these reaches the holder's own
// rows and nobody else's.
const (
	// SubmitPrivacyRequestsPermission allows asking for an export or an erasure of oneself.
	SubmitPrivacyRequestsPermission = dataprivacyhttp.PermissionSubmitRequests
	// ReadPrivacyRequestsPermission allows reading one's own privacy requests.
	ReadPrivacyRequestsPermission = dataprivacyhttp.PermissionReadRequests
	// CancelPrivacyRequestsPermission allows withdrawing one's own privacy request.
	CancelPrivacyRequestsPermission = dataprivacyhttp.PermissionCancelRequests
	// ListOperationsPermission allows paging one's own operations.
	ListOperationsPermission = operationshttp.PermissionListOperations
	// CancelOperationsPermission allows cancelling one's own operation.
	CancelOperationsPermission = operationshttp.PermissionCancelOperations
	// CreateMediaObjectsPermission allows uploading an object as one's own, or registering bytes
	// already in one's own part of the bucket.
	CreateMediaObjectsPermission = mediaregistrygrpc.PermissionCreateObjects
	// ReadMediaObjectsPermission allows reading an object one uploaded: its row over gRPC, and its
	// bytes over HTTP. Platform declares one permission for the two, so the byte-serve and the
	// row reads are one grant.
	ReadMediaObjectsPermission = mediaregistrygrpc.PermissionReadObjects
	// ArchiveMediaObjectsPermission allows hiding an object one uploaded.
	ArchiveMediaObjectsPermission = mediaregistrygrpc.PermissionArchiveObjects
)

var (
	// ServiceUserPermissions is what every user holds as a person, whichever accounts they are
	// in or are not. It is service-wide on purpose: a person asking for their data to be erased
	// must not need an account membership to ask.
	ServiceUserPermissions = []Permission{
		SubmitPrivacyRequestsPermission,
		ReadPrivacyRequestsPermission,
		CancelPrivacyRequestsPermission,
		ListOperationsPermission,
		CancelOperationsPermission,
		CreateMediaObjectsPermission,
		ReadMediaObjectsPermission,
		ArchiveMediaObjectsPermission,
	}
)
