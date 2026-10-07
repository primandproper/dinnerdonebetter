package identity

import (
	platformauthz "github.com/primandproper/primitives-go/v2/authorization"
	"github.com/primandproper/primitives-go/v2/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterSessionBuilder registers what turns a resolved principal into session context.
//
// It is separate from the identity surface because it is not part of one: every
// authenticated request on every surface resolves a principal through this, and the gRPC
// service is only one of the things that needs it.
func RegisterSessionBuilder(i do.Injector) {
	do.Provide[*SessionBuilder](i, func(i do.Injector) (*SessionBuilder, error) {
		return NewSessionBuilder(
			do.MustInvoke[platformauthz.PolicyResolver](i),
			do.MustInvoke[tracing.Provider](i),
		)
	})
}
