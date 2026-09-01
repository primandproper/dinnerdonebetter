//
//  GroceryListView.swift
//  ios
//
//  Created by Auto on 12/8/25.
//

import SwiftProtobuf
import SwiftUI

private enum GroceryItemSection {
  case need
  case have
  case unavailable
}

struct GroceryListView: View {
  @Environment(AuthenticationManager.self) private var authManager
  @Environment(EventReporterService.self) private var eventReporterService
  @State private var viewModel: GroceryListViewModel
  @State private var itemForEditSheet: Mealplanning_MealPlanGroceryListItem?
  @State private var quantityText: String = ""

  init(
    mealPlan: any MealPlanDisplayable, items: [Mealplanning_MealPlanGroceryListItem],
    authManager: AuthenticationManager
  ) {
    _viewModel = State(
      initialValue: GroceryListViewModel(
        mealPlan: mealPlan,
        items: items,
        authManager: authManager
      )
    )
  }

  var body: some View {
    ScrollView {
      LazyVStack(alignment: .leading, spacing: DSTheme.Spacing.xl, pinnedViews: [.sectionHeaders]) {
        headerSection
          .padding(.horizontal, DSTheme.Spacing.md)

        if viewModel.isLoading {
          DSLoadingView("Loading grocery list...")
            .padding(.horizontal, DSTheme.Spacing.md)
        } else if viewModel.items.isEmpty {
          emptyStateView
            .padding(.horizontal, DSTheme.Spacing.md)
        } else {
          if !viewModel.needsItems.isEmpty {
            Section {
              itemRows(items: viewModel.needsItems, section: GroceryItemSection.need)
            } header: {
              DSRuleFlankedHeader(title: "Need", color: .orange, strikethrough: isPast)
            }
          }

          if !viewModel.unavailableItems.isEmpty {
            Section {
              itemRows(items: viewModel.unavailableItems, section: GroceryItemSection.unavailable)
            } header: {
              DSRuleFlankedHeader(title: "Unavailable", color: .red, strikethrough: isPast)
            }
          }

          if !viewModel.haveItems.isEmpty {
            Section {
              itemRows(items: viewModel.haveItems, section: GroceryItemSection.have)
            } header: {
              DSRuleFlankedHeader(title: "Have", color: .green, strikethrough: isPast)
            }
          }
        }

        if let errorMessage = viewModel.errorMessage {
          HStack {
            Text(errorMessage)
              .font(DSTheme.Typography.caption)
              .foregroundColor(DSTheme.Colors.error)
            Spacer()
            DSButton("Retry", icon: "arrow.clockwise", style: .primary, size: .small) {
              eventReporterService.reporter.track(event: "grocery_retry", properties: [:])
              Task { await viewModel.loadItems() }
            }
          }
          .padding()
        }
      }
      .padding(.vertical, DSTheme.Spacing.md)
    }
    .navigationTitle(viewModel.mealPlan.notes.isEmpty ? "Grocery List" : viewModel.mealPlan.notes)
    .navigationBarTitleDisplayMode(.large)
    .refreshable {
      await viewModel.loadItems()
    }
    .onAppear {
      eventReporterService.reporter.track(
        event: "grocery_list_viewed",
        properties: ["meal_plan_id": viewModel.mealPlan.id])
    }
    .task {
      await viewModel.loadItems()
    }
    .sheet(item: $itemForEditSheet) { item in
      EditQuantitySheet(
        item: item,
        quantityText: $quantityText,
        onSave: {
          Task {
            await handleQuantitySubmit(for: item)
          }
        },
        onCancel: {
          itemForEditSheet = nil
          quantityText = ""
        }
      )
    }
  }

  private var isPast: Bool {
    MealPlanningHomeHelpers.isMealPlanInPast(viewModel.mealPlan)
  }

  private var headerSection: some View {
    VStack(alignment: .leading, spacing: DSTheme.Spacing.sm) {
      Text(MealPlanningHomeHelpers.formatMealPlanTimeRange(viewModel.mealPlan))
        .font(DSTheme.Typography.body)
        .foregroundColor(DSTheme.Colors.textSecondary)
        .strikethrough(isPast)

      Text("\(viewModel.items.count) item\(viewModel.items.count == 1 ? "" : "s")")
        .font(DSTheme.Typography.caption)
        .foregroundColor(DSTheme.Colors.textSecondary)
        .strikethrough(isPast)
    }
  }

  private var emptyStateView: some View {
    DSEmptyState(
      icon: "cart",
      title: "No grocery items",
      message: "Grocery list items will appear here once the meal plan is finalized."
    )
  }

  private func itemRows(
    items: [Mealplanning_MealPlanGroceryListItem],
    section: GroceryItemSection
  ) -> some View {
    VStack(alignment: .leading, spacing: DSTheme.Spacing.md) {
      ForEach(items, id: \.id) { item in
        GroceryListItemRow(
          item: item,
          section: section,
          isUpdating: viewModel.isUpdating,
          strikethrough: isPast,
          onMarkAsAcquired: {
            Task { await viewModel.markAsAcquired(item) }
          },
          onMarkAsNeeds: {
            Task { await viewModel.markAsNeeds(item) }
          },
          onEditQuantity: {
            eventReporterService.reporter.track(
              event: "grocery_item_edit_tapped",
              properties: ["item_id": item.id])
            itemForEditSheet = item
            quantityText = String(item.minQuantityNeeded)
          }
        )
      }
    }
    .padding(.horizontal, DSTheme.Spacing.md)
  }

  private func handleQuantitySubmit(for item: Mealplanning_MealPlanGroceryListItem) async {
    guard let value = Float(quantityText), value > 0 else {
      itemForEditSheet = nil
      quantityText = ""
      return
    }

    await viewModel.updateQuantityNeeded(item, min: value, max: nil)
    eventReporterService.reporter.track(
      event: "grocery_item_saved",
      properties: ["item_id": item.id])

    itemForEditSheet = nil
    quantityText = ""
  }
}

// MARK: - Edit Quantity Sheet

private struct EditQuantitySheet: View {
  let item: Mealplanning_MealPlanGroceryListItem
  @Binding var quantityText: String
  let onSave: () -> Void
  let onCancel: () -> Void

  var body: some View {
    NavigationStack {
      Form {
        Section {
          HStack {
            TextField("0", text: $quantityText)
              .keyboardType(.decimalPad)
              .onChange(of: quantityText) { _, newValue in
                let filtered = filterDecimalInput(newValue)
                if filtered != newValue { quantityText = filtered }
              }
            if !item.measurementUnit.name.isEmpty {
              Text(item.measurementUnit.name)
                .font(.subheadline)
                .foregroundColor(.secondary)
            }
          }
        } header: {
          Text("Quantity")
        }
      }
      .navigationTitle("Edit quantity")
      .navigationBarTitleDisplayMode(.inline)
      .toolbar {
        ToolbarItem(placement: .cancellationAction) {
          Button("Cancel") {
            onCancel()
          }
        }
        ToolbarItem(placement: .confirmationAction) {
          Button("Save") {
            onSave()
          }
          .disabled(Float(quantityText) == nil || (Float(quantityText) ?? 0) <= 0)
        }
      }
    }
  }

  private func filterDecimalInput(_ value: String) -> String {
    var filtered = value.filter { $0.isNumber || $0 == "." }
    let parts = filtered.split(separator: ".", omittingEmptySubsequences: false)
    if parts.count > 2 {
      filtered = parts[0] + "." + parts.dropFirst().joined()
    }
    return filtered
  }
}

// MARK: - Grocery List Item Row

private struct GroceryListItemRow: View {
  let item: Mealplanning_MealPlanGroceryListItem
  let section: GroceryItemSection
  let isUpdating: Bool
  var strikethrough: Bool = false
  let onMarkAsAcquired: () -> Void
  let onMarkAsNeeds: () -> Void
  let onEditQuantity: () -> Void

  private var displayText: String {
    formatIngredientWithQuantity(
      name: item.ingredient.name,
      minQuantityNeeded: item.minQuantityNeeded,
      maxQuantityNeeded: item.hasMaxQuantityNeeded ? item.maxQuantityNeeded : nil,
      unit: item.measurementUnit
    )
  }

  var body: some View {
    HStack(spacing: 12) {
      switch section {
      case .need:
        Button {
          onMarkAsAcquired()
        } label: {
          Image(systemName: "circle")
            .font(.title2)
            .foregroundColor(.orange)
        }
        .buttonStyle(.plain)
        .disabled(isUpdating)

        Button {
          onEditQuantity()
        } label: {
          HStack {
            Text(displayText)
              .font(.body)
              .fontWeight(.medium)
              .foregroundColor(.primary)
              .multilineTextAlignment(.leading)
              .strikethrough(strikethrough)
            Spacer(minLength: 0)
          }
          .frame(maxWidth: .infinity, maxHeight: .infinity)
          .contentShape(Rectangle())
        }
        .buttonStyle(.plain)
        .disabled(isUpdating)

      case .have:
        Button {
          onMarkAsNeeds()
        } label: {
          Image(systemName: "arrow.uturn.backward")
            .font(.title3)
            .foregroundColor(.green)
        }
        .buttonStyle(.plain)
        .disabled(isUpdating)

        Text(displayText)
          .font(.body)
          .fontWeight(.medium)
          .foregroundColor(.primary)
          .frame(maxWidth: .infinity, alignment: .leading)
          .strikethrough(strikethrough)

      case .unavailable:
        Button {
          onMarkAsNeeds()
        } label: {
          Image(systemName: "arrow.uturn.backward")
            .font(.title3)
            .foregroundColor(.red)
        }
        .buttonStyle(.plain)
        .disabled(isUpdating)

        Text(displayText)
          .font(.body)
          .fontWeight(.medium)
          .foregroundColor(.secondary)
          .frame(maxWidth: .infinity, alignment: .leading)
          .strikethrough(strikethrough)
      }
    }
    .padding()
    .background(Color(.systemGray6))
    .cornerRadius(8)
  }
}

// MARK: - Format Helpers

private func formatIngredientWithQuantity(
  name: String,
  minQuantityNeeded: Float,
  maxQuantityNeeded: Float?,
  unit: Mealplanning_ValidMeasurementUnit
) -> String {
  let unitName = MeasurementUnitFormatter.displayName(for: minQuantityNeeded, unit: unit)
  let unitSuffix = unitName.isEmpty ? "" : " \(unitName)"
  let quantityStr: String
  if let maxVal = maxQuantityNeeded, minQuantityNeeded != maxVal {
    quantityStr = "\(formatNumber(minQuantityNeeded))–\(formatNumber(maxVal))\(unitSuffix)"
  } else {
    quantityStr = "\(formatNumber(minQuantityNeeded))\(unitSuffix)"
  }
  return "\(name) (\(quantityStr))"
}

private func formatNumber(_ value: Float) -> String {
  value.truncatingRemainder(dividingBy: 1) == 0
    ? String(format: "%.0f", value) : String(format: "%g", value)
}

// MARK: - MealPlanGroceryListItem Identifiable

extension Mealplanning_MealPlanGroceryListItem: Identifiable {
  // Already has id property
}

#Preview {
  let authManager = AuthenticationManager()
  authManager.isAuthenticated = true
  authManager.username = "John Doe"
  authManager.userID = "user123"
  authManager.accountID = "account123"

  var mealPlan = Mealplanning_MealPlan()
  mealPlan.id = "mealplan123"
  mealPlan.notes = "Sample Meal Plan"

  var item = Mealplanning_MealPlanGroceryListItem()
  item.id = "item123"
  var ingredient = Mealplanning_ValidIngredient()
  ingredient.name = "Chicken Breast"
  item.ingredient = ingredient
  item.minQuantityNeeded = 2.0
  item.maxQuantityNeeded = 4.0
  item.status = .needs

  return NavigationStack {
    GroceryListView(
      mealPlan: mealPlan,
      items: [item],
      authManager: authManager
    )
  }
  .environment(authManager)
}
