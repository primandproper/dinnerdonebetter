import { describe, it, expect } from 'vitest';
import type { SettingDefinition, SettingResolution } from '@dinnerdonebetter/api-client/settings/settings_messages';
import { configurableSetting, configurableSettings } from './resolutions';

function definition(overrides: Partial<SettingDefinition> = {}): SettingDefinition {
  return {
    createdAt: undefined,
    lastUpdatedAt: undefined,
    archivedAt: undefined,
    id: 'setting-1',
    name: 'user_temperature_unit',
    description: 'Which temperature unit recipes are shown in.',
    kind: 'string',
    defaultValue: 'fahrenheit',
    enumeration: ['celsius', 'fahrenheit'],
    adminOnly: false,
    ...overrides,
  };
}

function resolution(overrides: Partial<SettingResolution> = {}): SettingResolution {
  return {
    definition: definition(),
    value: undefined,
    raw: 'fahrenheit',
    source: 'default',
    ...overrides,
  };
}

describe('configurableSetting', () => {
  it('starts on the answer the person chose', () => {
    const result = configurableSetting(resolution({ raw: 'celsius', source: 'subject' }));

    expect(result?.currentValue).toBe('celsius');
  });

  it('starts on the default when they have not chosen', () => {
    const result = configurableSetting(resolution({ raw: 'fahrenheit', source: 'default' }));

    expect(result?.currentValue).toBe('fahrenheit');
  });

  it('falls back to the first option when nothing has answered', () => {
    // source "unset" means no answer and no default, so `raw` is empty and the
    // picker has to start somewhere.
    const result = configurableSetting(resolution({ raw: '', source: 'unset' }));

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
      resolution({ definition: first, raw: 'celsius', source: 'subject' }),
      resolution({ definition: unrenderable, raw: 'Jeffrey', source: 'subject' }),
      resolution({ definition: last, raw: '', source: 'unset' }),
    ]);

    expect(result.map((item) => item.setting.id)).toEqual(['setting-1', 'setting-3']);
  });

  it('does not filter admin-only settings, because the server already has', () => {
    // An admin-only setting reaching this list means the caller is an
    // administrator entitled to see it. Filtering here would hide it from them.
    const adminOnly = definition({ id: 'setting-9', name: 'feature_gate', adminOnly: true });

    const result = configurableSettings([resolution({ definition: adminOnly, raw: 'celsius', source: 'subject' })]);

    expect(result.map((item) => item.setting.id)).toEqual(['setting-9']);
  });
});
