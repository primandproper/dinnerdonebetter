//
//  HouseholdDetailsView.swift
//  ios
//

import SwiftUI

struct HouseholdDetailsView: View {
  @Environment(EventReporterService.self) private var eventReporterService
  let viewModel: AccountSettingsViewModel

  var body: some View {
    DSContentState(
      isLoading: viewModel.isLoading,
      loadingMessage: "Loading household...",
      error: viewModel.errorMessage,
      errorTitle: viewModel.errorTitle,
      errorIcon: viewModel.errorIcon,
      errorIconColor: viewModel.errorIconColor,
      onRetry: { await viewModel.loadData() },
      showEnvironmentSelector: viewModel.isServerDownError,
      content: { detailsContent }
    )
    .navigationTitle("Household Details")
    .refreshable {
      await viewModel.loadData()
    }
  }

  @ViewBuilder
  private var detailsContent: some View {
    if viewModel.account != nil {
      ScrollView {
        DSSection(
          "Household Details",
          subtitle: viewModel.canUpdateAccount
            ? nil : "Only household admins can edit household details"
        ) {
          VStack(spacing: DSTheme.Spacing.lg) {
            DSTextField(
              "Household Name",
              text: Binding(get: { viewModel.accountName }, set: { viewModel.accountName = $0 }),
              isDisabled: !viewModel.canUpdateAccount
            )

            DSTextField(
              "Contact Phone",
              text: Binding(get: { viewModel.contactPhone }, set: { viewModel.contactPhone = $0 }),
              type: .phone,
              isDisabled: !viewModel.canUpdateAccount
            )

            DSTextField(
              "Address Line 1",
              text: Binding(get: { viewModel.addressLine1 }, set: { viewModel.addressLine1 = $0 }),
              isDisabled: !viewModel.canUpdateAccount
            )

            DSTextField(
              "Address Line 2",
              text: Binding(get: { viewModel.addressLine2 }, set: { viewModel.addressLine2 = $0 }),
              isDisabled: !viewModel.canUpdateAccount
            )

            HStack(spacing: DSTheme.Spacing.md) {
              DSTextField(
                "City",
                text: Binding(get: { viewModel.city }, set: { viewModel.city = $0 }),
                isDisabled: !viewModel.canUpdateAccount
              )

              DSTextField(
                "State",
                text: Binding(get: { viewModel.state }, set: { viewModel.state = $0 }),
                isDisabled: !viewModel.canUpdateAccount
              )

              DSTextField(
                "Zip Code",
                text: Binding(get: { viewModel.zipCode }, set: { viewModel.zipCode = $0 }),
                type: .number,
                isDisabled: !viewModel.canUpdateAccount
              )
            }

            DSTextField(
              "Country",
              text: Binding(get: { viewModel.country }, set: { viewModel.country = $0 }),
              isDisabled: !viewModel.canUpdateAccount
            )

            DSButton(
              "Update Household",
              icon: "checkmark",
              fullWidth: true,
              isDisabled: !viewModel.canUpdateAccount || !viewModel.accountDataHasChanged
            ) {
              eventReporterService.reporter.track(
                event: "household_details_updated", properties: [:])
              Task {
                await viewModel.updateAccount()
              }
            }
          }
        }
        .dsScreenPadding()
      }
    }
  }
}

#Preview {
  let authManager = AuthenticationManager()
  authManager.isAuthenticated = true
  authManager.username = "John Doe"
  authManager.userID = "user123"
  authManager.accountID = "account123"

  return NavigationStack {
    HouseholdDetailsView(viewModel: AccountSettingsViewModel(authManager: authManager))
      .environment(authManager)
      .environment(EventReporterService())
  }
}
