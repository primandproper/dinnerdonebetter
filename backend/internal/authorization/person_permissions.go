package authorization

import (
	dataprivacyhttp "github.com/primandproper/platform-go/v15/dataprivacy/http"
	mediaregistryhttp "github.com/primandproper/platform-go/v15/mediaregistry/http"
	operationshttp "github.com/primandproper/platform-go/v15/operations/http"
)

// The permissions of platform's HTTP surfaces, which are about a person rather than an account:
// a privacy request is about its subject, the operation fulfilling it is filed under them, and
// an uploaded object is its uploader's. Every route is narrowed to the caller its resolver
// answers, so holding one of these reaches the holder's own rows and nobody else's.
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
	// ReadMediaObjectsPermission allows fetching an object one uploaded.
	ReadMediaObjectsPermission = mediaregistryhttp.PermissionReadObjects
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
		ReadMediaObjectsPermission,
	}
)
