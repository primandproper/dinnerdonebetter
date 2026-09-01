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

  var body: some View {
    LoadedMealPlanView(mealPlanID: mealPlanID) { plan in
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
    .environment(authManager)
    .environment(eventReporterService)
    .environment(userSettingsService)
  }
}
