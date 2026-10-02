import {
  type ResolvedSetting,
  type SettingDefinition,
  SettingKind,
  type SettingSubject,
  type TypedValue,
  ValueSource,
} from '@primandproper/platform-client/settings/v1';

/** A setting this page can render, and the value its picker should start on. */
export interface ConfigurableSetting {
  setting: SettingDefinition;
  currentValue: string;
}

/**
 * The subject a person's own preferences are filed against. This deployment has
 * one subject type, the user, and the server refuses any subject but the caller
 * themselves, so there is nothing here to choose.
 */
export function settingsSubject(userId: string): SettingSubject {
  return { type: 'user', id: userId };
}

/** The text a picker compares against its options, whichever kind the value was. */
export function rawOf(value: TypedValue | undefined): string {
  if (value?.stringValue !== undefined) return value.stringValue;
  if (value?.boolValue !== undefined) return String(value.boolValue);
  if (value?.intValue !== undefined) return String(value.intValue);
  if (value?.floatValue !== undefined) return String(value.floatValue);
  return '';
}

/**
 * The typed value a write carries, from the text a form submitted. The server
 * checks the case against the setting's kind and refuses a mismatch, so a text
 * setting has to be written as text even when its answer looks like a number.
 */
export function typedValue(kind: SettingKind, raw: string): TypedValue | null {
  switch (kind) {
    case SettingKind.SETTING_KIND_STRING:
      return { stringValue: raw };
    case SettingKind.SETTING_KIND_BOOLEAN:
      return raw === 'true' || raw === 'false' ? { boolValue: raw === 'true' } : null;
    case SettingKind.SETTING_KIND_INTEGER:
      return /^-?\d+$/.test(raw) ? { intValue: Number(raw) } : null;
    case SettingKind.SETTING_KIND_FLOAT:
      return raw !== '' && Number.isFinite(Number(raw)) ? { floatValue: Number(raw) } : null;
    default:
      return null;
  }
}

/**
 * Pair one resolved setting with the value its picker should start on, or drop
 * it if this page cannot render it.
 *
 * A resolution whose source is "unset" is a setting nobody has answered that has
 * no default, so there is no value to show and the first enumerated option
 * stands in — which is what the picker would fall back to anyway. Every other
 * source already carries its answer in the typed value, whether the person chose
 * it ("subject") or the definition's default did ("default"); the distinction is
 * the server's to make and this no longer reimplements it.
 */
export function configurableSetting(resolution: ResolvedSetting): ConfigurableSetting | null {
  const setting = resolution.definition;
  if (!setting) return null;

  // A setting with no enumeration admits any value of its kind, and this page
  // only knows how to render a picker over a fixed set.
  if (!setting.enumeration?.length) return null;

  const currentValue =
    resolution.source === ValueSource.VALUE_SOURCE_UNSET
      ? (setting.enumeration[0] ?? '')
      : rawOf(resolution.typedValue);

  return { setting, currentValue };
}

/**
 * The settings a preferences page shows, in the order the server resolved them.
 *
 * Admin-only settings are absent rather than filtered here: the server decides
 * which settings a caller may see, so this never had to know who is asking.
 */
export function configurableSettings(resolutions: ResolvedSetting[]): ConfigurableSetting[] {
  return resolutions.map(configurableSetting).filter((item): item is ConfigurableSetting => item !== null);
}
