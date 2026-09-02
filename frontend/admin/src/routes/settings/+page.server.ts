import type { PageServerLoad } from './$types';
import { getSettingDefinitions } from '$lib/grpc/clients';
import { QueryFilter } from '@dinnerdonebetter/api-client';

const DEFAULT_LIST_FILTER = QueryFilter.create({ maxResponseSize: 100 });

interface SettingDefinitionRow {
  id?: string;
  name?: string;
  kind?: string;
  adminOnly?: boolean;
}

/**
 * The catalog has no search behind it: platform's settings store lists
 * definitions and nothing more, so the `q` box filters the page it already
 * fetched rather than asking the server a second question. A deployment's
 * catalog is small enough for that — the list is capped at the same 100 rows
 * it always was — and a filter that returns nothing means nothing on this page
 * matched, not that the setting does not exist.
 */
function matching(settings: SettingDefinitionRow[], query: string): SettingDefinitionRow[] {
  if (query === '') return settings;
  const needle = query.toLowerCase();
  return settings.filter((setting) => (setting.name ?? '').toLowerCase().includes(needle));
}

export const load: PageServerLoad = async ({ locals, url }) => {
  const token = locals.accessToken;
  if (!token) {
    return { settings: [], query: '', error: 'Not authenticated' };
  }
  const query = url.searchParams.get('q')?.trim() ?? '';
  try {
    const res = (await getSettingDefinitions(token, { filter: DEFAULT_LIST_FILTER })) as {
      results?: SettingDefinitionRow[];
    };
    return { settings: matching(res?.results ?? [], query), query };
  } catch (e) {
    return {
      settings: [],
      query,
      error: e instanceof Error ? e.message : 'Failed to load settings',
    };
  }
};
