package dataprivacy

import (
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/comments"
	domaindataprivacy "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/dataprivacy"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/issuereports"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/notifications"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/payments"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/settings"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/uploadedmedia"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/waitlists"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/webhooks"

	"github.com/primandproper/platform-go/v8/database"
	"github.com/primandproper/platform-go/v8/observability/logging"
	"github.com/primandproper/platform-go/v8/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterDataPrivacyRepository registers the data privacy repository with the injector.
func RegisterDataPrivacyRepository(i do.Injector) {
	do.Provide[domaindataprivacy.Repository](i, func(i do.Injector) (domaindataprivacy.Repository, error) {
		return ProvideDataPrivacyRepository(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.TracerProvider](i),
			do.MustInvoke[audit.Repository](i),
			do.MustInvoke[identity.Repository](i),
			do.MustInvoke[issuereports.Repository](i),
			do.MustInvoke[notifications.Repository](i),
			do.MustInvoke[settings.Repository](i),
			do.MustInvoke[uploadedmedia.Repository](i),
			do.MustInvoke[waitlists.Repository](i),
			do.MustInvoke[webhooks.Repository](i),
			do.MustInvoke[comments.Repository](i),
			do.MustInvoke[payments.Repository](i),
			do.MustInvoke[database.Client](i),
			do.MustInvoke[[]domaindataprivacy.UserDataCollector](i),
		), nil
	})
}
