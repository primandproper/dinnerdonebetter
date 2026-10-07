//
//  EventReporterService.swift
//  ios
//
//  Observable holder for EventReporter so it can be injected via SwiftUI Environment.
//  All event reporting must go through the reporter; no direct Segment usage elsewhere.
//

import Foundation
import SwiftUI

@Observable
final class EventReporterService {
  /// The shared event reporter. Use this for all analytics.
  let reporter: any EventReporter

  init() {
    reporter = AnalyticsConfiguration.provideEventReporter()
  }
}
