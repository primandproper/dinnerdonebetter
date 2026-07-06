//
//  SegmentBackedEventReporter.swift
//  ios
//
//  Segment-backed EventReporter using platform-swift's thin URLSession Segment reporter
//  (no third-party SDK). Maps the app's identify/track/reset surface onto platform-swift's
//  Go-style addUser/eventOccurred/eventOccurredAnonymous actor.
//
//  Named `SegmentBackedEventReporter` (not `SegmentEventReporter`) so the platform-swift
//  `SegmentEventReporter` actor is reachable unqualified — the Analytics module also exports a
//  namespace type `Analytics`, so `Analytics.SegmentEventReporter` would resolve to that type.
//

import Analytics
import Foundation

/// App `EventReporter` that delegates to platform-swift's `SegmentEventReporter` actor, which
/// batches and POSTs to Segment's HTTP Tracking API over URLSession.
///
/// The platform reporter is stateless with respect to identity, so this type tracks the current
/// user id and a stable anonymous id itself (mirroring `BackendEventReporter`): `track` routes to
/// `eventOccurred` when identified, else `eventOccurredAnonymous`.
///
/// Behavioral note vs. the previous analytics-swift SDK: automatic application-lifecycle / screen
/// events are no longer emitted (the SDK provided those for free). Track those explicitly if needed.
final class SegmentBackedEventReporter: EventReporter {
  private let reporter: SegmentEventReporter
  private let anonymousID: String
  private let lock = NSLock()
  private var identifiedUserID: String?

  /// Persisted anonymous id, sharing the analytics layer's existing convention.
  private static let anonymousIDKey = "AnalyticsAnonymousID"

  init?(writeKey: String) {
    guard
      let reporter = try? SegmentEventReporter(
        writeKey: writeKey,
        circuitBreaker: PlatformServices.shared.breaker,
        observer: PlatformServices.shared.observer("SegmentEventReporter"))
    else {
      return nil
    }
    self.reporter = reporter
    self.anonymousID = Self.resolveAnonymousID()
  }

  func identify(userID: String, properties: [String: Any]) {
    let props = Self.encode(properties)
    lock.withLock { identifiedUserID = userID }
    Task { try? await reporter.addUser(userID: userID, properties: props) }
  }

  func track(event: String, properties: [String: Any]) {
    let props = Self.encode(properties)
    let userID = lock.withLock { identifiedUserID }
    let anonymousID = self.anonymousID
    Task {
      if let userID {
        try? await reporter.eventOccurred(event: event, userID: userID, properties: props)
      } else {
        try? await reporter.eventOccurredAnonymous(
          event: event, anonymousID: anonymousID, properties: props)
      }
    }
  }

  func reset() {
    lock.withLock { identifiedUserID = nil }
  }

  // MARK: - Helpers

  /// Coerces the app's loosely-typed properties into platform `AnalyticsPropertyValue`s. Only
  /// String/Int/Bool/Double are retained (matching the previous Segment-SDK behavior).
  private static func encode(_ dict: [String: Any]) -> [String: AnalyticsPropertyValue] {
    dict.compactMapValues { value in
      switch value {
      case let s as String: return .string(s)
      case let b as Bool: return .bool(b)
      case let i as Int: return .number(Double(i))
      case let d as Double: return .number(d)
      default: return nil
      }
    }
  }

  private static func resolveAnonymousID() -> String {
    let defaults = UserDefaults.standard
    if let existing = defaults.string(forKey: anonymousIDKey) {
      return existing
    }
    let generated = UUID().uuidString
    defaults.set(generated, forKey: anonymousIDKey)
    return generated
  }
}
