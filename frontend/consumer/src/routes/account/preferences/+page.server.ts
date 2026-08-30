import { redirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { QueryFilter } from '@dinnerdonebetter/api-client';
import {
  getActiveAccount,
  getServiceSettings,
  getServiceSettingConfigurationsForUser,
  createServiceSettingConfiguration,
  updateServiceSettingConfiguration,
} from '$lib/grpc/clients';
import type {
  ServiceSetting,
  ServiceSettingConfiguration,
} from '@dinnerdonebetter/api-client/settings/settings_messages';

interface ConfigurableSetting {
  setting: ServiceSetting;
  config: ServiceSettingConfiguration | null;
  currentValue: string;
}

function mergeAndFilterSettings(
  settings: ServiceSetting[],
  configs: ServiceSettingConfiguration[],
  accountId: string,
): ConfigurableSetting[] {
  const configsForAccount = new Map<string, ServiceSettingConfiguration>();
  for (const cfg of configs) {
    if (cfg.belongsToAccount === accountId && cfg.serviceSetting) {
      configsForAccount.set(cfg.serviceSetting.id, cfg);
    }
  }

  const result: ConfigurableSetting[] = [];
  for (const setting of settings) {
    if (setting.type !== 'user' || setting.adminsOnly) continue;
    if (!setting.enumeration?.length) continue;

    const config = configsForAccount.get(setting.id) ?? null;
    let currentValue = '';
    if (config?.value) {
      currentValue = config.value;
    } else if (setting.defaultValue) {
      currentValue = setting.defaultValue;
    } else if (setting.enumeration?.length) {
      currentValue = setting.enumeration[0];
    }
    if (!currentValue && setting.enumeration?.length) {
      currentValue = setting.enumeration[0];
    }

    result.push({ setting, config, currentValue });
  }
  return result;
}

export const load: PageServerLoad = async ({ locals, url }) => {
  const token = locals.oauthToken;
  if (!token) {
    return { configurableSettings: [], error: null, updated: false };
  }

  try {
    const activeRes = await getActiveAccount(token);
    const account = activeRes.result;
    if (!account) {
      return {
        configurableSettings: [],
        error: null,
        updated: false,
      };
    }

    const [settingsRes, configsRes] = await Promise.all([
      getServiceSettings(token, { filter: QueryFilter.create({ maxResponseSize: 100 }) }),
      getServiceSettingConfigurationsForUser(token, { filter: QueryFilter.create({ maxResponseSize: 100 }) }),
    ]);

    const configurableSettings = mergeAndFilterSettings(
      settingsRes.results ?? [],
      configsRes.results ?? [],
      account.id,
    );

    const error = url.searchParams.get('error');
    const updated = url.searchParams.get('updated') === '1';

    return { configurableSettings, error, updated };
  } catch {
    return {
      configurableSettings: [],
      error: 'server',
      updated: false,
    };
  }
};

const _errorMessages: Record<string, string> = {
  invalid: 'Invalid input. Please try again.',
  update_failed: 'Failed to save preference. Please try again.',
  server: 'Something went wrong. Please try again.',
};

export const actions: Actions = {
  update: async ({ request, locals }) => {
    const token = locals.oauthToken;
    if (!token) throw redirect(302, '/login');

    const formData = await request.formData();
    const settingId = (formData.get('setting_id') as string)?.trim() ?? '';
    const configId = (formData.get('config_id') as string)?.trim() ?? '';
    const value = (formData.get('value') as string)?.trim() ?? '';

    if (!settingId || !value) {
      throw redirect(302, '/account/preferences?error=invalid');
    }

    try {
      if (configId) {
        await updateServiceSettingConfiguration(token, {
          serviceSettingConfigurationId: configId,
          input: { value },
        });
      } else {
        await createServiceSettingConfiguration(token, {
          input: {
            serviceSettingId: settingId,
            value,
            notes: '',
          },
        });
      }
      throw redirect(302, '/account/preferences?updated=1');
    } catch (e) {
      if (e && typeof e === 'object' && 'status' in e && (e as { status: number }).status === 302) {
        throw e;
      }
      throw redirect(302, '/account/preferences?error=update_failed');
    }
  },
};
