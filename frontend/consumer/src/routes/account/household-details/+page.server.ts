import { redirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { holds, Permission } from '$lib/auth/permissions';
import { getActiveAccount, getPrincipal, updateAccount } from '$lib/grpc/clients';

export const load: PageServerLoad = async ({ locals, url }) => {
  const session = locals.session;
  try {
    // One read answers both: the active account, and what the caller may do in it.
    const principal = await getPrincipal(session);
    const account = principal.activeAccount ?? null;

    if (!account) {
      return {
        account: null,
        canEdit: false,
        error: null,
        updated: false,
      };
    }

    const canEdit = holds(principal, Permission.updateAccount);
    if (!canEdit) {
      throw redirect(302, '/account/settings');
    }

    const error = url.searchParams.get('error');
    const updated = url.searchParams.get('updated') === '1';

    return { account, canEdit, error, updated };
  } catch (e) {
    if (e && typeof e === 'object' && 'status' in e && (e as { status: number }).status === 302) {
      throw e;
    }
    return {
      account: null,
      canEdit: false,
      error: 'server',
      updated: false,
    };
  }
};

export const actions: Actions = {
  update: async ({ request, locals }) => {
    const session = locals.session;
    const account = await getActiveAccount(session);
    if (!account) {
      throw redirect(302, '/account/household-details?error=server');
    }

    const formData = await request.formData();
    const name = (formData.get('name') as string)?.trim() ?? '';
    const contactPhone = (formData.get('contact_phone') as string)?.trim() ?? '';
    const addressLine1 = (formData.get('address_line_1') as string)?.trim() ?? '';
    const addressLine2 = (formData.get('address_line_2') as string)?.trim() ?? '';
    const city = (formData.get('city') as string)?.trim() ?? '';
    const state = (formData.get('state') as string)?.trim() ?? '';
    const zipCode = (formData.get('zip_code') as string)?.trim() ?? '';
    const country = (formData.get('country') as string)?.trim() ?? '';

    if (!name) {
      throw redirect(302, '/account/household-details?error=invalid_name');
    }

    try {
      await updateAccount(session, {
        accountId: account.id,
        input: {
          name,
          // One BillingAddress in place of six fields of the account, with the contact
          // phone on it. belongsToUser is not sent at all — ownership moves through
          // TransferAccountOwnership and is not something an update may carry.
          billingAddress: {
            line1: addressLine1,
            line2: addressLine2,
            city,
            state,
            postalCode: zipCode,
            country,
            phone: contactPhone,
          },
        },
      });
      throw redirect(302, '/account/household-details?updated=1');
    } catch (e) {
      if (e && typeof e === 'object' && 'status' in e && (e as { status: number }).status === 302) {
        throw e;
      }
      throw redirect(302, '/account/household-details?error=update_failed');
    }
  },
};
