//
//  ServiceSettingsViewModel.swift
//  ios
//
//  Created by Auto on 3/7/25.
//

import Foundation
import GRPCCore
import GRPCNIOTransportHTTP2
import SwiftProtobuf
import SwiftUI

/// A setting the signed-in user can change, and the value that currently applies
/// to them — their own answer, or the setting's default where they have not
/// answered.
struct ConfigurableSetting: Identifiable {
  let id: String
  let setting: Settings_SettingDefinition
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
  func saveSetting(definition: Settings_SettingDefinition, value: String) async -> Bool {
    if !definition.enumeration.isEmpty, !definition.enumeration.contains(value) {
      errorMessage = "Invalid value for \(definition.name)"
      return false
    }

    do {
      let (clientManager, metadata) = try await getClientManagerAndMetadata()

      var request = Settings_SetSettingValueRequest()
      request.settingName = definition.name
      request.value = value

      _ = try await clientManager.client.settings.setSettingValue(
        request,
        metadata: metadata,
        options: clientManager.defaultCallOptions
      )

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
  private func fetchResolvedSettings() async throws -> [Settings_SettingResolution] {
    let (clientManager, metadata) = try await getClientManagerAndMetadata()

    let request = Settings_ResolveSettingsRequest()

    let response = try await clientManager.client.settings.resolveSettings(
      request,
      metadata: metadata,
      options: clientManager.defaultCallOptions
    )

    return response.results
  }

  /// Pair one resolution with the value the picker should start on.
  ///
  /// A resolution whose source is "unset" is a setting nobody has answered that
  /// has no default. There is no value to show, so the first enumerated option
  /// stands in — which is what the picker would have to fall back to anyway.
  private func configurableSetting(from resolution: Settings_SettingResolution)
    -> ConfigurableSetting
  {
    let definition = resolution.definition
    let currentValue =
      resolution.source == "unset" ? (definition.enumeration.first ?? "") : resolution.raw

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

  private func getClientManagerAndMetadata() async throws -> (
    ClientManager<HTTP2ClientTransport.TransportServices>, GRPCCore.Metadata
  ) {
    guard let clientManager = try? authManager.getClientManager() else {
      throw NSError(
        domain: "ServiceSettingsViewModel", code: 1,
        userInfo: [NSLocalizedDescriptionKey: "Failed to get client manager"])
    }

    guard let oauth2Token = await authManager.getOAuth2AccessToken() else {
      throw NSError(
        domain: "ServiceSettingsViewModel", code: 2,
        userInfo: [NSLocalizedDescriptionKey: "Failed to get OAuth2 access token"])
    }

    let metadata = clientManager.authenticatedMetadata(accessToken: oauth2Token)
    return (clientManager, metadata)
  }
}
