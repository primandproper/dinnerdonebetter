//
//  BrandingThemeTests.swift
//  iosTests
//

import DesignSystem
@testable import ios
import Testing

struct BrandingThemeTests {
  // The app's own views read the static tokens, which always return `DSTheme.default`, while the
  // platform-ui-swift components follow `Branding.theme` from the environment. If the two differ,
  // the components and the views around them stop matching.
  @Test("Branding theme matches the default the static tokens return")
  func testBrandingThemeMatchesDefault() {
    #expect(String(reflecting: Branding.theme) == String(reflecting: DSTheme.default))
  }
}
