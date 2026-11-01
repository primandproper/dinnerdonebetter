//
//  CreateMealPlanWizardView.swift
//  ios
//
//  Created by Auto on 12/8/25.
//

import DesignSystem
import SwiftProtobuf
import SwiftUI

extension Notification.Name {
  static let mealPlanCreated = Notification.Name("mealPlanCreated")
  static let mealPlanArchived = Notification.Name("mealPlanArchived")
  static let mealPlanEventsUpdated = Notification.Name("mealPlanEventsUpdated")
}

struct CreateMealPlanWizardView: View {
  @Environment(AuthenticationManager.self) private var authManager
  @Environment(EventReporterService.self) private var eventReporterService
  @Environment(\.dismiss) var dismiss
  @State private var viewModel: CreateMealPlanViewModel?

  var acceptedOccupiedDates: Set<Date> = []
  var proposedOccupiedDates: Set<Date> = []

  private func totalSteps(for viewModel: CreateMealPlanViewModel) -> Int {
    let hasOptions = !viewModel.collectRecipesWithOptions(
      from: viewModel.allSelectedMeals
    ).isEmpty
    return hasOptions || viewModel.wizardStep == .optionSelection ? 3 : 2
  }

  var body: some View {
    Group {
      if let viewModel = viewModel {
        VStack(spacing: 0) {
          stepIndicator(
            currentStep: viewModel.wizardStep.rawValue,
            totalSteps: totalSteps(for: viewModel)
          )
          .padding()

          if viewModel.wizardStep == .mealAssignment,
            let date = viewModel.currentPlanningDate,
            viewModel.mealForDate(date) != nil
          {
            MealAssignmentNavigationButtons(
              viewModel: viewModel,
              onDismiss: { dismiss() }
            )
            .padding(.horizontal)
            .padding(.bottom, 8)
          }

          ScrollView {
            VStack(spacing: 24) {
              switch viewModel.wizardStep {
              case .weekSelection:
                WeekSelectionStepView(viewModel: viewModel)

              case .mealAssignment:
                MealAssignmentStepView(viewModel: viewModel)

              case .optionSelection:
                OptionSelectionStepView(
                  viewModel: viewModel,
                  onDismiss: { dismiss() }
                )
              }
            }
            .padding()
          }
          .scrollDismissesKeyboard(.interactively)

          if let error = viewModel.creationError {
            HStack {
              Image(systemName: "exclamationmark.triangle")
                .foregroundColor(.red)
              Text(error)
                .font(.subheadline)
                .foregroundColor(.red)
            }
            .padding(.horizontal)
          }
        }
      } else {
        DSInitializingView()
      }
    }
    .navigationTitle("Plan Dinners")
    .navigationBarTitleDisplayMode(.large)
    .onAppear {
      if viewModel == nil {
        viewModel = CreateMealPlanViewModel(
          authManager: authManager,
          acceptedOccupiedDates: acceptedOccupiedDates,
          proposedOccupiedDates: proposedOccupiedDates
        )
        eventReporterService.reporter.track(event: "meal_plan_wizard_started", properties: [:])
      }
    }
  }

  private func stepIndicator(currentStep: Int, totalSteps: Int) -> some View {
    HStack(spacing: 8) {
      ForEach(1...totalSteps, id: \.self) { step in
        Capsule()
          .fill(step <= currentStep ? Color.blue : Color(.systemGray5))
          .frame(height: 4)
          .frame(maxWidth: .infinity)
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
    CreateMealPlanWizardView()
      .environment(authManager)
      .environment(EventReporterService())
  }
}
