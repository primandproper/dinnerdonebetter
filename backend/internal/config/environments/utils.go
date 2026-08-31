package environments

import (
	"fmt"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/branding"

	circuitbreakingcfg "github.com/primandproper/platform-go/v13/circuitbreaking/config"
	retrycfg "github.com/primandproper/platform-go/v13/retry/config"
	"github.com/primandproper/platform-go/v13/webhooks"
	webhookscfg "github.com/primandproper/platform-go/v13/webhooks/config"
)

// Names the environment builders below repeat: the service's own name, the analytics
// and feature-flag sources, and the two placeholder markers a rendered prod config
// carries until deploy substitutes them.
const (
	serviceName          = "dinner-done-better"
	webPlatform          = "web"
	iosPlatform          = "ios"
	webAnalyticsSource   = "web_analytics"
	iosAnalyticsSource   = "ios_analytics"
	featureFlaggerSource = "feature_flagger"
	otelCollectorAddress = "otel_collector:4317"
	placeholderValue     = "placeholder"
	replaceAtDeploy      = "REPLACE_AT_DEPLOY"
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

// The user session store's expiry policy, shared by every environment.
//
// Two timeouts, because they answer different questions. Idle asks how long somebody may
// close the app and come back to it; absolute asks how long a session may exist at all,
// which is the only bound on a refresh token somebody stole — a thief is not idle.
//
// A week idle and thirty days absolute is the shape of a consumer application that people
// use a few times a week and expect to stay signed in to on their phone. Neither number is
// the platform's default, and the defaults are the reason to say so here: half an hour idle
// and a day absolute are a bank's numbers, and shipping them would sign this application's
// users out every morning.
//
// The touch interval is what keeps the idle timeout from costing a write per request. At one
// hour against a week, an active session is written to about twenty-four times a day instead
// of once per API call, and the price is an idle deadline that can be up to an hour stale —
// in the direction that expires a session early rather than late, which is the only direction
// a security control may be wrong in.
const (
	sessionIdleTimeout     = 7 * 24 * time.Hour
	sessionAbsoluteTimeout = 30 * 24 * time.Hour
	sessionTouchInterval   = time.Hour
)

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
