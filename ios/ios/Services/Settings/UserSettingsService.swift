//
//  UserSettingsService.swift
//  ios
//
//  Created by Auto on 3/7/25.
//

import Foundation
import GRPCCore
import GRPCNIOTransportHTTP2
import SwiftUI

/// App-wide cache of user service setting values. Load once when authenticated; values
/// are accessible throughout the app. Updated when the user changes settings in Preferences.
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
      let configs = try await fetchUserConfigs(authManager: authManager)
      let activeAccountID = authManager.accountID
      var newValues: [String: String] = [:]
      for config in configs where config.belongsToAccount == activeAccountID {
        let name = config.serviceSetting.name
        if !name.isEmpty {
          newValues[name] = config.value.isEmpty ? config.serviceSetting.defaultValue : config.value
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

  private func fetchUserConfigs(authManager: AuthenticationManager) async throws
    -> [Settings_ServiceSettingConfiguration]
  {
    var request = Settings_GetServiceSettingConfigurationsForUserRequest()
    request.filter = QueryFilterMessage()

    let response = try await authManager.authenticatedCall(
      "getServiceSettingConfigurationsForUser", idempotent: true
    ) { client, metadata, options in
      try await client.settings.getServiceSettingConfigurationsForUser(
        request, metadata: metadata, options: options)
    }

    return response.results
  }
}
