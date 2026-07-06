//
//  AnalyticsConfiguration.swift
//  ios
//
//  Configuration and factory for EventReporter.
//  Segment write key is read from Info.plist (Secrets.xcconfig -> $(SEGMENT_WRITE_KEY)).
//

import Foundation

enum AnalyticsConfiguration {
  private static let placeholderWriteKey = "your-segment-write-key-here"

  /// When true, use BackendEventReporter (analytics passthrough gRPC). When false, use Segment.
  /// Read from Info.plist (UseAnalyticsBackend), injected via Secrets.xcconfig (USE_ANALYTICS_BACKEND).
  /// When the key is absent from Secrets.xcconfig, defaults to false.
  static var useAnalyticsBackend: Bool {
    guard let value = Bundle.main.infoDictionary?["UseAnalyticsBackend"] else { return false }
    if let bool = value as? Bool { return bool }
    let string = (value as? String)?.trimmingCharacters(in: .whitespaces) ?? ""
    return !string.isEmpty && (string.lowercased() == "true" || string == "1")
  }

  /// Segment write key from Info.plist. Injected via Secrets.xcconfig at build time.
  static var segmentWriteKey: String {
    Bundle.main.infoDictionary?["SegmentWriteKey"] as? String ?? ""
  }

  /// Shared EventReporter instance. Cached so we don't create multiple Segment clients.
  private static var _sharedReporter: (any EventReporter)?
  private static let reporterLock = NSLock()

  /// Returns the shared EventReporter. Uses Backend when useAnalyticsBackend is true (and authManager provided),
  /// else Segment when write key is valid, else Noop.
  static func provideEventReporter(authManager: AuthenticationManaging? = nil) -> any EventReporter
  {
    reporterLock.lock()
    defer { reporterLock.unlock() }
    if let existing = _sharedReporter {
      return existing
    }
    let reporter: any EventReporter
    if useAnalyticsBackend, let auth = authManager {
      reporter = BackendEventReporter(authManager: auth)
    } else {
      let key = segmentWriteKey.trimmingCharacters(in: .whitespaces)
      if !key.isEmpty, key != placeholderWriteKey,
        let segment = SegmentBackedEventReporter(writeKey: key)
      {
        reporter = segment
      } else {
        reporter = NoopEventReporter()
      }
    }
    _sharedReporter = reporter
    return reporter
  }
}
