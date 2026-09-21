package authorization

import (
	platformauthz "github.com/primandproper/primitives-go/v2/authorization"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterPolicyResolver registers what a role grants with the injector.
//
// It reads the policy tables the migrator seeds, and every container that serves an
// authenticated request needs it: resolving a principal turns the role names a membership
// carries into the permissions they grant, and without this there is nothing to resolve
// them against.
//
// It lives here rather than beside a repository, which is where it used to live. The
// identity repository registered it because the identity repository was the only thing that
// used it; the directory is platform's now and has no opinion about what a role grants, so
// the registration belongs with the policy rather than with the rows it is resolved for.
func RegisterPolicyResolver(i do.Injector) {
	do.Provide[platformauthz.PolicyResolver](i, func(i do.Injector) (platformauthz.PolicyResolver, error) {
		return NewDatabaseResolver(
			do.MustInvoke[database.Client](i).Reader(),
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[metrics.Provider](i),
		)
	})
}
