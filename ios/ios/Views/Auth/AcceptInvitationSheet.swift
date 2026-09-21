//
//  AcceptInvitationSheet.swift
//  ios
//
//  Shown when a logged-in user taps an invite link. Lets them accept and join the household.
//

import SwiftUI

struct AcceptInvitationSheet: View {
  @Environment(EventReporterService.self) private var eventReporterService
  @Environment(AuthenticationManager.self) private var authManager
  @Environment(\.dismiss) private var dismiss

  let invitationID: String
  let invitationToken: String
  let onAccepted: () -> Void

  @State private var isLoading = false
  @State private var errorMessage: String?
  @State private var didAccept = false

  var body: some View {
    NavigationStack {
      VStack(spacing: DSTheme.Spacing.xl) {
        if didAccept {
          VStack(spacing: DSTheme.Spacing.md) {
            Image(systemName: "checkmark.circle.fill")
              .font(.system(size: 60))
              .foregroundColor(DSTheme.Colors.primary)
            Text("You've joined the household!")
              .font(DSTheme.Typography.title2)
              .foregroundColor(DSTheme.Colors.textPrimary)
            Text("You can switch to it from My Household in account settings.")
              .font(DSTheme.Typography.body)
              .foregroundColor(DSTheme.Colors.textSecondary)
              .multilineTextAlignment(.center)
          }
          .padding(DSTheme.Spacing.xl)
        } else {
          VStack(spacing: DSTheme.Spacing.lg) {
            Image(systemName: "envelope.badge.fill")
              .font(.system(size: 48))
              .foregroundColor(DSTheme.Colors.primary)

            Text("You've been invited to join a household")
              .font(DSTheme.Typography.title2)
              .foregroundColor(DSTheme.Colors.textPrimary)
              .multilineTextAlignment(.center)

            Text(
              "Accept to add this household to your account. You can switch between households in account settings."
            )
            .font(DSTheme.Typography.body)
            .foregroundColor(DSTheme.Colors.textSecondary)
            .multilineTextAlignment(.center)

            if let errorMessage {
              Text(errorMessage)
                .font(DSTheme.Typography.caption)
                .foregroundColor(DSTheme.Colors.error)
            }

            HStack(spacing: DSTheme.Spacing.md) {
              DSButton("Decline", style: .ghost, fullWidth: true) {
                eventReporterService.reporter.track(
                  event: "invitation_decline_tapped", properties: [:])
                dismiss()
              }
              .disabled(isLoading)

              DSButton("Accept", icon: "checkmark", fullWidth: true, isLoading: isLoading) {
                eventReporterService.reporter.track(
                  event: "invitation_accept_tapped", properties: [:])
                Task { await acceptInvitation() }
              }
            }
          }
          .padding(DSTheme.Spacing.xl)
        }
      }
      .navigationTitle("Household Invitation")
      .navigationBarTitleDisplayMode(.inline)
      .toolbar {
        if didAccept {
          ToolbarItem(placement: .confirmationAction) {
            DSButton("Done", style: .ghost, size: .small) {
              onAccepted()
              dismiss()
            }
          }
        } else {
          ToolbarItem(placement: .cancellationAction) {
            DSButton("Cancel", style: .ghost, size: .small) {
              dismiss()
            }
          }
        }
      }
    }
  }

  private func acceptInvitation() async {
    isLoading = true
    errorMessage = nil

    do {
      // The token is on the request rather than in an input, and it is what admits the
      // caller: whoever holds the link may answer it, and the comparison happens on the
      // row the id names rather than on an index of tokens.
      var request = Primandproper_Platform_Identity_V1_AcceptInvitationRequest()
      request.invitationID = invitationID
      request.token = invitationToken
      request.statusNote = "Accepted via invite link"

      _ = try await authManager.authenticatedCall("acceptInvitation") {
        client, metadata, options in
        try await client.identity.acceptInvitation(
          request, metadata: metadata, options: options)
      }

      eventReporterService.reporter.track(event: "invitation_accepted", properties: [:])
      didAccept = true
      onAccepted()
    } catch {
      errorMessage = "Failed to accept invitation: \(error.localizedDescription)"
    }

    isLoading = false
  }
}
