import { textFromTypedValue } from '@primandproper/platform-client';
import {
  type ResolvedSetting,
  type SettingDefinition,
  type SettingSubject,
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

/**
 * Pair one resolved setting with the value its picker should start on, or drop
 * it if this page cannot render it.
 *
 * A resolution whose source is "unset" is a setting nobody has answered that has
 * no default, so there is no value to show and the first enumerated option
 * stands in — which is what the picker would fall back to anyway. Every other
 * source already carries its answer in the typed value, whether the person chose
 * it ("subject") or the definition's default did ("default"); the distinction is
 * the server's to make and this no longer reimplements it. The value reads as the
 * server stores it, which is the form the enumeration is compared in.
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
      : (textFromTypedValue(resolution.typedValue) ?? '');

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
