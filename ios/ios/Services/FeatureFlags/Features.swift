//
//  Features.swift
//  ios
//
//  Single source of truth for the app's feature toggles.
//
//  Local, build-/environment-derived toggles are exposed as synchronous static properties (the
//  existing idiom, e.g. `APIConfiguration.useSearchService`), aggregated here so every flag is
//  discoverable in one place. Remote, per-user flags go through `manager` — a platform-swift
//  `FeatureFlagManager` seam that is a no-op today and can be pointed at PostHog later via
//  `FeatureFlagsConfig(provider: "posthog", ...).makeFeatureFlagManager()`.
//
//  Named `Features` (not `FeatureFlags`) so it does not collide with the platform-swift
//  `FeatureFlags` module — a same-named type makes the compiler ignore `import FeatureFlags`.
//

import FeatureFlags
import Foundation

enum Features {
  // MARK: - Local toggles (synchronous)

  /// Use the dedicated search service (vs. DB search). Production-only today.
  static var useSearchService: Bool { APIConfiguration.useSearchService }

  /// Route analytics through the backend passthrough instead of Segment.
  static var useAnalyticsBackend: Bool { AnalyticsConfiguration.useAnalyticsBackend }

  /// Running against the mock auth backend (UI tests).
  static var useMockAuth: Bool { ProcessInfo.processInfo.arguments.contains("--use-mock-auth") }

  // MARK: - Remote flags (platform-swift seam)

  /// Remote flag evaluator. No-op today; swap for a PostHog-backed manager via
  /// `FeatureFlagsConfig(provider: "posthog", ...).makeFeatureFlagManager()` when adopted.
  static let manager: any FeatureFlagManager = NoopFeatureFlagManager()

  /// Evaluate a remote boolean flag for `userID`, treating any error as `false`.
  static func canUse(_ feature: String, userID: String) async -> Bool {
    (try? await manager.canUseFeature(feature, context: EvaluationContext(targetingKey: userID)))
      ?? false
  }
}
