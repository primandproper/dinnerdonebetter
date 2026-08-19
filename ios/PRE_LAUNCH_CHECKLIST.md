# iOS App Pre-Launch Checklist

This document outlines all the steps required before submitting the Dinner Done Better iOS app to the App Store.

---

## Table of Contents

1. [Apple Developer Account Setup](#1-apple-developer-account-setup)
2. [App Store Connect Configuration](#2-app-store-connect-configuration)
3. [Xcode Project Configuration](#3-xcode-project-configuration)
4. [Deep Linking / Universal Links](#4-deep-linking--universal-links)
5. [Push Notifications](#5-push-notifications)
6. [App Privacy & Compliance](#6-app-privacy--compliance)
7. [Testing & Quality Assurance](#7-testing--quality-assurance)
8. [Final Submission](#8-final-submission)

---

## 1. Apple Developer Account Setup

### Prerequisites

- [ ] Enroll in Apple Developer Program ($99/year) at [developer.apple.com](https://developer.apple.com)
- [ ] Complete enrollment verification (may take 24-48 hours)

### Account Configuration

- [ ] Log in to [Apple Developer Portal](https://developer.apple.com/account)
- [ ] Note your **Team ID** (Membership → Team ID, 10-character alphanumeric string)
- [ ] Set up two-factor authentication if not already enabled

### Certificates & Provisioning

- [ ] Create a **Distribution Certificate** (Certificates, Identifiers & Profiles → Certificates)
- [ ] Create an **App ID** with bundle identifier `com.dinnerdonebetter.ios`
- [ ] Enable required capabilities on the App ID:
  - [ ] Associated Domains (for Universal Links)
  - [ ] Push Notifications (if using push)
  - [ ] Sign In with Apple (if using Apple SSO)
- [ ] Create a **Distribution Provisioning Profile**

---

## 2. App Store Connect Configuration

### Create App Record

- [ ] Log in to [App Store Connect](https://appstoreconnect.apple.com)
- [ ] Create new app (My Apps → + → New App)
- [ ] Set app name: "Dinner Done Better"
- [ ] Select bundle ID: `com.dinnerdonebetter.ios`
- [ ] Set primary language and category (Food & Drink)

### App Information

- [ ] Write app description (up to 4000 characters)
- [ ] Add keywords for App Store search optimization
- [ ] Set content rating (complete questionnaire)
- [ ] Add support URL
- [ ] Add privacy policy URL
- [ ] Set age rating

### App Media

- [ ] Create app icon (1024x1024px, no transparency)
- [ ] Create screenshots for required device sizes:
  - [ ] iPhone 6.7" (1290 x 2796px) - iPhone 15 Pro Max
  - [ ] iPhone 6.5" (1284 x 2778px) - iPhone 14 Plus
  - [ ] iPhone 5.5" (1242 x 2208px) - iPhone 8 Plus
  - [ ] iPad Pro 12.9" (2048 x 2732px) - if supporting iPad
- [ ] Create app preview videos (optional but recommended)

### Pricing & Availability

- [ ] Set pricing (Free or paid)
- [ ] Select availability by country/region
- [ ] Configure pre-orders (optional)

---

## 3. Xcode Project Configuration

### Project Settings

- [ ] Set correct Bundle Identifier: `com.dinnerdonebetter.ios`
- [ ] Set Version number (e.g., `1.0.0`)
- [ ] Set Build number (e.g., `1`)
- [ ] Select your Team in Signing & Capabilities
- [ ] Enable "Automatically manage signing" or configure manual signing

### Build Settings

- [ ] Set deployment target (minimum iOS version)
- [ ] Configure app icons in Assets.xcassets
- [ ] Configure launch screen
- [ ] Remove any debug/development code flags for release

### Capabilities

- [ ] Add required capabilities in Signing & Capabilities:
  - [ ] Associated Domains (for Universal Links)
  - [ ] Push Notifications (if applicable)
  - [ ] Sign in with Apple (if applicable)
  - [ ] Background Modes (if applicable)

### Info.plist

- [ ] Add required usage descriptions:
  - [ ] `NSCameraUsageDescription` (if using camera)
  - [ ] `NSPhotoLibraryUsageDescription` (if accessing photos)
  - [ ] `NSLocationWhenInUseUsageDescription` (if using location)
- [ ] Configure URL schemes if needed
- [ ] Set `ITSAppUsesNonExemptEncryption` to appropriate value

---

## 4. Deep Linking / Universal Links

Universal Links allow the app to open when users tap invitation links, password reset links, etc.

### Backend Configuration

The Apple App Site Association (AASA) endpoint is configured in the backend.

- [ ] Set your Team ID for prod: add `teamID` to the `appleAppSiteAssociation` section in `backend/deploy/environments/prod/kustomize/configs/api_service_config.json`, or set the `DINNER_DONE_BETTER_AASA_TEAM_ID` environment variable in your cluster secrets

- [ ] Deploy backend with updated configuration

### ADP (Apple Developer Portal)

- [ ] Enable **Associated Domains** capability on your App ID
  - Go to Certificates, Identifiers & Profiles → Identifiers
  - Select your App ID
  - Enable Associated Domains
  - Save

### Xcode Configuration

- [ ] Add Associated Domains capability in Xcode (Signing & Capabilities → + Capability)
- [ ] Verify entitlements file (`ios/ios.entitlements`) contains correct domains:

```xml
<key>com.apple.developer.associated-domains</key>
<array>
    <string>applinks:www.dinnerdonebetter.dev</string>
    <string>applinks:www.dinnerdonebetter.com</string>
</array>
```

### Verification

- [ ] Verify AASA file is accessible at `https://www.dinnerdonebetter.dev/.well-known/apple-app-site-association`
- [ ] Verify AASA file is accessible at `https://www.dinnerdonebetter.com/.well-known/apple-app-site-association` (production)
- [ ] Validate with Apple's tool: <https://search.developer.apple.com/appsearch-validation-tool/>
- [ ] Test on physical device (Universal Links don't work in Simulator):
  - [ ] Send test invitation email
  - [ ] Tap link and verify app opens
  - [ ] Verify registration screen shows invitation data

### Testing Without Apple Developer Account

You can test URL parsing logic without an Apple Developer account:

```swift
#if DEBUG
// In a debug view or test
deepLinkHandler.simulateInvitationLink(
    invitationID: "test-id",
    token: "test-token"
)
#endif
```

- [ ] Run unit tests: `iosTests/DeepLinkHandlerTests.swift`

### Troubleshooting

If Universal Links aren't working:

1. Verify AASA file is at `/.well-known/apple-app-site-association`
2. Ensure HTTPS with valid certificate
3. Check App ID matches `TEAMID.bundleid`
4. Clear Safari cache on device
5. Delete and reinstall app
6. Wait up to 24 hours (Apple caches AASA files)

---

## 5. Push Notifications

(Skip this section if not implementing push notifications)

### Developer Portal

- [ ] Enable Push Notifications on App ID
- [ ] Create APNs Key or Certificate

### Back end

- [ ] Configure push notification service (e.g., Firebase, AWS SNS)
- [ ] Store APNs credentials securely

### Xcode

- [ ] Add Push Notifications capability
- [ ] Add Background Modes → Remote notifications (if needed)

### Implementation

- [ ] Request notification permission from user
- [ ] Register for remote notifications
- [ ] Handle device token registration with backend
- [ ] Handle incoming notifications

---

## 6. App Privacy & Compliance

### Privacy Policy

- [ ] Create privacy policy document
- [ ] Host privacy policy at accessible URL
- [ ] Add URL to App Store Connect

### App Privacy Details (App Store Connect)

- [ ] Complete "App Privacy" section:
  - [ ] Data types collected
  - [ ] Data linked to user identity
  - [ ] Data used for tracking
  - [ ] Third-party data sharing

### Data Collection Disclosure

Document what data is collected:

- [ ] Account information (email, username)
- [ ] User content (recipes, meal plans)
- [ ] Usage data (analytics)
- [ ] Device identifiers

### Compliance

- [ ] Export compliance (encryption usage)
- [ ] GDPR compliance (for EU users)
- [ ] CCPA compliance (for California users)
- [ ] Age restrictions (if applicable)

---

## 7. Testing & Quality Assurance

### Unit Tests

- [ ] All unit tests passing
- [ ] Run: `cd ios && make unit_test`

### UI Tests

- [ ] UI tests passing
- [ ] Run: `cd ios && make ui_test`

### Manual Testing Checklist

#### Authentication

- [ ] User registration works
- [ ] User login works
- [ ] TOTP/2FA works (if enabled)
- [ ] Logout works
- [ ] Password reset flow works

#### Core Features

- [ ] Browse recipes
- [ ] Create meal plans
- [ ] View meal plan details
- [ ] Grocery list generation
- [ ] Recipe performance/cooking mode

#### Account Management

- [ ] View account settings
- [ ] Update profile
- [ ] Account invitation acceptance (via deep link)

#### Edge Cases

- [ ] No network connection handling
- [ ] Session expiration handling
- [ ] Large data sets (many recipes, meal plans)
- [ ] Different device sizes

### TestFlight Beta Testing

- [ ] Upload build to TestFlight
- [ ] Internal testing with team
- [ ] External beta testing (optional)
- [ ] Address feedback from beta testers

### Performance

- [ ] App launch time acceptable
- [ ] No memory leaks
- [ ] Smooth scrolling/animations
- [ ] Battery usage reasonable

---

## 8. Final Submission

### Pre-Submission Checks

- [ ] All tests passing
- [ ] No compiler warnings
- [ ] No console errors in release build
- [ ] App runs correctly on oldest supported iOS version
- [ ] App runs correctly on latest iOS version

### Archive & Upload

- [ ] Create Archive in Xcode (Product → Archive)
- [ ] Validate archive
- [ ] Upload to App Store Connect

### App Store Connect

- [ ] Select build for review
- [ ] Complete all required metadata
- [ ] Answer export compliance questions
- [ ] Submit for review

### Post-Submission

- [ ] Monitor app review status
- [ ] Respond to any reviewer questions promptly
- [ ] Prepare marketing materials for launch
- [ ] Plan announcement/launch communications

---

## Files Reference

### Deep Linking Files

#### iOS App

- `ios/ios/ios.entitlements` - Associated Domains configuration
- `ios/ios/Services/DeepLink/DeepLinkHandler.swift` - URL parsing logic
- `ios/ios/App/iosApp.swift` - onOpenURL handler
- `ios/ios/Views/ContentView.swift` - Deep link routing
- `ios/ios/Views/Auth/RegisterView.swift` - Invitation data display
- `ios/ios/Services/Networking/APIConfiguration.swift` - Environment configuration
- `ios/iosTests/DeepLinkHandlerTests.swift` - Unit tests

#### Backend

- `backend/internal/config/configs.go` - `AppleAppSiteAssociationConfig` struct
- `backend/internal/build/services/api/http/http_routes.go` - AASA endpoint
- `backend/internal/build/services/api/http/config.go` - Wire provider
- `backend/deploy/environments/prod/kustomize/configs/api_service_config.json` - Prod API config (including AppleAppSiteAssociation)
- `backend/internal/config/environments/localdev.go` - Local dev config

---

## Quick Reference

| Item                       | Location/Command                                                |
|----------------------------|-----------------------------------------------------------------|
| Team ID                    | Apple Developer Portal → Membership                             |
| Bundle ID                  | `com.dinnerdonebetter.ios`                                      |
| Run unit tests             | `cd ios && make unit_test`                                      |
| Run UI tests               | `cd ios && make ui_test`                                        |
| Regenerate backend configs | `cd backend && make configs`                                    |
| AASA validation            | <https://search.developer.apple.com/appsearch-validation-tool/> |
| App Store Connect          | <https://appstoreconnect.apple.com>                             |
| Apple Developer Portal     | <https://developer.apple.com/account>                           |
