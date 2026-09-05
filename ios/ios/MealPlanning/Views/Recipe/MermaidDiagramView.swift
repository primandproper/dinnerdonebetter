//
//  MermaidDiagramView.swift
//  ios
//
//  Created by Auto on 2/28/25.
//

import BeautifulMermaid
import SwiftUI

/// A SwiftUI view that renders Mermaid diagram source using BeautifulMermaid.
/// Handles empty source, loading, and render errors.
struct MermaidDiagramView: View {
  let source: String
  var theme: DiagramTheme?

  @Environment(\.colorScheme) private var colorScheme
  @SwiftUI.State private var renderedImage: UIImage?
  @SwiftUI.State private var renderError: Error?

  private var effectiveTheme: DiagramTheme {
    if let theme {
      return theme
    }
    return colorScheme == .dark ? .tokyoNight : .zincLight
  }

  var body: some View {
    Group {
      if source.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty {
        placeholderView
      } else if let error = renderError {
        errorView(error: error)
      } else if let image = renderedImage {
        Image(uiImage: image)
          .resizable()
          .scaledToFit()
          .frame(minHeight: 200)
      } else {
        ProgressView("Rendering diagram...")
          .frame(maxWidth: .infinity)
          .frame(minHeight: 200)
      }
    }
    .task(id: source) {
      await renderDiagram()
    }
  }

  private var placeholderView: some View {
    Text("No diagram available")
      .font(.subheadline)
      .foregroundColor(.secondary)
      .frame(maxWidth: .infinity)
      .frame(minHeight: 120)
  }

  private func errorView(error: Error) -> some View {
    VStack(spacing: 8) {
      Image(systemName: "exclamationmark.triangle")
        .font(.title2)
        .foregroundColor(.orange)
      Text("Diagram could not be rendered")
        .font(.subheadline)
        .foregroundColor(.secondary)
      Text(error.localizedDescription)
        .font(.caption)
        .foregroundColor(.secondary)
        .multilineTextAlignment(.center)
    }
    .frame(maxWidth: .infinity)
    .frame(minHeight: 120)
    .padding()
  }

  @MainActor
  private func renderDiagram() async {
    renderError = nil
    renderedImage = nil

    let trimmed = source.trimmingCharacters(in: .whitespacesAndNewlines)
    guard !trimmed.isEmpty else { return }

    let themeToUse = effectiveTheme
    do {
      let image = try await Task.detached {
        try MermaidRenderer.renderImage(source: trimmed, theme: themeToUse)
      }.value

      if let image {
        renderedImage = image
      } else {
        renderError = NSError(
          domain: "MermaidDiagramView",
          code: 1,
          userInfo: [NSLocalizedDescriptionKey: "Diagram could not be parsed"]
        )
      }
    } catch {
      renderError = error
    }
  }
}
