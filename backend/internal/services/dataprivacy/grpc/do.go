package grpc

import (
	dataprivacysvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/dataprivacy"

	platformdataprivacy "github.com/primandproper/platform-go/v12/dataprivacy"
	"github.com/primandproper/platform-go/v12/observability/logging"
	"github.com/primandproper/platform-go/v12/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterDataPrivacyService registers the data privacy gRPC service with the injector.
//
// It no longer needs a publisher or a queue name. Submitting a request writes a row
// the fulfillment worker claims, so the durability a message on a topic was
// providing now comes from the table the request lives in — and a request can no
// longer be accepted, acknowledged, and then lost because the broker dropped it.
func RegisterDataPrivacyService(i do.Injector) {
	do.Provide(i, func(i do.Injector) (dataprivacysvc.DataPrivacyServiceServer, error) {
		return NewDataPrivacyService(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[platformdataprivacy.Service](i),
		), nil
	})
}
