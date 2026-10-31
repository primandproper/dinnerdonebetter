//
//  HomeDrawerView.swift
//  ios
//
//  Side drawer for home screen navigation (Browse, Create Meal Plan, Account).
//

import DesignSystem
import SwiftUI

struct HomeDrawerView: View {
  @Environment(AuthenticationManager.self) private var authManager
  @Environment(EventReporterService.self) private var eventReporterService
  @Binding var isPresented: Bool

  var displayName: String = ""
  var avatarURL: URL?
  var acceptedOccupiedDates: Set<Date> = []
  var proposedOccupiedDates: Set<Date> = []

  private let drawerWidth: CGFloat = 280

  var body: some View {
    ZStack(alignment: .trailing) {
      // Backdrop - tap to dismiss (fades in/out)
      Color.black.opacity(isPresented ? 0.35 : 0)
        .ignoresSafeArea()
        .onTapGesture {
          if isPresented {
            eventReporterService.reporter.track(event: "drawer_closed", properties: [:])
          }
          isPresented = false
        }

      // Drawer panel (slides in from right)
      VStack(alignment: .leading, spacing: 0) {
        // Account section
        VStack(alignment: .leading, spacing: DSTheme.Spacing.sm) {
          DSSectionHeader(title: "Account", style: .label)
            .padding(.top, DSTheme.Spacing.lg)
            .padding(.horizontal, DSTheme.Spacing.md)

          NavigationLink(destination: AccountSettingsView()) {
            HStack(spacing: DSTheme.Spacing.md) {
              DSAvatar(
                name: displayName,
                size: .sm,
                imageURL: avatarURL
              )
              Text(displayName.isEmpty ? "Account" : displayName)
                .font(DSTheme.Typography.button)
                .foregroundColor(DSTheme.Colors.textPrimary)
            }
            .frame(maxWidth: .infinity, alignment: .leading)
            .contentShape(Rectangle())
            .padding(.vertical, DSTheme.Spacing.lg)
            .padding(.horizontal, DSTheme.Spacing.xl)
          }
          .buttonStyle(.plain)
          .simultaneousGesture(
            TapGesture().onEnded {
              eventReporterService.reporter.track(event: "drawer_account_tapped", properties: [:])
              isPresented = false
            }
          )
          .padding(.horizontal, DSTheme.Spacing.md)
        }

        // Create Meal Plan CTA
        VStack(alignment: .leading, spacing: DSTheme.Spacing.sm) {
          DSSectionHeader(title: "Meal Planning", style: .label)
            .padding(.top, DSTheme.Spacing.lg)
            .padding(.horizontal, DSTheme.Spacing.md)

          NavigationLink(
            destination: CreateMealPlanWizardView(
              acceptedOccupiedDates: acceptedOccupiedDates,
              proposedOccupiedDates: proposedOccupiedDates
            )
          ) {
            Label("Create Meal Plan", systemImage: "plus.circle.fill")
              .font(DSTheme.Typography.button)
              .foregroundColor(DSTheme.Colors.primary)
              .frame(maxWidth: .infinity, alignment: .leading)
              .contentShape(Rectangle())
              .padding(.vertical, DSTheme.Spacing.lg)
              .padding(.horizontal, DSTheme.Spacing.xl)
          }
          .buttonStyle(.plain)
          .simultaneousGesture(
            TapGesture().onEnded {
              eventReporterService.reporter.track(
                event: "drawer_create_meal_plan_tapped", properties: [:])
              isPresented = false
            }
          )
          .padding(.horizontal, DSTheme.Spacing.md)
        }

        // Browse section
        VStack(alignment: .leading, spacing: DSTheme.Spacing.sm) {
          DSSectionHeader(title: "Browse", style: .label)
            .padding(.top, DSTheme.Spacing.lg)
            .padding(.horizontal, DSTheme.Spacing.md)

          NavigationLink(destination: MealListView()) {
            Label("Meals", systemImage: "fork.knife")
              .font(DSTheme.Typography.button)
              .foregroundColor(DSTheme.Colors.textPrimary)
              .frame(maxWidth: .infinity, alignment: .leading)
              .contentShape(Rectangle())
              .padding(.vertical, DSTheme.Spacing.lg)
              .padding(.horizontal, DSTheme.Spacing.xl)
          }
          .buttonStyle(.plain)
          .simultaneousGesture(
            TapGesture().onEnded {
              eventReporterService.reporter.track(event: "drawer_meals_tapped", properties: [:])
              isPresented = false
            }
          )
          .padding(.horizontal, DSTheme.Spacing.md)

          NavigationLink(destination: RecipeListView()) {
            Label("Recipes", systemImage: "book.closed.fill")
              .font(DSTheme.Typography.button)
              .foregroundColor(DSTheme.Colors.textPrimary)
              .frame(maxWidth: .infinity, alignment: .leading)
              .contentShape(Rectangle())
              .padding(.vertical, DSTheme.Spacing.lg)
              .padding(.horizontal, DSTheme.Spacing.xl)
          }
          .buttonStyle(.plain)
          .simultaneousGesture(
            TapGesture().onEnded {
              eventReporterService.reporter.track(event: "drawer_recipes_tapped", properties: [:])
              isPresented = false
            }
          )
          .padding(.horizontal, DSTheme.Spacing.md)
        }

        Spacer()

        VStack(spacing: DSTheme.Spacing.lg) {
          DSDivider()
          DSButton("Sign Out", style: .destructiveGhost, size: .large, fullWidth: true) {
            eventReporterService.reporter.track(event: "sign_out_tapped", properties: [:])
            isPresented = false
            Task {
              await authManager.logout()
            }
          }
        }
        .padding(DSTheme.Spacing.md)
        .background(Color(.systemBackground))
      }
      .frame(width: drawerWidth)
      .background(Color(.systemBackground))
      .offset(x: isPresented ? 0 : drawerWidth)
    }
    .allowsHitTesting(isPresented)
    .animation(.spring(response: 0.35, dampingFraction: 0.85), value: isPresented)
  }
}
