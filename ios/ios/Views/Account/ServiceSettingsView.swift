//
//  ServiceSettingsView.swift
//  ios
//
//  Created by Auto on 3/7/25.
//

import SwiftUI

struct ServiceSettingsView: View {
  @Environment(AuthenticationManager.self) private var authManager
  @Environment(UserSettingsService.self) private var userSettingsService
  @State private var viewModel: ServiceSettingsViewModel?

  var body: some View {
    Group {
      if let viewModel = viewModel {
        DSContentState(
          isLoading: viewModel.isLoading,
          loadingMessage: "Loading preferences...",
          error: viewModel.errorMessage,
          errorTitle: viewModel.errorTitle,
          errorIcon: viewModel.errorIcon,
          errorIconColor: viewModel.errorIconColor,
          onRetry: { await viewModel.loadData() },
          showEnvironmentSelector: viewModel.isServerDownError,
          content: { settingsContent(viewModel: viewModel) }
        )
      } else {
        DSInitializingView()
      }
    }
    .navigationTitle("Preferences")
    .refreshable {
      if let viewModel = viewModel {
        await viewModel.loadData()
      }
    }
    .onAppear {
      if viewModel == nil {
        viewModel = ServiceSettingsViewModel(
          authManager: authManager, userSettingsService: userSettingsService)
      }
      if let viewModel = viewModel {
        Task {
          await viewModel.loadData()
        }
      }
    }
  }

  @ViewBuilder
  private func settingsContent(viewModel: ServiceSettingsViewModel) -> some View {
    ScrollView {
      if viewModel.configurableSettings.isEmpty {
        emptyState
      } else {
        DSSection("Preferences") {
          VStack(spacing: DSTheme.Spacing.sm) {
            ForEach(viewModel.configurableSettings) { item in
              settingRow(item: item, viewModel: viewModel)
            }
          }
        }
        .dsScreenPadding()
      }
    }
  }

  @ViewBuilder
  private func settingRow(item: ConfigurableSetting, viewModel: ServiceSettingsViewModel)
    -> some View
  {
    if !item.setting.enumeration.isEmpty {
      DSListRow(
        title: humanReadableName(for: item.setting.name),
        subtitle: item.setting.description_p.isEmpty ? nil : item.setting.description_p,
        icon: iconForSetting(item.setting.name),
        style: .card,
        trailing: {
          settingPicker(item: item, viewModel: viewModel)
        }
      )
    }
    // Settings with empty enumeration are skipped per plan
  }

  private var emptyState: some View {
    VStack(spacing: DSTheme.Spacing.lg) {
      Image(systemName: "slider.horizontal.3")
        .font(.system(size: DSTheme.IconSize.xxl))
        .foregroundColor(DSTheme.Colors.textTertiary)

      Text("No preferences to configure")
        .font(DSTheme.Typography.body)
        .foregroundColor(DSTheme.Colors.textSecondary)
        .multilineTextAlignment(.center)
    }
    .frame(maxWidth: .infinity)
    .padding(DSTheme.Spacing.xxl)
  }

  private func humanReadableName(for name: String) -> String {
    name
      .replacingOccurrences(of: "_", with: " ")
      .capitalized
  }

  private func humanReadableOption(_ value: String) -> String {
    value.capitalized
  }

  private func iconForSetting(_ name: String) -> String {
    if name.contains("temperature") {
      return "thermometer"
    }
    return "gearshape"
  }

  @ViewBuilder
  private func settingPicker(item: ConfigurableSetting, viewModel: ServiceSettingsViewModel)
    -> some View
  {
    let selection = Binding<String>(
      get: {
        let valid = item.setting.enumeration.contains(item.currentValue)
        return valid ? item.currentValue : (item.setting.enumeration.first ?? "")
      },
      set: { newValue in
        Task {
          _ = await viewModel.saveSetting(definition: item.setting, value: newValue)
        }
      }
    )

    let pickerContent = Picker("", selection: selection) {
      ForEach(item.setting.enumeration, id: \.self) { option in
        Text(humanReadableOption(option)).tag(option)
      }
    }

    if item.setting.enumeration.count <= 3 {
      pickerContent.pickerStyle(.segmented)
    } else {
      pickerContent.pickerStyle(.menu)
    }
  }
}

#Preview {
  let authManager = AuthenticationManager()
  authManager.isAuthenticated = true
  authManager.username = "John Doe"
  authManager.userID = "user123"
  authManager.accountID = "account123"
  let userSettingsService = UserSettingsService()
  userSettingsService.configure(authManager: authManager)

  return NavigationStack {
    ServiceSettingsView()
      .environment(authManager)
      .environment(userSettingsService)
  }
}
