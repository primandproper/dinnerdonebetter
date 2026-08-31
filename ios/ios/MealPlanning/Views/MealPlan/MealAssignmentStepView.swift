//
//  MealAssignmentStepView.swift
//  ios
//
//  Created by Auto on 12/8/25.
//

import SwiftProtobuf
import SwiftUI

struct MealAssignmentStepView: View {
  @Bindable var viewModel: CreateMealPlanViewModel
  @Environment(EventReporterService.self) private var eventReporterService
  @FocusState private var isSearchFocused: Bool
  @State private var displayedSearchError: String?

  private let dateFormatter: DateFormatter = {
    let formatter = DateFormatter()
    formatter.dateFormat = "EEEE, MMM d"
    return formatter
  }()

  private static let assignedDayFormatter: DateFormatter = {
    let formatter = DateFormatter()
    formatter.dateFormat = "EEE"
    return formatter
  }()

  var body: some View {
    VStack(alignment: .leading, spacing: 24) {
      if let date = viewModel.currentPlanningDate {
        Text("Planning for \(dateFormatter.string(from: date))")
          .font(.title2)
          .fontWeight(.semibold)

        if viewModel.selectedDates.count > 1 {
          Text("Day \(viewModel.currentDayIndex + 1) of \(viewModel.selectedDates.count)")
            .font(.subheadline)
            .foregroundColor(.secondary)
        }

        VStack(alignment: .leading, spacing: 16) {
          Text("Search for a meal")
            .font(.headline)
          HStack {
            TextField(
              "Search meals...",
              text: $viewModel.searchQuery
            )
            .textFieldStyle(.roundedBorder)
            .autocorrectionDisabled()
            .textInputAutocapitalization(.never)
            .focused($isSearchFocused)

            if viewModel.isSearching {
              ProgressView()
                .padding(.leading, 8)
            }
          }
          .task(id: viewModel.searchQuery) {
            guard !viewModel.searchQuery.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
            else {
              viewModel.searchResults = []
              viewModel.searchError = nil
              displayedSearchError = nil
              return
            }
            displayedSearchError = nil
            try? await Task.sleep(nanoseconds: 300_000_000)
            await viewModel.searchForMeals()
          }
          .onChange(of: viewModel.searchError) { _, newValue in
            if newValue == nil {
              displayedSearchError = nil
            } else {
              Task {
                try? await Task.sleep(nanoseconds: 500_000_000)
                if viewModel.searchError == newValue, !viewModel.isSearching {
                  displayedSearchError = newValue
                }
              }
            }
          }

          if let error = displayedSearchError {
            Text(error)
              .font(.caption)
              .foregroundColor(.red)
          }
        }

        if let meal = viewModel.mealForDate(date) {
          selectedMealCard(meal: meal, date: date)
          if viewModel.mealHasIngredientOptions(meal) {
            HStack(spacing: 6) {
              Image(systemName: "list.bullet.rectangle")
                .font(.caption)
              Text("Has ingredient options to choose")
                .font(.caption)
            }
            .foregroundColor(.secondary)
          }
        }

        if !viewModel.filteredSearchResults(for: date).isEmpty {
          Text("Search results")
            .font(.headline)

          ForEach(viewModel.filteredSearchResults(for: date), id: \.id) { meal in
            let assignedDay = viewModel.assignedDayForMeal(mealID: meal.id, excludingDate: date)
            let assignedLabel = assignedDay.map { Self.assignedDayFormatter.string(from: $0) }
            MealSearchResultCard(
              meal: meal,
              isSelected: false,
              assignedToOtherDayLabel: assignedLabel.map { "Added \($0)" },
              onTap: {
                isSearchFocused = false
                eventReporterService.reporter.track(
                  event: "meal_plan_meal_assigned",
                  properties: ["meal_id": meal.id])
                Task { await viewModel.assignMeal(fromSearchResult: meal, to: date) }
              }
            )
          }
        } else if !viewModel.searchQuery.isEmpty && !viewModel.isSearching
          && viewModel.searchResults.isEmpty
        {
          HStack {
            Image(systemName: "magnifyingglass")
              .foregroundColor(.secondary)
            Text("No meals found")
              .font(.subheadline)
              .foregroundColor(.secondary)
          }
          .padding()
          .frame(maxWidth: .infinity, alignment: .leading)
          .background(Color(.systemGray6))
          .cornerRadius(8)
        }
      }
    }
    .frame(maxWidth: .infinity, alignment: .leading)
    .onAppear {
      viewModel.resetMealAssignmentState()
    }
  }

  @ViewBuilder
  private func selectedMealCard(meal: Mealplanning_Meal, date: Date) -> some View {
    VStack(alignment: .leading, spacing: 12) {
      Text("Selected")
        .font(.headline)

      VStack(alignment: .leading, spacing: 8) {
        VStack(alignment: .leading, spacing: 4) {
          Text(meal.name.isEmpty ? "Unnamed Meal" : meal.name)
            .font(DSTheme.Typography.label)
            .foregroundColor(DSTheme.Colors.textPrimary)
          if !meal.components.isEmpty {
            let names = meal.components.compactMap { $0.recipe.name.isEmpty ? nil : $0.recipe.name }
            if !names.isEmpty {
              Text(names.joined(separator: ", "))
                .font(DSTheme.Typography.caption)
                .foregroundColor(DSTheme.Colors.textSecondary)
            }
          }
        }

        // Recipe scale
        HStack(spacing: DSTheme.Spacing.md) {
          Text("Scale:")
            .font(DSTheme.Typography.label)
            .foregroundColor(DSTheme.Colors.textSecondary)

          DSStepperButtons(
            onDecrement: { viewModel.adjustMealScale(for: date, by: -0.25) },
            onIncrement: { viewModel.adjustMealScale(for: date, by: 0.25) },
            center: {
              Text(viewModel.mealScaleText(for: date))
                .font(DSTheme.Typography.body)
                .frame(minWidth: 36, alignment: .center)
            }
          )

          let scale = viewModel.mealScale(for: date)
          let portionText = PortionsFormatter.formatScaled(
            min: meal.minEstimatedPortions,
            max: meal.hasMaxEstimatedPortions ? meal.maxEstimatedPortions : nil, scale: scale)
          Text("→ ~\(portionText) servings")
            .font(DSTheme.Typography.body)
            .foregroundColor(DSTheme.Colors.textSecondary)
        }
      }
      .padding()
      .frame(maxWidth: .infinity, alignment: .leading)
      .background(Color(.systemGray6))
      .cornerRadius(10)
    }
  }

}

// MARK: - Meal Assignment Navigation Buttons

struct MealAssignmentNavigationButtons: View {
  @Bindable var viewModel: CreateMealPlanViewModel
  let onDismiss: () -> Void
  @Environment(EventReporterService.self) private var eventReporterService

  var body: some View {
    HStack(spacing: 12) {
      Button {
        if viewModel.canGoToPreviousDay {
          viewModel.goToPreviousDay()
        } else {
          viewModel.wizardStep = .weekSelection
        }
      } label: {
        HStack(spacing: 6) {
          Image(systemName: "chevron.left")
          Text("Back")
        }
        .font(.subheadline.weight(.semibold))
        .frame(maxWidth: .infinity)
        .padding()
        .background(Color(.systemGray6))
        .foregroundColor(.blue)
        .cornerRadius(10)
      }

      if viewModel.canGoToNextDay {
        Button {
          viewModel.goToNextDay()
        } label: {
          HStack(spacing: 6) {
            Text("Next Day")
            Image(systemName: "chevron.right")
          }
          .font(.subheadline.weight(.semibold))
          .frame(maxWidth: .infinity)
          .padding()
          .background(Color.blue)
          .foregroundColor(.white)
          .cornerRadius(10)
        }
      } else {
        lastDayButton
      }
    }
  }

  @ViewBuilder
  private var lastDayButton: some View {
    let recipesWithOptions = viewModel.collectRecipesWithOptions(
      from: viewModel.allSelectedMeals)
    let hasOptions = !recipesWithOptions.isEmpty

    if hasOptions {
      Button {
        viewModel.wizardStep = .optionSelection
      } label: {
        HStack(spacing: 6) {
          Text("Choose Ingredient Options")
          Image(systemName: "chevron.right")
        }
        .font(.subheadline.weight(.semibold))
        .frame(maxWidth: .infinity)
        .padding()
        .background(Color.blue)
        .foregroundColor(.white)
        .cornerRadius(10)
      }
    } else {
      createButton
    }
  }

  private var createButton: some View {
    Button(
      action: {
        Task {
          let success = await viewModel.createMealPlan()
          if success {
            eventReporterService.reporter.track(
              event: "meal_plan_created",
              properties: [
                "events_count": viewModel.selectedDates.count,
                "meals_count": viewModel.selectedDates.count,
              ])
            NotificationCenter.default.post(name: .mealPlanCreated, object: nil)
            onDismiss()
          }
        }
      },
      label: {
        HStack {
          if viewModel.isCreating {
            ProgressView()
              .progressViewStyle(CircularProgressViewStyle(tint: .white))
          }
          Text(viewModel.isCreating ? "Creating..." : "Create Meal Plan")
            .fontWeight(.semibold)
        }
        .frame(maxWidth: .infinity)
        .padding()
        .background(
          viewModel.isCreating || !viewModel.canCreateMealPlan()
            ? Color.gray : Color.blue
        )
        .foregroundColor(.white)
        .cornerRadius(10)
      }
    )
    .disabled(viewModel.isCreating || !viewModel.canCreateMealPlan())
  }
}
