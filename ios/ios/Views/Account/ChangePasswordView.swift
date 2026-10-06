//
//  ChangePasswordView.swift
//  ios
//
//  The change-password form. Reached from account settings, and shown in place of the app when
//  an operator has forced a password change: the server refuses every other call until then.
//

import SwiftUI

struct ChangePasswordView: View {
  @Environment(AuthenticationManager.self) private var authManager
  @Environment(EventReporterService.self) private var eventReporterService
  @State private var viewModel: ChangePasswordViewModel?

  /// True when the server is holding this login at the form. There is nowhere else to go then,
  /// so the only other way out is signing out.
  var isRequired: Bool = false

  var body: some View {
    ScrollView {
      VStack(spacing: DSTheme.Spacing.xl) {
        if let viewModel {
          form(viewModel)
        } else {
          DSInitializingView()
        }
      }
      .dsScreenPadding()
      .padding(.bottom, DSTheme.Spacing.lg)
    }
    .navigationTitle("Change Password")
    .onAppear {
      if viewModel == nil {
        viewModel = ChangePasswordViewModel(authManager: authManager)
      }
      eventReporterService.reporter.track(
        event: "change_password_viewed", properties: ["required": isRequired ? "true" : "false"])
    }
  }

  @ViewBuilder
  private func form(_ viewModel: ChangePasswordViewModel) -> some View {
    DSSection(
      "Change Password",
      subtitle: isRequired
        ? "Your password has to be changed before you can continue."
        : nil
    ) {
      VStack(spacing: DSTheme.Spacing.lg) {
        DSTextField(
          "Current Password",
          text: Binding(
            get: { viewModel.currentPassword }, set: { viewModel.currentPassword = $0 }),
          type: .password,
          isDisabled: viewModel.isSaving
        )
        .accessibilityIdentifier("currentPasswordTextField")

        DSTextField(
          "New Password",
          text: Binding(get: { viewModel.newPassword }, set: { viewModel.newPassword = $0 }),
          type: .password,
          isDisabled: viewModel.isSaving
        )
        .accessibilityIdentifier("newPasswordTextField")

        DSTextField(
          "Confirm New Password",
          text: Binding(
            get: { viewModel.confirmPassword }, set: { viewModel.confirmPassword = $0 }),
          type: .password,
          isDisabled: viewModel.isSaving
        )
        .accessibilityIdentifier("confirmPasswordTextField")

        if viewModel.requiresTOTP {
          DSTextField(
            "2FA Code",
            text: Binding(get: { viewModel.totpCode }, set: { viewModel.totpCode = $0 }),
            type: .number,
            isDisabled: viewModel.isSaving
          )
          .accessibilityIdentifier("changePasswordTOTPTextField")
        }

        if let message = viewModel.errorMessage {
          Text(message)
            .font(DSTheme.Typography.caption)
            .foregroundColor(DSTheme.Colors.error)
            .multilineTextAlignment(.center)
        } else if viewModel.didChange {
          Text("Your password has been changed.")
            .font(DSTheme.Typography.caption)
            .foregroundColor(DSTheme.Colors.success)
        }

        DSButton(
          "Change Password",
          icon: "key",
          fullWidth: true,
          isLoading: viewModel.isSaving,
          isDisabled: !viewModel.canSubmit
        ) {
          Task {
            if await viewModel.submit() {
              eventReporterService.reporter.track(event: "password_changed", properties: [:])
            }
          }
        }

        if isRequired {
          DSButton("Sign Out", style: .ghost, fullWidth: true) {
            Task { await authManager.logout() }
          }
        }
      }
    }
  }
}
