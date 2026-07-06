//
//  PlatformServices.swift
//  ios
//
//  Composition root for the platform-swift cross-cutting services. Built once at app
//  startup and handed to components via their initializers — the Swift analog of how the
//  Go backend wires platform-go subsystems at boot. There is no service locator; the single
//  `shared` instance exists only because `AppDelegate` is constructed by UIKit outside our
//  control, and low-level networking helpers need a logger before any view exists.
//

import CircuitBreaking
import Foundation
import Observability
import Retry
import Secrets

/// Immutable, `Sendable` bundle of the platform observability + resilience primitives.
struct PlatformServices: Sendable {
  /// The three observability backends (logger + tracer + metrics), built once.
  let pillars: Pillars

  /// Retry policy applied to idempotent gRPC calls (see `authenticatedCall`).
  let retry: any RetryPolicy

  /// App-wide circuit breaker guarding the gRPC transport.
  let breaker: any CircuitBreaker

  /// Read-only secret source over env + Info.plist (build-time `Secrets.xcconfig` values).
  /// Note: token *storage* stays in `KeychainManager` — `SecretSource` is read-only.
  let secrets: any SecretSource

  /// Mint a component-scoped observer (a named logger + tracer). One per component.
  func observer(_ name: String) -> any Observer { makeObserver(name, pillars) }

  /// Convenience: a named logger for use outside of an `operation` scope.
  func logger(_ name: String) -> any Logger { pillars.logger.withName(name) }

  /// Build the platform services from configuration. `.default` observability is OSLog +
  /// signposts + swift-metrics — no infrastructure required.
  static func make() -> PlatformServices {
    let pillars = ObservabilityConfig(serviceName: Branding.companyNameSlug).bootstrap()

    var retryConfig = RetryConfig(
      maxAttempts: 3,
      initialDelay: .milliseconds(100),
      maxDelay: .seconds(5),
      multiplier: 2,
      useJitter: true
    )
    retryConfig.ensureDefaults()
    let retry = ExponentialBackoffPolicy(config: retryConfig)

    let breaker = StandardCircuitBreaker(
      name: "grpc",
      errorRatePercentage: 50,
      minimumSampleThreshold: 5,
      logger: pillars.logger,
      metrics: pillars.metrics
    )

    let secrets = EnvironmentSecretSource(pillars: pillars)

    return PlatformServices(pillars: pillars, retry: retry, breaker: breaker, secrets: secrets)
  }

  /// The single composition-root instance, built lazily on first access.
  static let shared = make()
}
