import SwiftUI

struct SessionsView: View {
  @Environment(AuthenticationManager.self) private var authManager
  @Environment(EventReporterService.self) private var eventReporterService
  @State private var viewModel: SessionsViewModel?
  @State private var sessionToRevoke: SignIn?
  @State private var showRevokeAllConfirmation = false

  var body: some View {
    DSContentState(
      isLoading: viewModel?.isLoading ?? true,
      loadingMessage: "Loading sessions...",
      error: viewModel?.errorMessage,
      errorTitle: "Error",
      errorIcon: "exclamationmark.triangle",
      errorIconColor: DSTheme.Colors.error,
      onRetry: { Task { await viewModel?.loadSessions() } },
      content: {
        ScrollView {
          VStack(spacing: DSTheme.Spacing.md) {
            if hasOtherSessions {
              revokeAllButton
            }

            ForEach(sessions, id: \.familyID) { session in
              sessionCard(session)
            }

            if sessions.isEmpty {
              Text("No active sessions found.")
                .foregroundColor(DSTheme.Colors.textSecondary)
                .padding(.top, DSTheme.Spacing.xl)
            }
          }
          .dsScreenPadding()
          .padding(.bottom, DSTheme.Spacing.lg)
        }
      }
    )
    .navigationTitle("Active Sessions")
    .task {
      let vm = SessionsViewModel(authManager: authManager)
      viewModel = vm
      await vm.loadSessions()
      eventReporterService.reporter.track(event: "sessions_viewed", properties: [:])
    }
    .alert(
      "Revoke Session",
      isPresented: .init(
        get: { sessionToRevoke != nil },
        set: { if !$0 { sessionToRevoke = nil } }
      )
    ) {
      Button("Cancel", role: .cancel) { sessionToRevoke = nil }
      Button("Revoke", role: .destructive) {
        if let session = sessionToRevoke {
          Task {
            _ = await viewModel?.revokeSession(familyID: session.familyID)
          }
        }
      }
    } message: {
      if let session = sessionToRevoke {
        Text(
          "Revoke the session from \(deviceName(session).isEmpty ? "Unknown device" : deviceName(session))?"
        )
      }
    }
    .alert("Revoke All Other Sessions", isPresented: $showRevokeAllConfirmation) {
      Button("Cancel", role: .cancel) {}
      Button("Revoke All", role: .destructive) {
        Task {
          _ = await viewModel?.revokeAllOtherSessions()
        }
      }
    } message: {
      Text("This will sign you out of all other devices. This cannot be undone.")
    }
  }

  // MARK: - Computed Properties

  private var sessions: [SignIn] {
    viewModel?.sessions ?? []
  }

  private var hasOtherSessions: Bool {
    sessions.contains { !$0.current }
  }

  // MARK: - Subviews

  private var revokeAllButton: some View {
    DSButton("Revoke All Other Sessions", style: .outline, size: .medium) {
      showRevokeAllConfirmation = true
    }
  }

  private func sessionCard(_ session: SignIn) -> some View {
    DSCard {
      VStack(alignment: .leading, spacing: DSTheme.Spacing.sm) {
        // Header: device name + current badge
        HStack {
          Image(systemName: iconForDevice(deviceName(session)))
            .foregroundColor(DSTheme.Colors.primary)
            .font(.title3)
          VStack(alignment: .leading, spacing: 2) {
            Text(deviceName(session).isEmpty ? "Unknown device" : deviceName(session))
              .font(.headline)
            if !session.credentialKind.isEmpty {
              Text("via \(session.credentialKind)")
                .font(.caption)
                .foregroundColor(DSTheme.Colors.textSecondary)
            }
          }
          Spacer()
          if session.current {
            Text("Current")
              .font(.caption)
              .fontWeight(.semibold)
              .foregroundColor(.white)
              .padding(.horizontal, DSTheme.Spacing.sm)
              .padding(.vertical, 4)
              .background(DSTheme.Colors.primary)
              .cornerRadius(DSTheme.Radius.sm)
          }
        }

        Divider()

        // Details
        VStack(alignment: .leading, spacing: 4) {
          if let address = session.attributes[Self.ipAddressAttribute], !address.isEmpty {
            detailRow(icon: "network", text: address)
          }
          if session.hasSignedInAt {
            detailRow(icon: "calendar", text: "Created \(formatDate(session.signedInAt.date))")
          }
          if session.hasLastRefreshedAt {
            detailRow(icon: "clock", text: "Active \(formatDate(session.lastRefreshedAt.date))")
          }
          if session.hasExpiresAt {
            detailRow(icon: "timer", text: "Expires \(formatDate(session.expiresAt.date))")
          }
        }

        // Revoke button (only for non-current sessions)
        if !session.current {
          HStack {
            Spacer()
            DSButton("Revoke", style: .ghost, size: .small) {
              sessionToRevoke = session
            }
          }
        }
      }
    }
  }

  private func detailRow(icon: String, text: String) -> some View {
    HStack(spacing: DSTheme.Spacing.sm) {
      Image(systemName: icon)
        .font(.caption)
        .foregroundColor(DSTheme.Colors.textSecondary)
        .frame(width: 16)
      Text(text)
        .font(.caption)
        .foregroundColor(DSTheme.Colors.textSecondary)
    }
  }

  /// The attributes the server's sign-in annotator answers each login with. See
  /// internal/authentication/devices in the backend.
  static let ipAddressAttribute = "ip_address"
  static let userAgentAttribute = "user_agent"
  static let deviceNameAttribute = "device_name"

  /// The device a login is on, as the server recorded it: the name a client gave its device, or
  /// the user agent it signed in with. Empty for a login nothing was recorded for.
  private func deviceName(_ session: SignIn) -> String {
    if let name = session.attributes[Self.deviceNameAttribute], !name.isEmpty {
      return name
    }
    return session.attributes[Self.userAgentAttribute] ?? ""
  }

  private func iconForDevice(_ deviceName: String) -> String {
    let lower = deviceName.lowercased()
    if lower.contains("iphone") || lower.contains("ios") {
      return "iphone"
    } else if lower.contains("ipad") {
      return "ipad"
    } else if lower.contains("mac") {
      return "laptopcomputer"
    } else if lower.contains("android") {
      return "phone"
    } else if lower.contains("windows") || lower.contains("linux") {
      return "desktopcomputer"
    } else {
      return "laptopcomputer.and.iphone"
    }
  }

  private func formatDate(_ date: Date) -> String {
    let formatter = DateFormatter()
    formatter.dateStyle = .medium
    formatter.timeStyle = .short
    return formatter.string(from: date)
  }
}
