//
//  RegisterView.swift
//  ios
//
//  Created by Auto on 12/8/25.
//

import SwiftUI

struct RegisterView: View {
  @Environment(AuthenticationManager.self) private var authManager
  @Environment(EventReporterService.self) private var eventReporterService
  @Binding var showLogin: Bool

  // Invitation data passed from deep link (immutable once set)
  let invitationID: String
  let invitationToken: String

  @State private var emailAddress: String = ""
  @State private var username: String = ""
  @State private var password: String = ""
  @State private var repeatedPassword: String = ""
  @State private var accountName: String = ""
  @State private var firstName: String = ""
  @State private var lastName: String = ""
  @State private var errorMessage: String = ""
  @State private var isLoading: Bool = false
  @State private var registrationTask: Task<Void, Never>?

  /// Whether this registration is via an invitation link
  private var isInvitedRegistration: Bool {
    !invitationID.isEmpty && !invitationToken.isEmpty
  }

  init(showLogin: Binding<Bool>, invitationID: String = "", invitationToken: String = "") {
    self._showLogin = showLogin
    self.invitationID = invitationID
    self.invitationToken = invitationToken
  }

  var body: some View {
    ScrollView {
      VStack(spacing: DSTheme.Spacing.xl) {
        Spacer()
          .frame(height: DSTheme.Spacing.xl)

        // App Title
        VStack(spacing: DSTheme.Spacing.sm) {
          Text(Branding.companyName)
            .font(DSTheme.Typography.largeTitle)
            .foregroundColor(DSTheme.Colors.textPrimary)

          Text(isInvitedRegistration ? "Accept your invitation" : "Create your account")
            .font(DSTheme.Typography.body)
            .foregroundColor(DSTheme.Colors.textSecondary)
        }

        // Show invitation banner if from deep link
        if isInvitedRegistration {
          HStack {
            Image(systemName: "envelope.badge")
              .foregroundColor(DSTheme.Colors.primary)
            Text("You've been invited to join an account!")
              .font(DSTheme.Typography.caption)
              .foregroundColor(DSTheme.Colors.textSecondary)
          }
          .padding(DSTheme.Spacing.md)
          .background(DSTheme.Colors.primary.opacity(0.1))
          .cornerRadius(DSTheme.Radius.sm)
        }

        // Registration Form
        VStack(spacing: DSTheme.Spacing.lg) {
          // Required fields
          DSTextField(
            "Email Address",
            text: $emailAddress,
            type: .email,
            isDisabled: isLoading
          )
          .accessibilityIdentifier("registrationEmailAddressInput")

          DSTextField(
            "Username",
            text: $username,
            type: .username,
            isDisabled: isLoading
          )
          .accessibilityIdentifier("registrationUsernameInput")

          DSTextField(
            "Password",
            text: $password,
            type: .password,
            isDisabled: isLoading
          )
          .accessibilityIdentifier("registrationPasswordInput")

          DSTextField(
            "Password (again)",
            text: $repeatedPassword,
            type: .password,
            isDisabled: isLoading
          )
          .accessibilityIdentifier("registrationPasswordConfirmInput")

          // Divider for optional fields
          HStack {
            Rectangle()
              .fill(DSTheme.Colors.border)
              .frame(height: 1)
            Text("optional fields")
              .font(DSTheme.Typography.caption)
              .foregroundColor(DSTheme.Colors.textSecondary)
            Rectangle()
              .fill(DSTheme.Colors.border)
              .frame(height: 1)
          }
          .padding(.vertical, DSTheme.Spacing.sm)

          // Optional fields
          DSTextField(
            "Account Name",
            text: $accountName,
            isDisabled: isLoading
          )
          .accessibilityIdentifier("registrationAccountNameInput")

          // The birthday field is gone with the column behind it. platform's user has
          // none, nothing on the backend ever read one, and storing it means a table of
          // this application's own — which is its own piece of work rather than part of
          // the directory adoption.

          HStack(spacing: DSTheme.Spacing.md) {
            DSTextField(
              "First Name",
              text: $firstName,
              isDisabled: isLoading
            )
            .accessibilityIdentifier("registrationFirstNameInput")

            DSTextField(
              "Last Name",
              text: $lastName,
              isDisabled: isLoading
            )
            .accessibilityIdentifier("registrationLastNameInput")
          }

          if !errorMessage.isEmpty {
            Text(errorMessage)
              .font(DSTheme.Typography.caption)
              .foregroundColor(DSTheme.Colors.error)
              .multilineTextAlignment(.center)
              .accessibilityIdentifier("errorMessage")
          }

          DSButton(
            isLoading ? "Registering..." : "Register",
            fullWidth: true,
            isLoading: isLoading,
            isDisabled: !isFormValid
          ) {
            eventReporterService.reporter.track(
              event: "register_started",
              properties: isInvitedRegistration ? ["from_invite": true] : [:]
            )
            registrationTask?.cancel()
            registrationTask = Task { await handleRegistration() }
          }
          .accessibilityIdentifier("registrationButton")
          .padding(.top, DSTheme.Spacing.sm)

          // Navigation to login
          Button {
            eventReporterService.reporter.track(event: "auth_switch_to_login", properties: [:])
            showLogin = true
          } label: {
            Text("Already have an account? Sign in")
              .font(DSTheme.Typography.caption)
              .foregroundColor(DSTheme.Colors.primary)
          }
          .padding(.top, DSTheme.Spacing.sm)
        }
        .padding(.horizontal, DSTheme.Spacing.xxl)

        Spacer()
          .frame(height: DSTheme.Spacing.xl)
      }
    }
    .dsScreenPadding()
  }

  // MARK: - Validation

  private var isFormValid: Bool {
    !emailAddress.trimmingCharacters(in: .whitespaces).isEmpty
      && !username.trimmingCharacters(in: .whitespaces).isEmpty
      && !password.isEmpty
      && !repeatedPassword.isEmpty
      && password == repeatedPassword
      && password.count >= 8
      && isValidEmail(emailAddress)
  }

  private func isValidEmail(_ email: String) -> Bool {
    let emailRegex = "[A-Z0-9a-z._%+-]+@[A-Za-z0-9.-]+\\.[A-Za-z]{2,64}"
    let emailPredicate = NSPredicate(format: "SELF MATCHES %@", emailRegex)
    return emailPredicate.evaluate(with: email)
  }

  // MARK: - Registration

  private func handleRegistration() async {
    guard !Task.isCancelled else { return }

    await MainActor.run {
      errorMessage = ""
      isLoading = true
    }

    if password != repeatedPassword {
      await MainActor.run {
        errorMessage = "Passwords do not match"
        isLoading = false
      }
      return
    }

    if password.count < 8 {
      await MainActor.run {
        errorMessage = "Password must have at least 8 characters"
        isLoading = false
      }
      return
    }

    if !isValidEmail(emailAddress) {
      await MainActor.run {
        errorMessage = "Invalid email address"
        isLoading = false
      }
      return
    }

    let registrationInput = RegistrationInput(
      emailAddress: emailAddress.trimmingCharacters(in: .whitespaces),
      username: username.trimmingCharacters(in: .whitespaces),
      password: password,
      accountName: accountName.trimmingCharacters(in: .whitespaces),
      firstName: firstName.trimmingCharacters(in: .whitespaces),
      lastName: lastName.trimmingCharacters(in: .whitespaces),
      invitationToken: invitationToken.trimmingCharacters(in: .whitespaces),
      invitationID: invitationID.trimmingCharacters(in: .whitespaces)
    )
    let result = await authManager.register(input: registrationInput)

    guard !Task.isCancelled else { return }

    await MainActor.run {
      isLoading = false

      if result.success {
        eventReporterService.reporter.track(event: "register_succeeded", properties: [:])
        clearForm()
        showLogin = true
      } else {
        eventReporterService.reporter.track(
          event: "register_failed",
          properties: ["error": result.error ?? "Unknown error occurred"]
        )
        errorMessage = result.error ?? "Unknown error occurred"
      }
    }
  }

  private func clearForm() {
    emailAddress = ""
    username = ""
    password = ""
    repeatedPassword = ""
    accountName = ""
    firstName = ""
    lastName = ""
  }
}

#Preview("Standard Registration") {
  RegisterView(showLogin: .constant(false))
    .environment(AuthenticationManager())
    .environment(EventReporterService())
}

#Preview("Invitation Registration") {
  RegisterView(
    showLogin: .constant(false),
    invitationID: "abc123",
    invitationToken: "token456"
  )
  .environment(AuthenticationManager())
  .environment(EventReporterService())
}
