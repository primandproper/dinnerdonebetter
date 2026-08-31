//
//  RecipeListView.swift
//  ios
//
//  Created by Auto on 12/8/25.
//

import SwiftProtobuf
import SwiftUI

struct RecipeListView: View {
  @Environment(AuthenticationManager.self) private var authManager
  @Environment(EventReporterService.self) private var eventReporterService
  @State private var viewModel: RecipeListViewModel?
  @State private var searchQuery: String = ""

  var body: some View {
    NavigationStack {
      Group {
        if let viewModel = viewModel {
          DSContentState(
            isLoading: viewModel.isLoading,
            loadingMessage: "Loading recipes...",
            error: viewModel.errorMessage,
            onRetry: { await viewModel.loadRecipes() },
            content: {
              let displayedRecipes = viewModel.displayedRecipes
              let isSearching = viewModel.isSearching
              let hasSearchQuery = !searchQuery.trimmingCharacters(in: .whitespacesAndNewlines)
                .isEmpty

              if isSearching {
                DSLoadingView("Searching recipes...")
              } else if displayedRecipes.isEmpty {
                if hasSearchQuery {
                  VStack(spacing: DSTheme.Spacing.lg) {
                    DSEmptyState(
                      icon: "magnifyingglass",
                      title: "No Results",
                      message: "No recipes found matching \"\(searchQuery)\""
                    )

                    if let searchError = viewModel.searchError {
                      Text(searchError)
                        .font(DSTheme.Typography.caption)
                        .foregroundColor(DSTheme.Colors.error)
                        .padding(.horizontal)
                    }
                  }
                } else {
                  DSEmptyState(
                    icon: "book.closed",
                    title: "No Recipes",
                    message: "No recipes found. Create some recipes to get started.",
                    size: .large
                  )
                }
              } else {
                ScrollView {
                  LazyVStack(spacing: DSTheme.Spacing.md) {
                    ForEach(displayedRecipes, id: \.id) { recipe in
                      NavigationLink(destination: PerformRecipeView(recipeID: recipe.id)) {
                        RecipeCard(recipe: recipe)
                      }
                      .simultaneousGesture(
                        TapGesture().onEnded {
                          eventReporterService.reporter.track(
                            event: "recipe_card_tapped",
                            properties: ["recipe_id": recipe.id])
                        })
                    }
                  }
                  .dsScreenPadding()
                }
              }
            })
        } else {
          DSInitializingView()
        }
      }
      .navigationTitle("Recipes")
      .navigationBarTitleDisplayMode(.large)
      .toolbar {
        if let viewModel = viewModel, viewModel.isServiceAdmin {
          ToolbarItem(placement: .primaryAction) {
            HStack(spacing: DSTheme.Spacing.sm) {
              Text("Submitted")
                .font(DSTheme.Typography.caption)
                .foregroundColor(
                  viewModel.recipeStatusFilter == "submitted"
                    ? DSTheme.Colors.textPrimary
                    : DSTheme.Colors.textSecondary
                )
              Toggle(
                "",
                isOn: Binding(
                  get: { viewModel.recipeStatusFilter == "approved" },
                  set: {
                    let newFilter = $0 ? "approved" : "submitted"
                    eventReporterService.reporter.track(
                      event: "recipe_status_filter_changed",
                      properties: ["filter": newFilter])
                    viewModel.setRecipeStatusFilter(newFilter)
                    Task { await viewModel.loadRecipes() }
                  }
                )
              )
              .labelsHidden()
              Text("Approved")
                .font(DSTheme.Typography.caption)
                .foregroundColor(
                  viewModel.recipeStatusFilter == "approved"
                    ? DSTheme.Colors.textPrimary
                    : DSTheme.Colors.textSecondary
                )
            }
          }
        }
      }
      .searchable(text: $searchQuery, prompt: "Search recipes...")
      .onChange(of: searchQuery) { _, newValue in
        if let viewModel = viewModel {
          viewModel.searchRecipes(query: newValue)
        }
      }
      .refreshable {
        if let viewModel = viewModel {
          searchQuery = ""
          viewModel.searchResults = []
          await viewModel.loadRecipes()
        }
      }
      .onAppear {
        if viewModel == nil {
          viewModel = RecipeListViewModel(authManager: authManager)
          eventReporterService.reporter.track(event: "recipes_list_viewed", properties: [:])
        }
        if let viewModel = viewModel {
          Task {
            await viewModel.loadCurrentUserForAdminCheck()
            await viewModel.loadRecipes()
          }
        }
      }
    }
  }
}

// MARK: - Recipe Card

struct RecipeCard: View {
  let recipe: Mealplanning_RecipeSummary

  var body: some View {
    DSCard(style: .outlined) {
      VStack(alignment: .leading, spacing: DSTheme.Spacing.sm) {
        Text(recipe.name.isEmpty ? "Unnamed Recipe" : recipe.name)
          .font(DSTheme.Typography.label)
          .foregroundColor(DSTheme.Colors.textPrimary)
          .frame(maxWidth: .infinity, alignment: .leading)

        if !recipe.description_p.isEmpty {
          Text(recipe.description_p)
            .font(DSTheme.Typography.body)
            .foregroundColor(DSTheme.Colors.textSecondary)
            .lineLimit(2)
            .frame(maxWidth: .infinity, alignment: .leading)
        }

        // Recipe metadata - primary row. No step count or time estimate: a
        // RecipeSummary carries no steps. GetRecipes never returned them
        // either, so these labels have always been blank on this screen.
        HStack(spacing: DSTheme.Spacing.md) {
          Label(
            "\(PortionsFormatter.format(min: recipe.minEstimatedPortions, max: recipe.hasMaxEstimatedPortions ? recipe.maxEstimatedPortions : nil))",
            systemImage: "person.2"
          )
          .font(DSTheme.Typography.caption)
          .foregroundColor(DSTheme.Colors.textSecondary)
        }

        // Recipe metadata - secondary row (source, yields type, prep tasks, eligible)
        HStack(spacing: DSTheme.Spacing.md) {
          if !recipe.source.isEmpty {
            RecipeSourceView(source: recipe.source)
              .lineLimit(1)
          }

          if recipe.yieldsComponentType != .unspecified {
            let yieldsLabel = MealComponentTypeFormatter.format(recipe.yieldsComponentType)
            if !yieldsLabel.isEmpty {
              Text(yieldsLabel)
                .font(DSTheme.Typography.caption)
                .foregroundColor(DSTheme.Colors.textSecondary)
                .padding(.horizontal, DSTheme.Spacing.xs)
                .padding(.vertical, 2)
                .background(DSTheme.Colors.borderSubtle.opacity(0.5))
                .clipShape(RoundedRectangle(cornerRadius: 4))
            }
          }

        }
      }
      .frame(maxWidth: .infinity, alignment: .leading)
    }
  }

}

// MARK: - Preview

#Preview {
  let authManager = AuthenticationManager()
  authManager.isAuthenticated = true
  authManager.username = "Test User"
  authManager.userID = "user123"
  authManager.accountID = "account123"

  return RecipeListView()
    .environment(authManager)
}
