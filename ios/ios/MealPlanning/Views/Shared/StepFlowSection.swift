//
//  StepFlowSection.swift
//  ios
//
//  Created by Auto on 2/23/26.
//

import DesignSystem
import SwiftUI
import UniformTypeIdentifiers

struct StepFlowGroup<Item: Identifiable>: Identifiable {
  let id: String
  let title: String
  let color: Color
  let items: [Item]

  init(title: String, color: Color, items: [Item]) {
    self.id = title
    self.title = title
    self.color = color
    self.items = items
  }
}

private struct StepReorderDropDelegate<Item: Identifiable>: DropDelegate {
  let groupId: String
  let dropTargetItem: Item
  let items: [Item]
  @Binding var activeDrag: (groupId: String, sourceIndex: Int)?
  @Binding var hasChangedLocation: Bool
  let onReorder: (String, IndexSet, Int) -> Void

  func dropEntered(info: DropInfo) {
    guard activeDrag?.groupId == groupId,
      let sourceIndex = activeDrag?.sourceIndex,
      let toIndex = items.firstIndex(where: { $0.id == dropTargetItem.id })
    else { return }
    guard sourceIndex != toIndex else { return }
    hasChangedLocation = true
    let destination = toIndex > sourceIndex ? toIndex + 1 : toIndex
    if destination != sourceIndex && destination != sourceIndex + 1 {
      onReorder(groupId, IndexSet(integer: sourceIndex), destination)
      activeDrag = (groupId, destination)
    }
  }

  func dropUpdated(info: DropInfo) -> DropProposal? {
    DropProposal(operation: .move)
  }

  func performDrop(info: DropInfo) -> Bool {
    // On iPhone, dropEntered may not fire during drag; do reorder here when user releases
    if let sourceIndex = activeDrag?.sourceIndex,
      activeDrag?.groupId == groupId,
      let toIndex = items.firstIndex(where: { $0.id == dropTargetItem.id }),
      sourceIndex != toIndex
    {
      let destination = toIndex > sourceIndex ? toIndex + 1 : toIndex
      onReorder(groupId, IndexSet(integer: sourceIndex), destination)
    }
    hasChangedLocation = false
    activeDrag = nil
    return true
  }
}

private struct StepReorderDropOutsideDelegate: DropDelegate {
  @Binding var activeDrag: (groupId: String, sourceIndex: Int)?

  func dropUpdated(info: DropInfo) -> DropProposal? {
    DropProposal(operation: .move)
  }

  func performDrop(info: DropInfo) -> Bool {
    activeDrag = nil
    return true
  }
}

private struct FocusModeGroupsView<Item: Identifiable, Row: View>: View {
  let focusedGroups: [StepFlowGroup<Item>]
  @Binding var activeDrag: (groupId: String, sourceIndex: Int)?
  @Binding var hasChangedLocation: Bool
  let onReorder: (String, IndexSet, Int) -> Void
  let rowContent: (Item) -> Row

  var body: some View {
    LazyVStack(alignment: .leading, spacing: 0, pinnedViews: .sectionHeaders) {
      ForEach(focusedGroups) { group in
        if !group.items.isEmpty {
          Section {
            ForEach(Array(group.items.enumerated()), id: \.element.id) { index, item in
              rowContent(item)
                .padding(.vertical, 4)
                .contentShape(Rectangle())
                .opacity(
                  activeDrag?.groupId == group.id && activeDrag?.sourceIndex == index
                    && hasChangedLocation ? 0.5 : 1
                )
                .onDrag {
                  activeDrag = (group.id, index)
                  hasChangedLocation = false
                  return NSItemProvider(object: "\(group.id)|\(index)" as NSString)
                }
                .onDrop(
                  of: [.text, .plainText],
                  delegate: StepReorderDropDelegate(
                    groupId: group.id,
                    dropTargetItem: item,
                    items: group.items,
                    activeDrag: $activeDrag,
                    hasChangedLocation: $hasChangedLocation,
                    onReorder: onReorder
                  )
                )
            }
          } header: {
            DSRuleFlankedHeader(title: group.title, color: group.color)
          }
        }
      }
    }
  }
}

private struct FocusModeGroupsViewNoReorder<Item: Identifiable, Row: View>: View {
  let focusedGroups: [StepFlowGroup<Item>]
  let rowContent: (Item) -> Row

  var body: some View {
    LazyVStack(alignment: .leading, spacing: 0, pinnedViews: .sectionHeaders) {
      ForEach(focusedGroups) { group in
        if !group.items.isEmpty {
          Section {
            ForEach(group.items) { item in
              rowContent(item)
            }
          } header: {
            DSRuleFlankedHeader(title: group.title, color: group.color)
          }
        }
      }
    }
  }
}

/// Grey overlay shown over the steps section when wash hands gate is active.
/// Uses a clear band at the top so section headers like "Up Next" remain visible.
private struct StepsBlockedOverlay: View {
  private static let headerClearHeight: CGFloat = 56
  private static let grayOverlayTopPadding: CGFloat = -10

  var body: some View {
    VStack(spacing: 0) {
      Color.clear
        .frame(height: Self.headerClearHeight)
        .frame(maxWidth: .infinity)
      Color(.systemGray5)
        .opacity(0.65)
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .padding(.top, Self.grayOverlayTopPadding)
    }
    .frame(maxWidth: .infinity, maxHeight: .infinity)
    .overlay(
      VStack(spacing: 8) {
        Image(systemName: "hands.sparkles")
          .font(.largeTitle)
          .foregroundColor(.secondary)
        Text("Complete wash hands above to unlock steps")
          .font(.subheadline)
          .foregroundColor(.secondary)
          .multilineTextAlignment(.center)
          .padding(.horizontal)
      }
    )
    .allowsHitTesting(true)
  }
}

struct StepFlowSection<
  Item: Identifiable,
  Header: View,
  AllLead: View,
  FocusLead: View,
  Row: View
>: View {
  let showCompleted: Binding<Bool>
  let allSteps: [Item]
  let focusedGroups: [StepFlowGroup<Item>]
  let allowToggle: Bool
  let allStepsTitle: String
  let emptyMessage: String?
  let showStepsOverlay: Bool
  let headerContent: () -> Header
  let allModeLeadingContent: () -> AllLead
  let focusModeLeadingContent: () -> FocusLead
  let rowContent: (Item) -> Row
  let onReorder: ((String, IndexSet, Int) -> Void)?

  @State private var activeDrag: (groupId: String, sourceIndex: Int)?
  @State private var hasChangedLocation = false

  init(
    showCompleted: Binding<Bool>,
    allSteps: [Item],
    focusedGroups: [StepFlowGroup<Item>],
    allowToggle: Bool = true,
    allStepsTitle: String = "All Steps",
    emptyMessage: String? = nil,
    showStepsOverlay: Bool = false,
    onReorder: ((String, IndexSet, Int) -> Void)? = nil,
    @ViewBuilder headerContent: @escaping () -> Header,
    @ViewBuilder allModeLeadingContent: @escaping () -> AllLead,
    @ViewBuilder focusModeLeadingContent: @escaping () -> FocusLead,
    @ViewBuilder rowContent: @escaping (Item) -> Row
  ) {
    self.showCompleted = showCompleted
    self.allSteps = allSteps
    self.focusedGroups = focusedGroups
    self.allowToggle = allowToggle
    self.allStepsTitle = allStepsTitle
    self.emptyMessage = emptyMessage
    self.showStepsOverlay = showStepsOverlay
    self.headerContent = headerContent
    self.allModeLeadingContent = allModeLeadingContent
    self.focusModeLeadingContent = focusModeLeadingContent
    self.rowContent = rowContent
    self.onReorder = onReorder
  }

  @ViewBuilder
  private func stepsContentWithOverlay<Content: View>(
    @ViewBuilder content: () -> Content
  ) -> some View {
    if showStepsOverlay {
      ZStack(alignment: .topLeading) {
        content()
        StepsBlockedOverlay()
      }
    } else {
      content()
    }
  }

  @ViewBuilder
  private var focusModeContent: some View {
    if let onReorder {
      FocusModeGroupsView(
        focusedGroups: focusedGroups,
        activeDrag: $activeDrag,
        hasChangedLocation: $hasChangedLocation,
        onReorder: onReorder,
        rowContent: rowContent
      )
    } else {
      FocusModeGroupsViewNoReorder(
        focusedGroups: focusedGroups,
        rowContent: rowContent
      )
    }
  }

  var body: some View {
    VStack(alignment: .leading, spacing: 16) {
      HStack {
        headerContent()
        Spacer()

        if allowToggle {
          Button(showCompleted.wrappedValue ? "Focus mode" : "Complete view") {
            withAnimation {
              showCompleted.wrappedValue.toggle()
            }
          }
          .font(.caption)
          .buttonStyle(.bordered)
          .controlSize(.small)
        }
      }

      if allowToggle && showCompleted.wrappedValue {
        VStack(alignment: .leading, spacing: 8) {
          Text(allStepsTitle)
            .font(.subheadline)
            .fontWeight(.semibold)
            .foregroundColor(.secondary)
            .padding(.horizontal, 4)

          allModeLeadingContent()

          stepsContentWithOverlay(
            content: {
              ForEach(allSteps) { item in
                rowContent(item)
              }
            }
          )
        }
      } else {
        focusModeLeadingContent()

        stepsContentWithOverlay(
          content: {
            focusModeContent
              .onDrop(
                of: [.text, .plainText],
                delegate: StepReorderDropOutsideDelegate(activeDrag: $activeDrag)
              )
          }
        )
      }

      if let emptyMessage, allSteps.isEmpty {
        Text(emptyMessage)
          .font(.caption)
          .foregroundColor(.secondary)
      }
    }
  }
}
