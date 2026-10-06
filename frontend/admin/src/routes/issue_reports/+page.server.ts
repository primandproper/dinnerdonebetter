import type { PageServerLoad } from './$types';
import { listIssueReportsAcrossScopes } from '$lib/grpc/clients';
import { QueryFilter } from '@dinnerdonebetter/api-client';
import { toOperatorRows } from './rows';

export const load: PageServerLoad = async ({ locals }) => {
  const session = locals.session;
  try {
    const res = await listIssueReportsAcrossScopes(session, {
      filter: QueryFilter.create({ maxResponseSize: 100 }),
    });
    return { reports: toOperatorRows(res?.results ?? []) };
  } catch (e) {
    return {
      reports: [],
      error: e instanceof Error ? e.message : 'Failed to load issue reports',
    };
  }
};
