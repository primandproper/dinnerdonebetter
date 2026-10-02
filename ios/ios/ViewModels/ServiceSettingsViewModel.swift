//
//  ServiceSettingsViewModel.swift
//  ios
//
//  Created by Auto on 3/7/25.
//

import Foundation
import GRPCCore
import GRPCNIOTransportHTTP2
import PlatformClient
import SwiftProtobuf
import SwiftUI

/// A setting the signed-in user can change, and the value that currently applies
/// to them — their own answer, or the setting's default where they have not
/// answered.
struct ConfigurableSetting: Identifiable {
  let id: String
  let setting: SettingDefinition
  let currentValue: String
}

@Observable
@MainActor
class ServiceSettingsViewModel {
  // Data
  var configurableSettings: [ConfigurableSetting] = []

  // Loading states
  var isLoading = false
  var errorMessage: String?
  var errorTitle: String = "Error"
  var errorIcon: String = "exclamationmark.triangle"
  var errorIconColor = DSTheme.Colors.warning
  var isServerDownError = false

  private let authManager: AuthenticationManager
  private let userSettingsService: UserSettingsService

  init(authManager: AuthenticationManager, userSettingsService: UserSettingsService) {
    self.authManager = authManager
    self.userSettingsService = userSettingsService
  }

  func loadData() async {
    isLoading = true
    errorMessage = nil
    errorTitle = "Error"
    errorIcon = "exclamationmark.triangle"
    errorIconColor = DSTheme.Colors.warning
    isServerDownError = false

    do {
      configurableSettings = try await fetchResolvedSettings().map(configurableSetting(from:))
    } catch {
      await authManager.invalidateCredentialsIfSessionError(error)
      let display = ErrorDisplayFormatter.format(error, context: "load settings")
      errorMessage = display.message
      errorTitle = display.title
      errorIcon = display.icon
      errorIconColor = display.iconColor
      isServerDownError = ErrorDisplayFormatter.isServerDown(error)
      print("❌ Error loading service settings: \(error)")
    }

    isLoading = false
  }

  /// Store the user's answer to one setting.
  ///
  /// There is one call for it whether or not they had answered before: the server
  /// converges on the row, so a first answer and a changed one are the same write
  /// and this no longer has to know which it is making.
  func saveSetting(definition: SettingDefinition, value: String) async -> Bool {
    if !definition.enumeration.isEmpty, !definition.enumeration.contains(value) {
      errorMessage = "Invalid value for \(definition.name)"
      return false
    }

    let typed: Primandproper_Platform_Settings_V1_TypedValue
    do {
      typed = try .init(text: value, kind: definition.kind)
    } catch {
      errorMessage = "Invalid value for \(definition.name): \(error)"
      return false
    }

    do {
      var request = Primandproper_Platform_Settings_V1_SetValueRequest()
      request.subject = SettingValues.subject(userID: authManager.userID)
      request.name = definition.name
      request.value = typed

      _ = try await authManager.authenticatedCall("setValue") { client, metadata, options in
        try await client.settings.setValue(request, metadata: metadata, options: options)
      }

      updateSettingLocally(settingID: definition.id, value: value)
      userSettingsService.updateValue(value, for: definition.name)

      return true
    } catch {
      await authManager.invalidateCredentialsIfSessionError(error)
      let display = ErrorDisplayFormatter.format(error, context: "save setting")
      errorMessage = display.message
      errorTitle = display.title
      errorIcon = display.icon
      errorIconColor = display.iconColor
      isServerDownError = ErrorDisplayFormatter.isServerDown(error)
      print("❌ Error saving setting: \(error)")
      return false
    }
  }

  /// Fetch every setting resolved for the signed-in user, in one call.
  ///
  /// The catalog and the user's answers used to be two requests joined here, with
  /// the fallback to a setting's default reimplemented alongside. The server does
  /// both now, and it also decides which settings this user may see — so the
  /// admin-only ones are absent rather than filtered out below.
  private func fetchResolvedSettings() async throws -> [ResolvedSetting] {
    var request = Primandproper_Platform_Settings_V1_ResolveAllRequest()
    request.subject = SettingValues.subject(userID: authManager.userID)

    let response = try await authManager.authenticatedCall("resolveAll", idempotent: true) {
      client, metadata, options in
      try await client.settings.resolveAll(request, metadata: metadata, options: options)
    }

    return response.resolutions
  }

  /// Pair one resolution with the value the picker should start on.
  ///
  /// A resolution whose source is "unset" is a setting nobody has answered that
  /// has no default. There is no value to show, so the first enumerated option
  /// stands in — which is what the picker would have to fall back to anyway.
  private func configurableSetting(from resolution: ResolvedSetting) -> ConfigurableSetting {
    let definition = resolution.definition
    let currentValue = resolution.typedValue.text ?? (definition.enumeration.first ?? "")

    return ConfigurableSetting(
      id: definition.id,
      setting: definition,
      currentValue: currentValue
    )
  }

  /// Updates a single setting's value in configurableSettings without a full reload.
  private func updateSettingLocally(settingID: String, value: String) {
    guard let index = configurableSettings.firstIndex(where: { $0.setting.id == settingID })
    else {
      return
    }
    let existing = configurableSettings[index]
    configurableSettings[index] = ConfigurableSetting(
      id: existing.id,
      setting: existing.setting,
      currentValue: value
    )
  }

}
