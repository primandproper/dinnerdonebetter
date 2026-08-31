//
//  CreateMealView.swift
//  ios
//
//  Meal creation flow: basic info, add recipes with validation, review.
//

import SwiftProtobuf
import SwiftUI

extension Notification.Name {
  static let mealCreated = Notification.Name("mealCreated")
}

// MARK: - Wizard Step

enum CreateMealWizardStep: Int, CaseIterable {
  case addRecipes = 1
  case review = 2
}

// MARK: - Create Meal View

struct CreateMealView: View {
  @Environment(AuthenticationManager.self) private var authManager
  @Environment(\.dismiss) private var dismiss
  @State private var viewModel: CreateMealViewModel?
  @State private var searchQuery: String = ""

  var body: some View {
    Group {
      if let viewModel = viewModel {
        VStack(spacing: 0) {
          stepIndicator(currentStep: viewModel.wizardStep, totalSteps: 2)
            .padding()

          ScrollView {
            VStack(alignment: .leading, spacing: DSTheme.Spacing.xl) {
              switch viewModel.wizardStep {
              case 1:
                addRecipesSection(viewModel: viewModel)
              default:
                reviewSection(viewModel: viewModel)
              }
            }
            .dsScreenPadding()
          }

          if let error = viewModel.creationError {
            HStack {
              Image(systemName: "exclamationmark.triangle")
                .foregroundColor(DSTheme.Colors.error)
              Text(error)
                .font(.subheadline)
                .foregroundColor(DSTheme.Colors.error)
            }
            .padding(.horizontal)
          }
        }
      } else {
        DSInitializingView()
      }
    }
    .navigationTitle("Create Meal")
    .navigationBarTitleDisplayMode(.large)
    .onAppear {
      if viewModel == nil {
        viewModel = CreateMealViewModel(authManager: authManager)
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

  // MARK: - Add Recipes

  private func addRecipesSection(viewModel: CreateMealViewModel) -> some View {
    VStack(alignment: .leading, spacing: DSTheme.Spacing.lg) {
      if !viewModel.draftComponents.isEmpty {
        Text("Added to meal")
          .font(.subheadline)
          .foregroundColor(.secondary)

        ForEach(Array(viewModel.draftComponents.enumerated()), id: \.element.id) {
          index, component in
          draftComponentCard(viewModel: viewModel, component: component, index: index)
        }
      }

      Text("Add Recipes")
        .font(.headline)

      HStack {
        TextField("Search recipes...", text: $searchQuery)
          .textFieldStyle(.roundedBorder)
          .autocorrectionDisabled()
          .textInputAutocapitalization(.never)
          .onChange(of: searchQuery) { _, newValue in
            viewModel.searchRecipes(query: newValue)
          }
        if viewModel.isSearching {
          ProgressView()
            .padding(.leading, 8)
        }
      }

      if let error = viewModel.searchError {
        Text(error)
          .font(.caption)
          .foregroundColor(DSTheme.Colors.error)
      }

      if !viewModel.searchResultsNotInDraft().isEmpty {
        Text("Search results")
          .font(.subheadline)
          .foregroundColor(.secondary)

        ForEach(viewModel.searchResultsNotInDraft(), id: \.id) { recipe in
          HStack {
            VStack(alignment: .leading, spacing: 4) {
              Text(recipe.name.isEmpty ? "Unnamed Recipe" : recipe.name)
                .font(DSTheme.Typography.label)
              // A search result is a RecipeSummary and carries no steps, so
              // the description stands in for the step count that used to be
              // here. The whole recipe arrives on add.
              if !recipe.description_p.isEmpty {
                Text(recipe.description_p)
                  .font(.caption)
                  .foregroundColor(.secondary)
                  .lineLimit(2)
              }
            }
            Spacer()
            Button {
              Task {
                await viewModel.addRecipe(recipe)
              }
            } label: {
              if viewModel.isAddingRecipe {
                ProgressView()
                  .scaleEffect(0.8)
              } else {
                Image(systemName: "plus.circle.fill")
                  .foregroundColor(.blue)
              }
            }
            .disabled(viewModel.isAddingRecipe)
          }
          .padding()
          .background(Color(.systemGray6))
          .cornerRadius(8)
        }
      }

      if let error = viewModel.addRecipeError {
        Text(error)
          .font(.caption)
          .foregroundColor(DSTheme.Colors.error)
      }

      HStack(spacing: 12) {
        Button {
          viewModel.wizardStep = 1
        } label: {
          Text("Back")
            .frame(maxWidth: .infinity)
            .padding()
            .background(Color(.systemGray6))
            .cornerRadius(10)
        }

        Button {
          viewModel.ensureDefaultMealNameFromComponents()
          viewModel.wizardStep = 2
        } label: {
          Text("Review")
            .fontWeight(.semibold)
            .frame(maxWidth: .infinity)
            .padding()
            .background(
              viewModel.draftComponents.isEmpty || !viewModel.hasAtLeastOneMain
                ? Color.gray : Color.blue
            )
            .foregroundColor(.white)
            .cornerRadius(10)
        }
        .disabled(viewModel.draftComponents.isEmpty || !viewModel.hasAtLeastOneMain)
      }
    }
  }

  private func draftComponentCard(
    viewModel: CreateMealViewModel, component: CreateMealDraftComponent, index: Int
  ) -> some View {
    DSCard(style: .outlined) {
      VStack(alignment: .leading, spacing: DSTheme.Spacing.sm) {
        HStack {
          VStack(alignment: .leading, spacing: 4) {
            Text(component.recipe.name.isEmpty ? "Unnamed Recipe" : component.recipe.name)
              .font(DSTheme.Typography.label)
            HStack(spacing: DSTheme.Spacing.md) {
              if !component.recipe.steps.isEmpty {
                Text("\(component.recipe.steps.count) steps")
                  .font(.caption)
                  .foregroundColor(.secondary)
              }
              if let estimate = RecipeTimeEstimation.estimate(steps: component.recipe.steps),
                estimate.minSeconds > 0 || estimate.maxSeconds > 0
              {
                Text(
                  RecipeTimeEstimation.format(
                    minSeconds: estimate.minSeconds,
                    maxSeconds: estimate.maxSeconds)
                )
                .font(.caption)
                .foregroundColor(.secondary)
              }
            }
          }
          Spacer()
          Button {
            viewModel.removeRecipe(atOffsets: IndexSet(integer: index))
          } label: {
            Image(systemName: "minus.circle.fill")
              .foregroundColor(.red)
          }
        }

        HStack(alignment: .top, spacing: DSTheme.Spacing.md) {
          Picker(
            "Type",
            selection: Binding(
              get: { component.componentType },
              set: { viewModel.setComponentType($0, for: component.id) }
            )
          ) {
            Text("Main").tag(Mealplanning_MealComponentType.main)
            Text("Side").tag(Mealplanning_MealComponentType.side)
            Text("Appetizer").tag(Mealplanning_MealComponentType.appetizer)
            Text("Dessert").tag(Mealplanning_MealComponentType.dessert)
            Text("Beverage").tag(Mealplanning_MealComponentType.beverage)
          }
          .pickerStyle(.menu)

          RecipeScaleInput(
            scale: Binding(
              get: { component.recipeScale },
              set: { viewModel.setRecipeScale($0, for: component.id) }
            ),
            estimatedPortionsMin: component.recipe.minEstimatedPortions,
            estimatedPortionsMax: component.recipe.hasMaxEstimatedPortions
              ? component.recipe.maxEstimatedPortions : nil
          )
          .frame(maxWidth: .infinity)
        }
      }
    }
  }

  // MARK: - Review

  private func reviewSection(viewModel: CreateMealViewModel) -> some View {
    VStack(alignment: .leading, spacing: DSTheme.Spacing.lg) {
      Text("Review")
        .font(.headline)

      VStack(alignment: .leading, spacing: DSTheme.Spacing.sm) {
        Text("Name")
          .font(.subheadline)
          .foregroundColor(.secondary)
        TextField(
          "e.g. Sunday Roast Dinner",
          text: Binding(
            get: { viewModel.mealName },
            set: { viewModel.mealName = $0 }
          )
        )
        .textFieldStyle(.roundedBorder)
        .autocorrectionDisabled()
      }

      VStack(alignment: .leading, spacing: DSTheme.Spacing.sm) {
        Text("Description (optional)")
          .font(.subheadline)
          .foregroundColor(.secondary)
        TextField(
          "A brief description of the meal",
          text: Binding(
            get: { viewModel.mealDescription },
            set: { viewModel.mealDescription = $0 }
          )
        )
        .textFieldStyle(.roundedBorder)
        .lineLimit(3...6)
      }

      let totalSteps = viewModel.draftComponents.reduce(0) { $0 + $1.recipe.steps.count }
      let timeEstimates = viewModel.draftComponents.compactMap {
        RecipeTimeEstimation.estimate(steps: $0.recipe.steps)
      }
      let totalTimeMin = timeEstimates.map(\.minSeconds).reduce(0, +)
      let totalTimeMax = timeEstimates.map(\.maxSeconds).reduce(0, +)

      VStack(alignment: .leading, spacing: DSTheme.Spacing.sm) {
        Text("Meal summary")
          .font(.subheadline)
          .foregroundColor(.secondary)
        HStack(spacing: DSTheme.Spacing.md) {
          Label("\(totalSteps) steps", systemImage: "list.number")
            .font(.caption)
            .foregroundColor(.secondary)
          Label(
            RecipeTimeEstimation.format(minSeconds: totalTimeMin, maxSeconds: totalTimeMax),
            systemImage: "clock"
          )
          .font(.caption)
          .foregroundColor(.secondary)
        }
      }

      ForEach(viewModel.draftComponents) { component in
        let estimate = RecipeTimeEstimation.estimate(steps: component.recipe.steps)
        HStack {
          Text(component.recipe.name.isEmpty ? "Unnamed Recipe" : component.recipe.name)
            .font(.subheadline)
          Spacer()
          Text(String(format: "%.2g×", component.recipeScale))
            .font(.caption)
            .foregroundColor(.secondary)
          Text("\(component.recipe.steps.count) steps")
            .font(.caption)
            .foregroundColor(.secondary)
          if let estimate = estimate {
            Text(
              RecipeTimeEstimation.format(
                minSeconds: estimate.minSeconds,
                maxSeconds: estimate.maxSeconds)
            )
            .font(.caption)
            .foregroundColor(.secondary)
          }
        }
        .padding(.vertical, 4)
      }

      let groceryItems = MealPreviewHelper.groceryItems(from: viewModel.draftComponents)
      if !groceryItems.isEmpty {
        VStack(alignment: .leading, spacing: DSTheme.Spacing.sm) {
          Text("Grocery list preview")
            .font(.subheadline)
            .foregroundColor(.secondary)
          VStack(alignment: .leading, spacing: 4) {
            ForEach(groceryItems) { item in
              HStack(alignment: .top, spacing: 8) {
                Image(systemName: "circle")
                  .font(.caption2)
                  .foregroundColor(.secondary)
                Text(item.displayText)
                  .font(.subheadline)
                  .foregroundColor(.primary)
              }
            }
          }
          .padding()
          .frame(maxWidth: .infinity, alignment: .leading)
          .background(Color(.systemGray6))
          .cornerRadius(8)
        }
      }

      let prepTasks = MealPreviewHelper.prepTasks(from: viewModel.draftComponents)
      if !prepTasks.isEmpty {
        VStack(alignment: .leading, spacing: DSTheme.Spacing.sm) {
          Text("Prep tasks")
            .font(.subheadline)
            .foregroundColor(.secondary)
          VStack(alignment: .leading, spacing: 4) {
            ForEach(prepTasks) { task in
              HStack(alignment: .top, spacing: 8) {
                Image(systemName: "checklist")
                  .font(.caption2)
                  .foregroundColor(.secondary)
                VStack(alignment: .leading, spacing: 2) {
                  Text(task.name)
                    .font(.subheadline)
                    .foregroundColor(.primary)
                  HStack(spacing: 4) {
                    Text("from \(task.recipeName)")
                      .font(.caption)
                      .foregroundColor(.secondary)
                    if let advance = task.advanceText {
                      Text("•")
                        .font(.caption)
                        .foregroundColor(.secondary)
                      Text(advance)
                        .font(.caption)
                        .foregroundColor(.secondary)
                    }
                  }
                }
                Spacer()
              }
            }
          }
          .padding()
          .frame(maxWidth: .infinity, alignment: .leading)
          .background(Color(.systemGray6))
          .cornerRadius(8)
        }
      }

      HStack(spacing: 12) {
        Button {
          viewModel.wizardStep = 1
        } label: {
          Text("Back")
            .frame(maxWidth: .infinity)
            .padding()
            .background(Color(.systemGray6))
            .cornerRadius(10)
        }

        Button {
          Task {
            let success = await viewModel.createMeal()
            if success {
              NotificationCenter.default.post(name: .mealCreated, object: nil)
              dismiss()
            }
          }
        } label: {
          HStack {
            if viewModel.isCreating {
              ProgressView()
                .progressViewStyle(CircularProgressViewStyle(tint: .white))
            }
            Text(viewModel.isCreating ? "Creating..." : "Create Meal")
              .fontWeight(.semibold)
          }
          .frame(maxWidth: .infinity)
          .padding()
          .background(
            viewModel.isCreating || !viewModel.canCreateMeal
              ? Color.gray : Color.blue
          )
          .foregroundColor(.white)
          .cornerRadius(10)
        }
        .disabled(viewModel.isCreating || !viewModel.canCreateMeal)
      }
    }
  }
}
