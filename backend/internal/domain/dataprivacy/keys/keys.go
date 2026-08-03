package keys

const (
	idSuffix = ".id"

	// RequestIDKey is the standard key for referring to a data privacy request — an
	// export or an erasure — by ID.
	//
	// One key where there were two. The old model had a disclosure ID naming a row and
	// a report ID naming an object, and a log line carrying one told you nothing about
	// the other. A request is now a single row that owns its artifact, so there is one
	// identifier to correlate on.
	RequestIDKey = "data_privacy_request" + idSuffix

	// RequestTypeKey distinguishes an export from an erasure.
	RequestTypeKey = "data_privacy_request.type"
)
