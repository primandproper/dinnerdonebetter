//
//  UserSettingsService.swift
//  ios
//
//  Created by Auto on 3/7/25.
//

import Foundation
import GRPCCore
import GRPCNIOTransportHTTP2
import PlatformClient
import SwiftUI

/// App-wide cache of the signed-in user's resolved setting values. Load once when
/// authenticated; values are accessible throughout the app. Updated when the user
/// changes settings in Preferences.
@Observable
@MainActor
class UserSettingsService {
  /// Setting name -> value. Empty until load() succeeds.
  private(set) var values: [String: String] = [:]

  /// True after a successful load. Stays true until clear() or load fails.
  private(set) var isLoaded = false

  private weak var authManager: AuthenticationManager?

  init() {}

  /// Configure with auth manager. Call once when the app has auth available (e.g. from iosApp).
  func configure(authManager: AuthenticationManager) {
    self.authManager = authManager
  }

  /// Load user settings from the API. Call when user is authenticated. Safe to call multiple times.
  func load() async {
    guard let authManager = authManager, authManager.isAuthenticated else {
      return
    }

    do {
      let resolutions = try await fetchResolvedSettings(authManager: authManager)
      var newValues: [String: String] = [:]
      for resolution in resolutions {
        let name = resolution.definition.name
        // "unset" is a setting nobody has answered that has no default, which is
        // not the same as an answer of "". Leaving it out of the cache is what
        // lets value(for:default:) hand back the caller's own fallback.
        if !name.isEmpty, let raw = resolution.typedValue.text {
          newValues[name] = raw
        }
      }
      values = newValues
      isLoaded = true
    } catch {
      // Keep existing values on error; don't clear
      print("❌ UserSettingsService: Failed to load: \(error)")
    }
  }

  /// Get the value for a setting. Returns default if not set or not loaded.
  func value(for settingName: String, default defaultValue: String = "") -> String {
    values[settingName] ?? defaultValue
  }

  /// Update a setting value locally (e.g. after user saves in Preferences). Also persists to API
  /// via the caller; this method only updates the cache.
  func updateValue(_ value: String, for settingName: String) {
    values[settingName] = value
  }

  /// Clear cached values. Call on logout.
  func clear() {
    values = [:]
    isLoaded = false
  }

  /// Fetch every setting resolved for the signed-in user: their own answer where
  /// they have one, the setting's default where they have not, in one call.
  ///
  /// The server applies the fallback, which is why this no longer pairs a catalog
  /// against a list of stored values. It also decides which settings the user may
  /// see, so an admin-only setting is simply absent rather than filtered here.
  private func fetchResolvedSettings(authManager: AuthenticationManager) async throws
    -> [ResolvedSetting]
  {
    var request = Primandproper_Platform_Settings_V1_ResolveAllRequest()
    request.subject = SettingValues.subject(userID: authManager.userID)

    let response = try await authManager.authenticatedCall(
      "resolveAll", idempotent: true
    ) { client, metadata, options in
      try await client.settings.resolveAll(request, metadata: metadata, options: options)
    }

    return response.resolutions
  }
}
