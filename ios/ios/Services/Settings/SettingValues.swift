//
//  SettingValues.swift
//  ios
//
//  Whose settings a screen reads and writes. Converting between a typed value and a form's
//  text is platform-client's (`TypedValue(text:kind:)` and `TypedValue.text`); the subject's
//  vocabulary is this application's, so it stays here.
//

import Foundation
import PlatformClient

typealias SettingDefinition = Primandproper_Platform_Settings_V1_SettingDefinition
typealias ResolvedSetting = Primandproper_Platform_Settings_V1_ResolvedSetting

enum SettingValues {
  /// Whose settings these are: always the signed-in user's own. The server refuses any other
  /// subject, so this is not a choice the app makes.
  static func subject(userID: String) -> Primandproper_Platform_Settings_V1_SettingSubject {
    var subject = Primandproper_Platform_Settings_V1_SettingSubject()
    subject.type = "user"
    subject.id = userID
    return subject
  }
}
