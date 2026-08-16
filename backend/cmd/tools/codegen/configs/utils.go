package main

import (
	"fmt"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/branding"

	circuitbreakingcfg "github.com/primandproper/platform-go/v10/circuitbreaking/config"
	retrycfg "github.com/primandproper/platform-go/v10/retry/config"
	"github.com/primandproper/platform-go/v10/webhooks"
	webhookscfg "github.com/primandproper/platform-go/v10/webhooks/config"
)

// passkeyCeremonyTimeout bounds a WebAuthn ceremony everywhere it is bounded: the timeout the
// browser is asked to honor, the deadline go-webauthn enforces when the response comes back, and
// the TTL the ceremony's row is stored under. It used to be three numbers that could disagree.
//
// Two minutes rather than the platform's one. The specification's suggestion assumes a local
// authenticator — touch the key, done — and the slowest case here is cross-device: scan the QR
// code, unlock the phone, approve there. One minute is enough to fail that on a bad network, and
// now that the deadline is enforced server-side rather than merely requested of the browser, a
// ceremony that runs over is a login that fails rather than one that merely should have.
const passkeyCeremonyTimeout = 2 * time.Minute

func internalKubernetesEndpoint(serviceName, namespace string, port int) string {
	return fmt.Sprintf("%s.%s.svc.cluster.local:%d", serviceName, namespace, port)
}

// buildWebhooksConfig renders the outbound webhook configuration every environment shares.
//
// The delivery knobs are the platform's own defaults, materialized rather than left zero: these
// configs are validated at generation time, and validation requires values rather than accepting
// blanks and defaulting later. Writing them out also means the knobs an operator would reach for
// are visible in the config file instead of implied.
//
// The circuit breaker is named, and named per endpoint at runtime — this is the prefix. Without
// one, a tripped breaker reports that something is failing but not which subscriber.
func buildWebhooksConfig() webhookscfg.Config {
	cfg := webhookscfg.Config{
		CircuitBreaker: circuitbreakingcfg.Config{
			Name: "webhook_delivery",
			// A subscriber is allowed to fail half its deliveries before it stops
			// competing with healthy ones for the worker pool, and only once there
			// have been enough attempts for that rate to mean anything.
			ErrorRate:              .5,
			MinimumSampleThreshold: 100,
		},
		Worker: webhooks.WorkerConfig{
			// Identifies these deliveries in a subscriber's access log. The platform's
			// default names the library, which tells a subscriber which code sent the
			// request but not who did.
			UserAgent: branding.CompanyName + " Webhooks/1",
			Backoff: retrycfg.Config{
				// A subscriber that is down for a deploy should not exhaust its budget
				// during it. Ten attempts over a schedule that reaches an hour covers a
				// long restart without keeping a permanently broken endpoint alive.
				MaxAttempts:  10,
				InitialDelay: 5 * time.Second,
				MaxDelay:     time.Hour,
				Multiplier:   2,
				// Full jitter, because several workers share this table: without it their
				// retries re-collide on every round after one contended claim.
				UseJitter: true,
			},
		},
	}

	cfg.EnsureDefaults()

	return cfg
}
