//
//  BackendEventReporter.swift
//  ios
//
//  EventReporter that sends events to the backend's analytics passthrough gRPC service.
//  Uses source "ios" for the backend's MultiSourceEventReporter.
//

import Foundation
import GRPCNIOTransportHTTP2TransportServices
import PlatformClient

private let analyticsSource = "ios"
private let anonymousIDKey = "AnalyticsAnonymousID"

/// EventReporter that forwards events to the backend's analytics passthrough gRPC service.
/// Requires backend proxySources.ios to be configured.
final class BackendEventReporter: EventReporter {
  private let authManager: AuthenticationManaging
  private let lock = NSLock()

  init(authManager: AuthenticationManaging) {
    self.authManager = authManager
  }

  func identify(userID: String, properties: [String: Any]) {
    // Backend derives user from session for TrackEvent; no server-side identify call.
  }

  func track(event: String, properties: [String: Any]) {
    let props = Self.encodeToProperties(properties)
    let isAuthenticated = authManager.isAuthenticated

    Task {
      await sendEvent(event: event, properties: props ?? [:], isAuthenticated: isAuthenticated)
    }
  }

  func reset() {
    lock.lock()
    defer { lock.unlock() }
    UserDefaults.standard.removeObject(forKey: anonymousIDKey)
  }

  // MARK: - Private

  private func sendEvent(
    event: String,
    properties: [String: String],
    isAuthenticated: Bool
  ) async {
    do {
      let manager = try authManager.getClientManager()

      if isAuthenticated {
        var request = Analytics_TrackEventRequest()
        request.source = analyticsSource
        request.event = event
        request.properties = properties

        let client = manager.client
        let options = manager.defaultCallOptions
        _ = try await manager.session.call { metadata in
          try await client.analytics.trackEvent(request, metadata: metadata, options: options)
        }
      } else {
        let anonymousID = getOrCreateAnonymousID()
        var request = Analytics_TrackAnonymousEventRequest()
        request.source = analyticsSource
        request.event = event
        request.anonymousID = anonymousID
        request.properties = properties

        _ = try await manager.client.analytics.trackAnonymousEvent(
          request,
          metadata: [:],
          options: manager.defaultCallOptions
        )
      }
    } catch {
      print("⚠️ BackendEventReporter: Failed to track event '\(event)': \(error)")
    }
  }

  private func getOrCreateAnonymousID() -> String {
    lock.lock()
    defer { lock.unlock() }

    if let existing = UserDefaults.standard.string(forKey: anonymousIDKey), !existing.isEmpty {
      return existing
    }
    let id = UUID().uuidString
    UserDefaults.standard.set(id, forKey: anonymousIDKey)
    return id
  }

  private static func encodeToProperties(_ dict: [String: Any]) -> [String: String]? {
    let converted = dict.compactMapValues { value -> String? in
      if let s = value as? String { return s }
      if let n = value as? Int { return String(n) }
      if let b = value as? Bool { return b ? "true" : "false" }
      return nil
    }
    return converted.isEmpty ? nil : converted
  }
}
