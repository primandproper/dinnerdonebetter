package authorization

const (
	// CreateUserDataReportsPermission is a permission for a user to aggregate their own data report.
	CreateUserDataReportsPermission Permission = "create.user_data_reports"
	// ReadUserDataReportsPermission is a permission for a user to fetch their own data report.
	ReadUserDataReportsPermission Permission = "read.user_data_reports"
	// DestroyUserDataPermission is a permission for a user to destroy all of their own data.
	DestroyUserDataPermission Permission = "destroy.user_data"
)

var (
	// DataPrivacyPermissions contains all data-privacy-related permissions.
	DataPrivacyPermissions = []Permission{
		CreateUserDataReportsPermission,
		ReadUserDataReportsPermission,
		DestroyUserDataPermission,
	}
)
