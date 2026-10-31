//
//  LoginView.swift
//  ios
//
//  Created by Jeffrey Dorrycott on 12/8/25.
//

import DesignSystem
import SwiftUI

struct LoginView: View {
  @Environment(AuthenticationManager.self) private var authManager
  @Environment(EventReporterService.self) private var eventReporterService
  @Binding var showRegister: Bool

  @State private var username: String = ""
  @State private var password: String = ""
  @State private var totpCode: String = ""
  @State private var requiresTOTP: Bool = false
  @State private var errorMessage: String = ""
  @State private var isLoading: Bool = false
  @State private var loginTask: Task<Void, Never>?
  #if DEBUG
    @State private var showEnvironmentPicker: Bool = false
    @State private var selectedEnvironment: AppEnvironment = APIConfiguration.currentEnvironment
  #endif

  var body: some View {
    VStack(spacing: DSTheme.Spacing.xl) {
      Spacer()

      // App Title
      VStack(spacing: DSTheme.Spacing.sm) {
        Text(Branding.companyName)
          .font(DSTheme.Typography.largeTitle)
          .foregroundColor(DSTheme.Colors.textPrimary)

        Text("Sign in to continue")
          .font(DSTheme.Typography.body)
          .foregroundColor(DSTheme.Colors.textSecondary)
      }

      Spacer()

      // Login Form
      VStack(spacing: DSTheme.Spacing.lg) {
        DSTextField(
          "Username",
          text: $username,
          type: .username,
          isDisabled: isLoading
        )
        .accessibilityIdentifier("usernameTextField")

        DSTextField(
          "Password",
          text: $password,
          type: .password,
          isDisabled: isLoading
        )
        .accessibilityIdentifier("passwordTextField")

        if requiresTOTP {
          HStack(spacing: DSTheme.Spacing.sm) {
            DSTextField(
              "2FA Code",
              text: $totpCode,
              type: .number,
              isDisabled: isLoading
            )
            .accessibilityIdentifier("totpTextField")

            PasteButton(payloadType: String.self) { strings in
              if let code = strings.first?.trimmingCharacters(in: .whitespacesAndNewlines),
                !code.isEmpty
              {
                totpCode = code
                loginTask?.cancel()
                loginTask = Task { await handleLogin() }
              }
            }
            .labelStyle(.iconOnly)
            .disabled(isLoading)
            .accessibilityLabel("Paste 2FA code and sign in")
            .accessibilityIdentifier("totpPasteButton")
          }
        }

        if !errorMessage.isEmpty {
          Text(errorMessage)
            .font(DSTheme.Typography.caption)
            .foregroundColor(DSTheme.Colors.error)
            .multilineTextAlignment(.center)
            .accessibilityIdentifier("errorMessage")
        }

        DSButton(
          isLoading ? "Signing In..." : "Sign In",
          fullWidth: true,
          isLoading: isLoading,
          isDisabled: username.isEmpty || password.isEmpty
            || (requiresTOTP && totpCode.isEmpty)
        ) {
          eventReporterService.reporter.track(event: "login_started", properties: [:])
          loginTask?.cancel()
          loginTask = Task { await handleLogin() }
        }
        .accessibilityIdentifier("signInButton")
        .padding(.top, DSTheme.Spacing.sm)

        // Navigation to register
        Button {
          eventReporterService.reporter.track(event: "auth_switch_to_register", properties: [:])
          showRegister = true
        } label: {
          Text("Don't have an account? Sign up")
            .font(DSTheme.Typography.caption)
            .foregroundColor(DSTheme.Colors.primary)
        }
        .padding(.top, DSTheme.Spacing.sm)
      }
      .padding(.horizontal, DSTheme.Spacing.xxl)
      .animation(DSTheme.Animation.normal, value: requiresTOTP)

      Spacer()
      Spacer()

      #if DEBUG
        // Environment selector (Local vs Production) — dev builds only
        environmentButton
      #endif
    }
    .dsScreenPadding()
    #if DEBUG
      .sheet(isPresented: $showEnvironmentPicker) {
        EnvironmentPickerSheet(
          selectedEnvironment: $selectedEnvironment,
          onDismiss: { showEnvironmentPicker = false }
        )
        .presentationDetents([.medium])
      }
    #endif
  }

  #if DEBUG
    // MARK: - Environment Button

    private var environmentButton: some View {
      Button {
        showEnvironmentPicker = true
      } label: {
        HStack(spacing: DSTheme.Spacing.xs) {
          Image(systemName: selectedEnvironment.iconName)
            .font(.system(size: 12))
          Text(selectedEnvironment.displayName)
            .font(DSTheme.Typography.caption)
          Image(systemName: "chevron.up")
            .font(.system(size: 10, weight: .semibold))
        }
        .foregroundColor(DSTheme.Colors.textSecondary)
        .padding(.horizontal, DSTheme.Spacing.md)
        .padding(.vertical, DSTheme.Spacing.sm)
        .background(DSTheme.Colors.cardBackground)
        .cornerRadius(DSTheme.Radius.full)
      }
      .padding(.bottom, DSTheme.Spacing.sm)
    }
  #endif

  private func handleLogin() async {
    guard !Task.isCancelled else { return }

    await MainActor.run {
      errorMessage = ""
      isLoading = true
    }

    let totpToken = totpCode.isEmpty ? nil : totpCode
    let result = await authManager.login(
      username: username, password: password, totpToken: totpToken)

    guard !Task.isCancelled else { return }

    await MainActor.run {
      isLoading = false

      if result.success {
        username = ""
        password = ""
        totpCode = ""
        requiresTOTP = false
      } else {
        if result.requiresTOTP {
          requiresTOTP = true
          // Don't show error text when 2FA field first appears—the field itself is the prompt
          errorMessage = ""
        } else {
          errorMessage = result.error ?? "Unknown error occurred"
        }
      }
    }
  }
}

// MARK: - Environment Picker Sheet

struct EnvironmentPickerSheet: View {
  @Binding var selectedEnvironment: AppEnvironment
  let onDismiss: () -> Void

  var body: some View {
    NavigationStack {
      List {
        ForEach(AppEnvironment.allCases, id: \.self) { env in
          Button {
            selectedEnvironment = env
            APIConfiguration.currentEnvironment = env
            onDismiss()
          } label: {
            HStack(spacing: DSTheme.Spacing.md) {
              Image(systemName: env.iconName)
                .font(.system(size: 18))
                .foregroundColor(
                  env == selectedEnvironment ? DSTheme.Colors.primary : DSTheme.Colors.textSecondary
                )
                .frame(width: 28)

              VStack(alignment: .leading, spacing: DSTheme.Spacing.xxs) {
                Text(env.displayName)
                  .font(DSTheme.Typography.label)
                  .foregroundColor(DSTheme.Colors.textPrimary)

                Text(env.subtitle)
                  .font(DSTheme.Typography.caption)
                  .foregroundColor(DSTheme.Colors.textSecondary)
              }

              Spacer()

              if env == selectedEnvironment {
                Image(systemName: "checkmark")
                  .font(.system(size: 14, weight: .semibold))
                  .foregroundColor(DSTheme.Colors.primary)
              }
            }
            .contentShape(Rectangle())
          }
          .buttonStyle(.plain)
        }
      }
      .navigationTitle("Environment")
      .navigationBarTitleDisplayMode(.inline)
      .toolbar {
        ToolbarItem(placement: .cancellationAction) {
          Button("Done") { onDismiss() }
        }
      }
    }
  }
}

#Preview {
  LoginView(showRegister: .constant(true))
    .environment(AuthenticationManager())
    .environment(EventReporterService())
}
