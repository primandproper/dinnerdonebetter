package keys

const (
	idSuffix = ".id"

	// UserDataAggregationReportIDKey is the standard key for referring to a user data aggregation report.
	UserDataAggregationReportIDKey = "user_data_aggregation_report" + idSuffix

	// UserDataDisclosureIDKey is the standard key for referring to a user data disclosure.
	UserDataDisclosureIDKey = "user_data_disclosure" + idSuffix

	// UserDataDisclosureArtifactPathKey is the standard key for referring to the object storage
	// path of a user data disclosure's report artifact.
	UserDataDisclosureArtifactPathKey = "user_data_disclosure.artifact_path"
)
