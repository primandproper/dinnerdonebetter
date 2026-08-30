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

/// Represents a user-configurable setting with its current configuration (if any).
struct ConfigurableSetting: Identifiable {
  let id: String
  let setting: Settings_ServiceSetting
  let config: Settings_ServiceSettingConfiguration?
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
      let (settings, configs) = try await fetchSettingsAndConfigs()
      configurableSettings = mergeAndFilter(settings: settings, configs: configs)
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

  func saveSetting(serviceSetting: Settings_ServiceSetting, value: String) async -> Bool {
    if !serviceSetting.enumeration.isEmpty, !serviceSetting.enumeration.contains(value) {
      errorMessage = "Invalid value for \(serviceSetting.name)"
      return false
    }

    let config = configurableSettings.first { $0.setting.id == serviceSetting.id }?.config

    do {
      let (clientManager, metadata) = try await getClientManagerAndMetadata()

      if let config = config {
        var request = Settings_UpdateServiceSettingConfigurationRequest()
        request.serviceSettingConfigurationID = config.id
        var input = Settings_ServiceSettingConfigurationUpdateRequestInput()
        input.value = value
        request.input = input

        _ = try await clientManager.client.settings.updateServiceSettingConfiguration(
          request,
          metadata: metadata,
          options: clientManager.defaultCallOptions
        )
        updateSettingLocally(serviceSettingID: serviceSetting.id, value: value, config: nil)
        userSettingsService.updateValue(value, for: serviceSetting.name)
      } else {
        var request = Settings_CreateServiceSettingConfigurationRequest()
        var input = Settings_ServiceSettingConfigurationCreationRequestInput()
        input.serviceSettingID = serviceSetting.id
        input.value = value
        input.notes = ""
        request.input = input

        let response = try await clientManager.client.settings.createServiceSettingConfiguration(
          request,
          metadata: metadata,
          options: clientManager.defaultCallOptions
        )
        if response.hasCreated {
          updateSettingLocally(
            serviceSettingID: serviceSetting.id, value: value, config: response.created)
        } else {
          updateSettingLocally(serviceSettingID: serviceSetting.id, value: value, config: nil)
        }
        userSettingsService.updateValue(value, for: serviceSetting.name)
      }

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

  private func fetchSettingsAndConfigs() async throws -> (
    [Settings_ServiceSetting], [Settings_ServiceSettingConfiguration]
  ) {
    let (clientManager, metadata) = try await getClientManagerAndMetadata()

    async let settingsTask = fetchServiceSettings(clientManager: clientManager, metadata: metadata)
    async let configsTask = fetchUserConfigs(clientManager: clientManager, metadata: metadata)

    return (try await settingsTask, try await configsTask)
  }

  private func fetchServiceSettings(
    clientManager: ClientManager<HTTP2ClientTransport.TransportServices>,
    metadata: GRPCCore.Metadata
  ) async throws -> [Settings_ServiceSetting] {
    var request = Settings_GetServiceSettingsRequest()
    request.filter = QueryFilterMessage()

    let response = try await clientManager.client.settings.getServiceSettings(
      request,
      metadata: metadata,
      options: clientManager.defaultCallOptions
    )

    return response.results
  }

  private func fetchUserConfigs(
    clientManager: ClientManager<HTTP2ClientTransport.TransportServices>,
    metadata: GRPCCore.Metadata
  ) async throws -> [Settings_ServiceSettingConfiguration] {
    var request = Settings_GetServiceSettingConfigurationsForUserRequest()
    request.filter = QueryFilterMessage()

    let response = try await clientManager.client.settings.getServiceSettingConfigurationsForUser(
      request,
      metadata: metadata,
      options: clientManager.defaultCallOptions
    )

    return response.results
  }

  /// Updates a single setting's value in configurableSettings without a full reload.
  private func updateSettingLocally(
    serviceSettingID: String,
    value: String,
    config: Settings_ServiceSettingConfiguration?
  ) {
    guard let index = configurableSettings.firstIndex(where: { $0.setting.id == serviceSettingID })
    else {
      return
    }
    let existing = configurableSettings[index]
    let updated = ConfigurableSetting(
      id: existing.id,
      setting: existing.setting,
      config: config ?? existing.config,
      currentValue: value
    )
    configurableSettings[index] = updated
  }

  private func mergeAndFilter(
    settings: [Settings_ServiceSetting],
    configs: [Settings_ServiceSettingConfiguration]
  ) -> [ConfigurableSetting] {
    let activeAccountID = authManager.accountID
    let configsForActiveAccount = configs.filter { $0.belongsToAccount == activeAccountID }
    let configsBySettingID = Dictionary(
      uniqueKeysWithValues: configsForActiveAccount.map { ($0.serviceSetting.id, $0) }
    )

    return
      settings
      .filter { $0.type == "user" && !$0.adminsOnly }
      .map { setting in
        let config = configsBySettingID[setting.id]
        let currentValue: String
        if let config = config, !config.value.isEmpty {
          currentValue = config.value
        } else if setting.hasDefaultValue, !setting.defaultValue.isEmpty {
          currentValue = setting.defaultValue
        } else if !setting.enumeration.isEmpty {
          currentValue = setting.enumeration[0]
        } else {
          currentValue = ""
        }

        return ConfigurableSetting(
          id: setting.id,
          setting: setting,
          config: config,
          currentValue: currentValue
        )
      }
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
