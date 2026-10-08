import type { PageServerLoad } from './$types';
import { listOAuth2Clients } from '$lib/grpc/clients';
import { QueryFilter } from '@primandproper/platform-client/filtering/v1';

export const load: PageServerLoad = async ({ locals }) => {
  const session = locals.session;
  try {
    const res = (await listOAuth2Clients(session, { filter: QueryFilter.create({ maxResponseSize: 100 }) })) as {
      results?: Array<{ id?: string; clientId?: string; name?: string }>;
    };
    return { clients: res?.results ?? [] };
  } catch (e) {
    return {
      clients: [],
      error: e instanceof Error ? e.message : 'Failed to load OAuth2 clients',
    };
  }
};
