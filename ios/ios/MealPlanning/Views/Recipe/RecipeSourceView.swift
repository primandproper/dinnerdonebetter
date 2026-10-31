//
//  RecipeSourceView.swift
//  ios
//
//  Displays recipe source as a clickable link when it's a URL, or static attribution text otherwise.
//

import DesignSystem
import SwiftUI

struct RecipeSourceView: View {
  let source: String

  private var sourceURL: URL? {
    guard let url = URL(string: source),
      url.scheme == "http" || url.scheme == "https"
    else { return nil }
    return url
  }

  var body: some View {
    if source.isEmpty {
      EmptyView()
    } else if let url = sourceURL {
      Link(destination: url) {
        Label("View source", systemImage: "arrow.up.right.square")
          .font(DSTheme.Typography.caption)
          .foregroundColor(DSTheme.Colors.textSecondary)
      }
    } else {
      Label(source, systemImage: "link")
        .font(DSTheme.Typography.caption)
        .foregroundColor(DSTheme.Colors.textSecondary)
    }
  }
}
