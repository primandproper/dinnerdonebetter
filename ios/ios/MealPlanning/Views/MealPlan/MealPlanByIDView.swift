//
//  MealPlanByIDView.swift
//  ios
//
//  Loads a meal plan by ID (e.g. from a Universal Link) and presents Vote or Detail view.
//

import SwiftProtobuf
import SwiftUI

struct MealPlanByIDView: View {
  @Environment(AuthenticationManager.self) private var authManager
  @Environment(EventReporterService.self) private var eventReporterService
  @Environment(UserSettingsService.self) private var userSettingsService
  @Environment(\.dismiss) private var dismiss

  let mealPlanID: String

  @State private var mealPlan: Mealplanning_MealPlan?
  @State private var loadError: String?
  @State private var isLoading = true

  var body: some View {
    Group {
      if isLoading {
        DSInitializingView()
      } else if let error = loadError {
        VStack(spacing: DSTheme.Spacing.lg) {
          Text(error)
            .multilineTextAlignment(.center)
            .padding()
          Button("Close") {
            dismiss()
          }
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
      } else if let plan = mealPlan {
        if plan.status == .awaitingVotes {
          VoteMealPlanView(mealPlan: plan)
        } else {
          NavigationStack {
            MealPlanDetailView(mealPlan: plan, groceryListItems: nil)
              .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                  Button("Close") {
                    dismiss()
                  }
                }
              }
          }
        }
      }
    }
    .task {
      await fetchMealPlan()
    }
    .environment(authManager)
    .environment(eventReporterService)
    .environment(userSettingsService)
  }

  private func fetchMealPlan() async {
    isLoading = true
    loadError = nil

    do {
      var request = Mealplanning_GetMealPlanRequest()
      request.mealPlanID = mealPlanID

      let response = try await authManager.authenticatedCall("getMealPlan", idempotent: true) {
        client, metadata, options in
        try await client.mealPlanning.getMealPlan(request, metadata: metadata, options: options)
      }

      mealPlan = response.result
    } catch {
      loadError = error.localizedDescription
    }
    isLoading = false
  }
}
