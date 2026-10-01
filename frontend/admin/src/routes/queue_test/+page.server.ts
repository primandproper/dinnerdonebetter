import { fail } from '@sveltejs/kit';
import type { Actions } from './$types';
import { testQueueMessage } from '$lib/grpc/clients';

export const actions: Actions = {
  default: async ({ request, locals }) => {
    const session = locals.session;
    const formData = await request.formData();
    const queueName = (formData.get('queue_name') as string)?.trim() ?? '';
    if (!queueName) {
      return fail(400, { error: 'Queue name is required' });
    }
    try {
      const res = (await testQueueMessage(session, { queueName })) as {
        success?: boolean;
        testId?: string;
        roundTripMs?: number;
      };
      return {
        success: res?.success ?? false,
        testId: res?.testId ?? '',
        roundTripMs: res?.roundTripMs ?? 0,
      };
    } catch (e) {
      return fail(500, {
        error: e instanceof Error ? e.message : 'Queue test failed',
      });
    }
  },
};
