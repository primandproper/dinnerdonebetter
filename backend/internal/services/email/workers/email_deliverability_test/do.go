package emaildeliverabilitytest

import (
	"github.com/primandproper/primitives-go/v2/email"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterEmailDeliverabilityTest registers the email deliverability test job with the injector.
func RegisterEmailDeliverabilityTest(i do.Injector) {
	do.Provide[*Job](i, func(i do.Injector) (*Job, error) {
		return NewJob(
			do.MustInvoke[email.Emailer](i),
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[*JobParams](i),
		)
	})
}
