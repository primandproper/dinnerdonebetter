import type { PageServerLoad } from './$types';
import { listSettingDefinitions } from '$lib/grpc/clients';
import { QueryFilter } from '@dinnerdonebetter/api-client';
import type { SettingDefinition } from '@primandproper/platform-client/settings/v1';

const DEFAULT_LIST_FILTER = QueryFilter.create({ maxResponseSize: 100 });

/**
 * The catalog has no search behind it: platform's settings store lists
 * definitions and nothing more, so the `q` box filters the page it already
 * fetched rather than asking the server a second question. A deployment's
 * catalog is small enough for that — the list is capped at the same 100 rows
 * it always was — and a filter that returns nothing means nothing on this page
 * matched, not that the setting does not exist.
 */
function matching(settings: SettingDefinition[], query: string): SettingDefinition[] {
  if (query === '') return settings;
  const needle = query.toLowerCase();
  return settings.filter((setting) => setting.name.toLowerCase().includes(needle));
}

export const load: PageServerLoad = async ({ locals, url }) => {
  const token = locals.accessToken;
  if (!token) {
    return { settings: [], query: '', error: 'Not authenticated' };
  }
  const query = url.searchParams.get('q')?.trim() ?? '';
  try {
    const res = await listSettingDefinitions(token, { filter: DEFAULT_LIST_FILTER });
    return { settings: matching(res.results, query), query };
  } catch (e) {
    return {
      settings: [],
      query,
      error: e instanceof Error ? e.message : 'Failed to load settings',
    };
  }
};
