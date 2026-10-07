# Analytics (Event Capture)

Event capture follows the same architecture as the Go backend: a generic interface with flexible implementations and rigid exclusive use.

## Implementations

- **SegmentEventReporter**: Sends directly to Segment via Analytics-Swift (default when `SegmentWriteKey` is set).

## Deployment

The deploy workflow (`.github/workflows/deploy_ios.yaml`) creates `Secrets.xcconfig` at build time by writing values from GitHub Actions secrets. Add **`PROD_IOS_SEGMENT_WRITE_KEY`** to your repository secrets (Settings → Secrets and variables → Actions). The value comes from Segment: Connections → Sources → [Apple source] → Write Key.

## Exclusive Use Rule

**All event reporting MUST go through the `EventReporter` protocol.** No direct Analytics-Swift or Segment SDK usage anywhere except `SegmentEventReporter.swift`.

- Only `SegmentEventReporter.swift` may `import Segment`
- All other files must use `EventReporter` (obtained via `AnalyticsConfiguration.provideEventReporter()` or `@Environment(EventReporterService.self)`)
- Enforce via code review

## Adding New Events

Use `EventReporter.track(event:properties:)` for events and `identify(userID:properties:)` when the user identity is known. Obtain the reporter from `AnalyticsConfiguration.provideEventReporter()` or from the SwiftUI environment.
