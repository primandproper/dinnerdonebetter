import type { PageServerLoad } from './$types';
import { listProducts } from '$lib/grpc/clients';
import { QueryFilter } from '@primandproper/platform-client/filtering/v1';

export const load: PageServerLoad = async ({ locals }) => {
  const session = locals.session;
  try {
    const res = (await listProducts(session, { filter: QueryFilter.create({ maxResponseSize: 100 }) })) as {
      results?: Array<{ id?: string; name?: string }>;
    };
    return { products: res?.results ?? [] };
  } catch (e) {
    return {
      products: [],
      error: e instanceof Error ? e.message : 'Failed to load products',
    };
  }
};
