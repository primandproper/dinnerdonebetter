//
//  HomeView.swift
//  ios
//
//  Created by Jeffrey Dorrycott on 12/8/25.
//

import Combine
import SwiftUI

struct HomeView: View {
  @Environment(AuthenticationManager.self) private var authManager
  @Environment(EventReporterService.self) private var eventReporterService
  @Environment(UserSettingsService.self) private var userSettingsService
  @State private var viewModel: HomeViewModel?
  @State private var showDrawer = false

  var body: some View {
    NavigationStack {
      Group {
        if let viewModel = viewModel {
          DSContentState(
            isLoading: viewModel.isLoading,
            loadingMessage: "Loading...",
            error: viewModel.errorMessage,
            errorTitle: viewModel.errorTitle,
            errorIcon: viewModel.errorIcon,
            errorIconColor: viewModel.errorIconColor,
            onRetry: { await viewModel.loadData() },
            showEnvironmentSelector: viewModel.isServerDownError,
            content: {
              VStack(spacing: 0) {
                // Header: welcome text (flex-grow) + placeholder for overlay hamburger
                HStack(alignment: .center, spacing: DSTheme.Spacing.md) {
                  Text("\(greeting), \(viewModel.currentUserDisplayName)!")
                    .font(DSTheme.Typography.title1)
                    .foregroundColor(DSTheme.Colors.textPrimary)
                    .frame(maxWidth: .infinity, alignment: .leading)

                  Color.clear
                    .frame(width: 24, height: 24)
                }
                .padding(.horizontal, DSTheme.Spacing.lg)
                .padding(.vertical, 14)

                MealPlanningHomeContent(viewModel: viewModel)
              }
            }
          )
        } else {
          DSInitializingView()
        }
      }
      .navigationTitle("")
      .navigationBarTitleDisplayMode(.inline)
      .toolbar(.hidden, for: .navigationBar)
      .refreshable {
        eventReporterService.reporter.track(event: "home_pull_to_refresh", properties: [:])
        if let viewModel = viewModel {
          await viewModel.loadData()
        }
      }
      .onAppear {
        if viewModel == nil {
          viewModel = HomeViewModel(authManager: authManager)
        }
        if let viewModel = viewModel {
          Task {
            await viewModel.loadData()
          }
        }
      }
      .onReceive(NotificationCenter.default.publisher(for: .mealPlanCreated)) { _ in
        if let viewModel = viewModel {
          Task {
            await viewModel.loadData()
          }
        }
      }
      .onReceive(NotificationCenter.default.publisher(for: .mealPlanArchived)) { _ in
        if let viewModel = viewModel {
          Task {
            await viewModel.loadData()
          }
        }
      }
      .onReceive(NotificationCenter.default.publisher(for: .mealPlanEventsUpdated)) { _ in
        if let viewModel = viewModel {
          Task {
            await viewModel.loadData()
          }
        }
      }
      .overlay {
        HomeDrawerView(
          isPresented: $showDrawer,
          displayName: viewModel?.currentUserDisplayName ?? authManager.username,
          avatarURL: viewModel.flatMap { homeViewModel in
            guard let user = homeViewModel.currentUser,
              user.hasAvatar,
              !user.avatar.objectKey.isEmpty
            else { return nil }
            return APIConfiguration.mediaURL(
              forStoragePath: user.avatar.objectKey, bucket: "avatars")
          },
          acceptedOccupiedDates: viewModel?.acceptedOccupiedDates ?? [],
          proposedOccupiedDates: viewModel?.proposedOccupiedDates ?? []
        )
      }
      .overlay(alignment: .topTrailing) {
        Button {
          eventReporterService.reporter.track(
            event: showDrawer ? "drawer_closed" : "drawer_opened",
            properties: [:])
          showDrawer.toggle()
        } label: {
          Image(systemName: showDrawer ? "xmark" : "line.3.horizontal")
            .font(.system(size: 24, weight: .medium))
            .foregroundColor(showDrawer ? .red : DSTheme.Colors.textPrimary)
            .contentTransition(.symbolEffect(.replace))
        }
        .padding(.trailing, DSTheme.Spacing.lg)
        .padding(.top, 14)
        .animation(.spring(response: 0.35, dampingFraction: 0.85), value: showDrawer)
      }
    }
  }

  private var greeting: String {
    let hour = Calendar.current.component(.hour, from: Date())
    switch hour {
    case 0..<12:
      return "Good morning"
    case 12..<17:
      return "Good afternoon"
    default:
      return "Good evening"
    }
  }
}

#Preview {
  let authManager = AuthenticationManager()
  authManager.isAuthenticated = true
  authManager.username = "John Doe"
  authManager.userID = "user123"
  authManager.accountID = "account123"

  return HomeView()
    .environment(authManager)
}
