package queuetest

import (
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/internalops"

	"github.com/primandproper/platform-go/v5/messagequeue"
	"github.com/primandproper/platform-go/v5/observability/logging"
	"github.com/primandproper/platform-go/v5/observability/metrics"
	"github.com/primandproper/platform-go/v5/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterQueueTest registers the queue test job with the injector.
func RegisterQueueTest(i do.Injector) {
	do.Provide[*Job](i, func(i do.Injector) (*Job, error) {
		return NewJob(
			do.MustInvoke[internalops.InternalOpsDataManager](i),
			do.MustInvoke[messagequeue.PublisherProvider](i),
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.TracerProvider](i),
			do.MustInvoke[metrics.Provider](i),
			do.MustInvoke[*JobParams](i),
		)
	})
}
