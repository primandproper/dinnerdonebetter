import { describe, it, expect } from 'vitest';
import {
  type ResolvedSetting,
  type SettingDefinition,
  SettingKind,
  ValueSource,
} from '@primandproper/platform-client/settings/v1';
import { configurableSetting, configurableSettings, typedValue } from './resolutions';

function definition(overrides: Partial<SettingDefinition> = {}): SettingDefinition {
  return {
    createdAt: undefined,
    lastUpdatedAt: undefined,
    archivedAt: undefined,
    id: 'setting-1',
    name: 'user_temperature_unit',
    description: 'Which temperature unit recipes are shown in.',
    kind: SettingKind.SETTING_KIND_STRING,
    defaultValue: 'fahrenheit',
    enumeration: ['celsius', 'fahrenheit'],
    adminOnly: false,
    ...overrides,
  };
}

interface ResolutionOverrides extends Partial<Omit<ResolvedSetting, 'typedValue'>> {
  raw?: string;
}

function resolution({ raw = 'fahrenheit', ...overrides }: ResolutionOverrides = {}): ResolvedSetting {
  return {
    definition: definition(),
    value: undefined,
    // An unset setting has no value at all, not an empty one.
    typedValue: overrides.source === ValueSource.VALUE_SOURCE_UNSET ? undefined : { stringValue: raw },
    source: ValueSource.VALUE_SOURCE_DEFAULT,
    ...overrides,
  };
}

describe('configurableSetting', () => {
  it('starts on the answer the person chose', () => {
    const result = configurableSetting(resolution({ raw: 'celsius', source: ValueSource.VALUE_SOURCE_SUBJECT }));

    expect(result?.currentValue).toBe('celsius');
  });

  it('starts on the default when they have not chosen', () => {
    const result = configurableSetting(resolution({ raw: 'fahrenheit', source: ValueSource.VALUE_SOURCE_DEFAULT }));

    expect(result?.currentValue).toBe('fahrenheit');
  });

  it('falls back to the first option when nothing has answered', () => {
    // source "unset" means no answer and no default, so there is no value and
    // the picker has to start somewhere.
    const result = configurableSetting(resolution({ raw: '', source: ValueSource.VALUE_SOURCE_UNSET }));

    expect(result?.currentValue).toBe('celsius');
  });

  it('drops a setting that admits any value of its kind', () => {
    const result = configurableSetting(resolution({ definition: definition({ enumeration: [] }) }));

    expect(result).toBeNull();
  });

  it('drops a resolution carrying no definition', () => {
    const result = configurableSetting(resolution({ definition: undefined }));

    expect(result).toBeNull();
  });
});

describe('configurableSettings', () => {
  it('keeps the renderable settings in the order the server resolved them', () => {
    const first = definition({ id: 'setting-1', name: 'user_temperature_unit' });
    const unrenderable = definition({ id: 'setting-2', name: 'display_name', enumeration: [] });
    const last = definition({ id: 'setting-3', name: 'user_measurement_system', enumeration: ['metric', 'imperial'] });

    const result = configurableSettings([
      resolution({ definition: first, raw: 'celsius', source: ValueSource.VALUE_SOURCE_SUBJECT }),
      resolution({ definition: unrenderable, raw: 'Jeffrey', source: ValueSource.VALUE_SOURCE_SUBJECT }),
      resolution({ definition: last, raw: '', source: ValueSource.VALUE_SOURCE_UNSET }),
    ]);

    expect(result.map((item) => item.setting.id)).toEqual(['setting-1', 'setting-3']);
  });

  it('does not filter admin-only settings, because the server already has', () => {
    // An admin-only setting reaching this list means the caller is an
    // administrator entitled to see it. Filtering here would hide it from them.
    const adminOnly = definition({ id: 'setting-9', name: 'feature_gate', adminOnly: true });

    const result = configurableSettings([
      resolution({ definition: adminOnly, raw: 'celsius', source: ValueSource.VALUE_SOURCE_SUBJECT }),
    ]);

    expect(result.map((item) => item.setting.id)).toEqual(['setting-9']);
  });
});

describe('typedValue', () => {
  it('writes a text setting as text, even when it looks like a number', () => {
    expect(typedValue(SettingKind.SETTING_KIND_STRING, '12')).toEqual({ stringValue: '12' });
  });

  it('writes each other kind in its own case', () => {
    expect(typedValue(SettingKind.SETTING_KIND_BOOLEAN, 'false')).toEqual({ boolValue: false });
    expect(typedValue(SettingKind.SETTING_KIND_INTEGER, '-3')).toEqual({ intValue: -3 });
    expect(typedValue(SettingKind.SETTING_KIND_FLOAT, '1.5')).toEqual({ floatValue: 1.5 });
  });

  it('refuses text its kind cannot parse, rather than sending the server a guess', () => {
    expect(typedValue(SettingKind.SETTING_KIND_BOOLEAN, 'yes')).toBeNull();
    expect(typedValue(SettingKind.SETTING_KIND_INTEGER, '1.5')).toBeNull();
    expect(typedValue(SettingKind.SETTING_KIND_FLOAT, '')).toBeNull();
    expect(typedValue(SettingKind.SETTING_KIND_UNSPECIFIED, 'celsius')).toBeNull();
  });
});
