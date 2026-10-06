//
//  AccountSettingsView.swift
//  ios
//
//  Created by Auto on 12/8/25.
//

import RevenueCat
import RevenueCatUI
import SwiftUI

struct AccountSettingsView: View {
  @Environment(AuthenticationManager.self) private var authManager
  @Environment(EventReporterService.self) private var eventReporterService
  @State private var viewModel: AccountSettingsViewModel?
  @State private var showCustomerCenter = false
  @State private var showPaywall = false
  @State private var isProActive = false
  @State private var launchOffering: Offering?

  var body: some View {
    NavigationStack {
      accountSettingsContent
    }
    .navigationTitle("My Household")
    .refreshable {
      if let viewModel = viewModel {
        await viewModel.loadData()
      }
    }
    .onAppear {
      if viewModel == nil {
        viewModel = AccountSettingsViewModel(authManager: authManager)
      }
      if let viewModel = viewModel {
        Task {
          await viewModel.loadData()
        }
      }
      Task {
        isProActive = await EntitlementService.isProActive()
      }
      if RevenueCatConfiguration.isConfigured && launchOffering == nil {
        Task { launchOffering = await SubscriptionService.launchOffering() }
      }
    }
    .sheet(isPresented: $showCustomerCenter) {
      CustomerCenterView()
        .onAppear {
          eventReporterService.reporter.track(event: "customer_center_viewed", properties: [:])
        }
    }
    .sheet(isPresented: $showPaywall) {
      if let offering = launchOffering {
        PaywallView(offering: offering)
          .onAppear {
            eventReporterService.reporter.track(event: "paywall_viewed", properties: [:])
          }
          .onDisappear {
            eventReporterService.reporter.track(event: "paywall_dismissed", properties: [:])
          }
      } else {
        ProgressView("Loading...")
          .task { launchOffering = await SubscriptionService.launchOffering() }
      }
    }
    .onChange(of: showCustomerCenter) { _, isPresented in
      if !isPresented { Task { isProActive = await EntitlementService.isProActive() } }
    }
    .onChange(of: showPaywall) { _, isPresented in
      if !isPresented { Task { isProActive = await EntitlementService.isProActive() } }
    }
    .onAppear {
      eventReporterService.reporter.track(event: "account_settings_viewed", properties: [:])
    }
  }

  @ViewBuilder
  private var accountSettingsContent: some View {
    if let viewModel = viewModel {
      DSContentState(
        isLoading: viewModel.isLoading,
        loadingMessage: "Loading household...",
        error: viewModel.errorMessage,
        errorTitle: viewModel.errorTitle,
        errorIcon: viewModel.errorIcon,
        errorIconColor: viewModel.errorIconColor,
        onRetry: { await viewModel.loadData() },
        showEnvironmentSelector: viewModel.isServerDownError,
        content: { accountSettingsScrollContent(viewModel: viewModel) }
      )
    } else {
      DSInitializingView()
    }
  }

  private func accountSettingsScrollContent(viewModel: AccountSettingsViewModel) -> some View {
    ScrollView {
      VStack(spacing: DSTheme.Spacing.xl) {
        subscriptionSection
        navigationLinksSection(viewModel: viewModel)
      }
      .dsScreenPadding()
      .padding(.bottom, DSTheme.Spacing.lg)
    }
    .frame(maxHeight: .infinity)
  }

  // MARK: - Navigation Links Section

  private func navigationLinksSection(viewModel: AccountSettingsViewModel) -> some View {
    VStack(spacing: DSTheme.Spacing.sm) {
      DSListRowLink(
        title: "Household Members",
        subtitle: "Members and invitations",
        icon: "person.2",
        style: .card,
        destination: HouseholdMembersView(viewModel: viewModel)
      )
      .simultaneousGesture(
        TapGesture().onEnded {
          eventReporterService.reporter.track(event: "household_members_tapped", properties: [:])
        })

      if viewModel.account != nil {
        DSListRowLink(
          title: "Household Details",
          subtitle: "Edit household details",
          icon: "house",
          style: .card,
          destination: HouseholdDetailsView(viewModel: viewModel)
        )
        .simultaneousGesture(
          TapGesture().onEnded {
            eventReporterService.reporter.track(event: "household_details_tapped", properties: [:])
          })

        // DSListRowLink(
        //   title: "Kitchen Instruments",
        //   subtitle: "Tools and appliances your household owns",
        //   icon: "frying.pan",
        //   style: .card,
        //   destination: HouseholdInstrumentsView(viewModel: viewModel)
        // )
      }

      DSListRowLink(
        title: "Preferences",
        subtitle: "Configure your preferences",
        icon: "slider.horizontal.3",
        style: .card,
        destination: ServiceSettingsView()
      )

      DSListRowLink(
        title: "Profile",
        subtitle: "Photo, name, and account details",
        icon: "person.crop.circle",
        style: .card,
        destination: UserProfileView()
      )
      .simultaneousGesture(
        TapGesture().onEnded {
          eventReporterService.reporter.track(event: "profile_viewed", properties: [:])
        })

      DSListRowLink(
        title: "Password",
        subtitle: "Change your password",
        icon: "key",
        style: .card,
        destination: ChangePasswordView()
      )
      .simultaneousGesture(
        TapGesture().onEnded {
          eventReporterService.reporter.track(event: "change_password_tapped", properties: [:])
        })

      DSListRowLink(
        title: "Active Sessions",
        subtitle: "Manage your signed-in devices",
        icon: "laptopcomputer.and.iphone",
        style: .card,
        destination: SessionsView()
      )
      .simultaneousGesture(
        TapGesture().onEnded {
          eventReporterService.reporter.track(event: "sessions_tapped", properties: [:])
        })
    }
  }

  // MARK: - Subscription Section
  @ViewBuilder
  private var subscriptionSection: some View {
    if RevenueCatConfiguration.isConfigured {
      DSSection("Subscription") {
        VStack(spacing: DSTheme.Spacing.lg) {
          if isProActive {
            HStack(spacing: DSTheme.Spacing.md) {
              Image(systemName: "crown.fill")
                .foregroundColor(DSTheme.Colors.primary)
              Text(Branding.proEntitlementName)
                .font(DSTheme.Typography.label)
                .foregroundColor(DSTheme.Colors.textPrimary)
              Spacer()
              DSStatusBadge(.success, style: .minimal)
            }
            DSButton("Manage Subscription", icon: "gearshape", style: .ghost, fullWidth: true) {
              eventReporterService.reporter.track(
                event: "manage_subscription_tapped", properties: [:])
              showCustomerCenter = true
            }
          } else {
            Text("Upgrade to Pro for full access")
              .font(DSTheme.Typography.body)
              .foregroundColor(DSTheme.Colors.textSecondary)
            DSButton("Upgrade to Pro", icon: "crown", fullWidth: true) {
              eventReporterService.reporter.track(event: "upgrade_to_pro_tapped", properties: [:])
              showPaywall = true
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

  return AccountSettingsView()
    .environment(authManager)
}
