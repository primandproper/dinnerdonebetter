package authorization

const (
	// ReadUserDataPermission is a service admin permission.
	ReadUserDataPermission Permission = "admin.read_user_data"
	// PublishArbitraryQueueMessagePermission is a service admin permission.
	PublishArbitraryQueueMessagePermission Permission = "queues.publish.message"
)
