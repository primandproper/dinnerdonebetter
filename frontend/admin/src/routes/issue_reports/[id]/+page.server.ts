import type { PageServerLoad } from './$types';
import { getIssueReport } from '$lib/grpc/clients';

export const load: PageServerLoad = async ({ locals, params }) => {
  const session = locals.session;
  const id = params.id;
  try {
    // GetReport reads only the caller's own account, so a report listed across accounts is not
    // found here unless it is the operator's. See platform-go#1149.
    const res = (await getIssueReport(session, { reportId: id })) as { result?: Record<string, unknown> };
    return { report: res?.result ?? null };
  } catch (e) {
    return {
      report: null,
      error: e instanceof Error ? e.message : 'Failed to load issue report',
    };
  }
};
