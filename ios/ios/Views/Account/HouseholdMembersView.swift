//
//  HouseholdMembersView.swift
//  ios
//

import SwiftUI
import UIKit

struct HouseholdMembersView: View {
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
      content: { membersContent }
    )
    .navigationTitle("Household Members")
    .refreshable {
      await viewModel.loadData()
    }
  }

  @ViewBuilder
  private var membersContent: some View {
    if let account = viewModel.account {
      ScrollView {
        VStack(spacing: DSTheme.Spacing.xl) {
          membersSection(members: viewModel.members)
          if viewModel.isAccountAdmin {
            sendInvitationSection
          }
          if !viewModel.invitations.isEmpty {
            pendingInvitationsSection
          }
        }
        .dsScreenPadding()
      }
    }
  }

  // The roster is a read of its own rather than a field of the account.
  private func membersSection(members: [Primandproper_Platform_Identity_V1_MembershipWithUser])
    -> some View
  {
    DSSection("Household Members") {
      if members.isEmpty {
        DSSectionEmptyContent(
          "No members yet. Invite someone to join your household.",
          icon: "person.2"
        )
      } else {
        ForEach(members, id: \.membership.id) { member in
          MemberCard(
            member: member,
            currentUserID: viewModel.currentUserID,
            isAccountAdmin: viewModel.isAccountAdmin,
            onRoleChange: { newRole, reason in
              Task {
                await viewModel.updateMemberRole(
                  membershipID: member.membership.id, newRole: newRole, reason: reason)
              }
            }
          )
        }
      }
    }
  }

  private var sendInvitationSection: some View {
    DSSection(
      "Add Someone to Your Household",
      subtitle: "Send an invitation by email. They can join once they have an account."
    ) {
      VStack(spacing: DSTheme.Spacing.lg) {
        DSTextField(
          "Email Address",
          text: Binding(
            get: { viewModel.invitationEmail }, set: { viewModel.invitationEmail = $0 }),
          type: .email
        )

        DSTextField(
          "Name (Optional)",
          text: Binding(get: { viewModel.invitationName }, set: { viewModel.invitationName = $0 })
        )

        DSTextField(
          "Note (Optional)",
          text: Binding(get: { viewModel.invitationNote }, set: { viewModel.invitationNote = $0 }),
          type: .multiline
        )

        DSButton(
          "Send Invitation",
          icon: "envelope",
          fullWidth: true,
          isDisabled: viewModel.invitationEmail.isEmpty
        ) {
          eventReporterService.reporter.track(event: "household_invite_sent", properties: [:])
          Task {
            await viewModel.sendInvitation()
          }
        }
      }
    }
  }

  private var pendingInvitationsSection: some View {
    DSSection(
      "Invitations",
      subtitle: "Invitations you've sent for this household and their status."
    ) {
      ForEach(viewModel.invitations, id: \.id) { invitation in
        InvitationCard(
          invitation: invitation,
          isAccountAdmin: viewModel.isAccountAdmin,
          onCancel: {
            _ = await viewModel.cancelInvitation(invitationID: invitation.id)
          }
        )
      }
    }
  }
}

// MARK: - Member Card

struct MemberCard: View {
  let member: Primandproper_Platform_Identity_V1_MembershipWithUser
  let currentUserID: String
  let isAccountAdmin: Bool
  let onRoleChange: (String, String) -> Void

  @State private var selectedRole: String
  @State private var showReasonAlert = false
  @State private var pendingNewRole: String = ""
  @State private var reasonText: String = ""
  @State private var originalRole: String

  init(
    member: Primandproper_Platform_Identity_V1_MembershipWithUser,
    currentUserID: String,
    isAccountAdmin: Bool,
    onRoleChange: @escaping (String, String) -> Void
  ) {
    self.member = member
    self.currentUserID = currentUserID
    self.isAccountAdmin = isAccountAdmin
    self.onRoleChange = onRoleChange
    let initialRole = member.membership.roles.contains("account_admin") ? "Admin" : "Member"
    _selectedRole = State(initialValue: initialRole)
    _originalRole = State(initialValue: initialRole)
  }

  var body: some View {
    DSCard {
      HStack(spacing: DSTheme.Spacing.md) {
        // Avatar
        DSAvatar(
          name: displayName,
          size: .lg,
          // See HomeView: an avatar is not a field of the directory's user.
          imageURL: nil
        )

        VStack(alignment: .leading, spacing: DSTheme.Spacing.xxs) {
          Text(displayName)
            .font(DSTheme.Typography.label)
            .foregroundColor(DSTheme.Colors.textPrimary)

          if member.hasUser && member.user.id == currentUserID {
            Text("(You)")
              .font(DSTheme.Typography.caption)
              .foregroundColor(DSTheme.Colors.textSecondary)
          }
        }

        Spacer()

        if isAccountAdmin {
          Picker(
            "Role",
            selection: Binding(
              get: { selectedRole },
              set: { newValue in
                let newRole = newValue == "Admin" ? "account_admin" : "account_member"
                let currentRole = member.membership.roles.first ?? "account_member"
                if newRole != currentRole {
                  pendingNewRole = newRole
                  reasonText = ""
                  showReasonAlert = true
                } else {
                  selectedRole = newValue
                }
              }
            )
          ) {
            Text("Member").tag("Member")
            Text("Admin").tag("Admin")
          }
          .pickerStyle(.menu)
        } else {
          DSStatusBadge(
            .custom(
              member.membership.roles.contains("account_admin") ? "Admin" : "Member",
              DSTheme.Colors.textSecondary
            ),
            style: .minimal
          )
        }
      }
    }
    .onChange(of: member.membership.roles) { _, newRoles in
      let newRoleString = newRoles.contains("account_admin") ? "Admin" : "Member"
      originalRole = newRoleString
      selectedRole = newRoleString
    }
    .sheet(isPresented: $showReasonAlert) {
      NavigationStack {
        Form {
          Section {
            let roleName = pendingNewRole == "account_admin" ? "Admin" : "Member"
            Text("Please provide a reason for changing this user's role to \(roleName).")
              .font(DSTheme.Typography.body)
              .foregroundColor(DSTheme.Colors.textSecondary)
          }
          Section {
            TextField("Reason", text: $reasonText, axis: .vertical)
              .lineLimit(3...6)
          }
        }
        .navigationTitle("Change Role")
        .navigationBarTitleDisplayMode(.inline)
        .toolbar {
          ToolbarItem(placement: .cancellationAction) {
            Button("Cancel") {
              pendingNewRole = ""
              reasonText = ""
              showReasonAlert = false
            }
          }
          ToolbarItem(placement: .confirmationAction) {
            Button("Confirm") {
              if !reasonText.isEmpty {
                selectedRole = pendingNewRole == "account_admin" ? "Admin" : "Member"
                onRoleChange(pendingNewRole, reasonText)
                pendingNewRole = ""
                reasonText = ""
                showReasonAlert = false
              }
            }
            .disabled(reasonText.isEmpty)
          }
        }
      }
      .presentationDetents([.medium])
    }
    .onChange(of: showReasonAlert) { _, isPresented in
      if !isPresented && !pendingNewRole.isEmpty {
        pendingNewRole = ""
        reasonText = ""
      }
    }
  }

  // DisplayName is never empty on a read — a row whose column is blank reads its handle
  // back — so this is one field rather than a fallback chain that could produce a blank.
  private var displayName: String {
    guard member.hasUser else {
      return "Unknown User"
    }

    return member.user.displayName.isEmpty ? "Unknown User" : member.user.displayName
  }
}

// MARK: - Invitation Card

struct InvitationCard: View {
  @Environment(EventReporterService.self) private var eventReporterService
  let invitation: Primandproper_Platform_Identity_V1_Invitation
  let isAccountAdmin: Bool
  let onCancel: (() async -> Void)?

  private var isPending: Bool {
    invitation.status == .pending
  }

  var body: some View {
    DSCard {
      HStack {
        VStack(alignment: .leading, spacing: DSTheme.Spacing.xs) {
          Text(invitation.toEmail)
            .font(DSTheme.Typography.label)
            .foregroundColor(DSTheme.Colors.textPrimary)

          if !invitation.toName.isEmpty {
            Text(invitation.toName)
              .font(DSTheme.Typography.body)
              .foregroundColor(DSTheme.Colors.textSecondary)
          }

          DSStatusBadge(status: isPending ? "pending" : "answered")
        }

        Spacer()

        // No "Copy Link". An invitation's token is the secret half of a bearer credential
        // for joining somebody else's household, and no read returns one — the sender
        // mints it, the store keeps a digest, and the only copy goes to the address it was
        // sent to. A sender who could copy the link could join as the person they invited.
        if isPending && isAccountAdmin {
          DSButton("Cancel", style: .ghost, size: .small) {
            Task {
              await onCancel?()
            }
          }
        }
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
    HouseholdMembersView(viewModel: AccountSettingsViewModel(authManager: authManager))
      .environment(authManager)
      .environment(EventReporterService())
  }
}
