<script lang="ts">
  import { enhance } from '$app/forms';
  import { PageContainer, FormField, Input, Button, Alert, Link } from '@dinnerdonebetter/ui';
  import {
    InvitationStatus,
    type Account,
    type Invitation,
    type MembershipWithUser,
  } from '@dinnerdonebetter/api-client/primandproper/platform/identity/v1/identity';

  let { data } = $props();
  const account = $derived(data?.account as Account | null | undefined);
  // The roster is a read of its own rather than a field of the account.
  const members = $derived((data?.members ?? []) as MembershipWithUser[]);
  const invitations = $derived((data?.invitations ?? []) as Invitation[]);
  const currentUserId = $derived(data?.currentUserId ?? '');
  const isAdmin = $derived(data?.isAdmin ?? false);
  const error = $derived(data?.error as string | null | undefined);
  const invited = $derived(data?.invited ?? false);

  const errorMessages: Record<string, string> = {
    invalid: 'Invalid input. Please check your entries.',
    invalid_email: 'Please enter a valid email address.',
    invalid_role: 'Invalid role selected.',
    invitation_failed: 'Failed to send invitation. Please try again.',
    cancel_failed: 'Failed to cancel invitation.',
    role_update_failed: 'Failed to update member role.',
    server: 'Something went wrong. Please try again.',
  };
  const displayError = $derived(error ? (errorMessages[error] ?? 'Something went wrong.') : null);

  // DisplayName is never empty on a read — a row whose column is blank reads its handle
  // back — so this is one field rather than a fallback chain that could produce a blank.
  function memberDisplayName(m: MembershipWithUser): string {
    return m.user?.displayName || m.user?.username || 'Unknown User';
  }

  // A membership carries a set of roles rather than one, so "is this person an admin" is
  // a membership question rather than a string comparison.
  function isAccountAdmin(m: MembershipWithUser): boolean {
    return (m.membership?.roles ?? []).includes('account_admin');
  }
</script>

<PageContainer>
  <h1>Household Members</h1>
  <p class="back-link"><Link href="/account/settings">Back to Account Settings</Link></p>

  {#if invited}
    <Alert variant="info">Invitation sent successfully.</Alert>
  {/if}
  {#if displayError}
    <Alert variant="error">{displayError}</Alert>
  {/if}

  {#if !account}
    <p>No household found. Create an account to get started.</p>
  {:else}
    <div class="sections">
      <section>
        <h2>Household Members</h2>
        {#if !members.length}
          <p class="muted">No members yet. Invite someone to join your household.</p>
        {:else}
          <div class="member-list">
            {#each members as m (m.user?.id ?? m.membership?.id ?? 'member')}
              {@const isYou = m.user?.id === currentUserId}
              {@const roleLabel = isAccountAdmin(m) ? 'Admin' : 'Member'}
              <div class="member-card">
                <div class="member-info">
                  <span class="member-name" title={memberDisplayName(m)}>{memberDisplayName(m)}</span>
                  {#if isYou}
                    <span class="member-you">(You)</span>
                  {/if}
                </div>
                {#if isAdmin && !isYou && m.user?.id}
                  <form method="POST" action="?/update-role" use:enhance class="role-form">
                    <input type="hidden" name="user_id" value={m.user.id} data-testid="member-user-id" />
                    <div class="role-form-row">
                      <select name="new_role" class="role-select" data-testid="member-role">
                        <option value="account_member" selected={!isAccountAdmin(m)}> Member </option>
                        <option value="account_admin" selected={isAccountAdmin(m)}> Admin </option>
                      </select>
                      <span class="reason-input">
                        <Input
                          name="reason"
                          type="text"
                          placeholder="Reason (required)"
                          required
                          dataTestId="member-reason"
                        />
                      </span>
                      <Button type="submit" variant="default">Update</Button>
                    </div>
                  </form>
                {:else}
                  <span class="role-badge">{roleLabel}</span>
                {/if}
              </div>
            {/each}
          </div>
        {/if}
      </section>

      {#if isAdmin}
        <section>
          <h2>Add Someone to Your Household</h2>
          <p class="muted">Send an invitation by email. They can join once they have an account.</p>
          <form method="POST" action="?/send-invitation" use:enhance class="invite-form">
            <FormField id="email" label="Email Address" required>
              <Input id="email" name="email" type="email" required dataTestId="invite-email" />
            </FormField>
            <FormField id="name" label="Name (Optional)">
              <Input id="name" name="name" type="text" dataTestId="invite-name" />
            </FormField>
            <FormField id="note" label="Note (Optional)">
              <Input id="note" name="note" type="text" dataTestId="invite-note" />
            </FormField>
            <Button type="submit">Send Invitation</Button>
          </form>
        </section>
      {/if}

      {#if invitations.length > 0}
        <section>
          <h2>Invitations</h2>
          <p class="muted">Invitations you've sent for this household and their status.</p>
          <div class="invitation-list">
            {#each invitations as inv (inv.id)}
              {@const pending = inv.status === InvitationStatus.INVITATION_STATUS_PENDING}
              {@const status = pending ? 'pending' : 'answered'}
              <div class="invitation-card">
                <div class="invitation-info">
                  <span class="invitation-email">{inv.toEmail}</span>
                  {#if inv.toName}
                    <span class="invitation-name">{inv.toName}</span>
                  {/if}
                  <span class="status-badge">{status}</span>
                </div>
                <!--
                  The "Copy Link" button is gone, and its absence is the point. An
                  invitation's token is the secret half of a bearer credential for joining
                  somebody else's household, and no read returns one — the sender mints it,
                  the store keeps a digest, and the only copy goes to the address the
                  invitation was sent to. A sender who could copy the link could join the
                  household as the person they invited.
                -->
                {#if pending}
                  <div class="invitation-actions">
                    {#if isAdmin}
                      <form method="POST" action="?/cancel-invitation" use:enhance class="inline-form">
                        <input type="hidden" name="invitation_id" value={inv.id} data-testid="cancel-invitation-id" />
                        <span class="cancel-btn"><Button type="submit" variant="default">Cancel</Button></span>
                      </form>
                    {/if}
                  </div>
                {/if}
              </div>
            {/each}
          </div>
        </section>
      {/if}
    </div>
  {/if}
</PageContainer>

<style>
  .back-link :global(a) {
    font-size: 0.875rem;
  }
  .sections {
    display: flex;
    flex-direction: column;
    gap: var(--space-xl);
    margin-top: var(--space-lg);
  }
  .sections section h2 {
    font-size: 1.125rem;
    font-weight: var(--font-weight-medium);
    margin: 0 0 var(--space-sm);
  }
  .muted {
    font-size: 0.875rem;
    color: var(--color-muted, #666);
    margin: 0 0 var(--space-md);
  }
  .member-list,
  .invitation-list {
    display: flex;
    flex-direction: column;
    gap: var(--space-sm);
  }
  .member-card,
  .invitation-card {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-md);
    padding: var(--space-md);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-md);
    background: var(--color-surface);
  }
  .member-card {
    flex-wrap: wrap;
    align-items: flex-start;
  }
  .member-info {
    flex: 1 1 0;
    min-width: 0;
  }
  .member-card .role-badge {
    flex: 0 0 auto;
  }
  .member-card .role-form {
    flex: 1 1 100%;
    min-width: 0;
  }
  .member-name {
    font-weight: var(--font-weight-medium);
    display: inline-block;
    max-width: 100%;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .member-you {
    font-size: 0.75rem;
    color: var(--color-muted, #666);
    margin-left: 0.25rem;
  }
  .role-badge {
    font-size: 0.875rem;
    padding: 0.25rem 0.5rem;
    border-radius: var(--radius-sm);
    background: var(--color-surface-muted, #eee);
  }
  .role-form {
    display: flex;
    flex-direction: column;
    gap: var(--space-sm);
    align-items: flex-start;
    width: 100%;
    max-width: 24rem;
  }
  .role-form-row {
    display: flex;
    align-items: center;
    gap: var(--space-sm);
    flex-wrap: wrap;
    width: 100%;
  }
  .role-select {
    font-size: 0.875rem;
    padding: 0.35rem 0.5rem;
    border: 1px solid var(--color-border);
    border-radius: var(--radius-sm);
    flex-shrink: 0;
  }
  .reason-input {
    min-width: 12rem;
    flex: 1;
  }
  .invite-form {
    display: flex;
    flex-direction: column;
    gap: var(--space-md);
    max-width: 24rem;
  }
  .invitation-info {
    flex: 1;
    min-width: 0;
  }
  .invitation-email {
    display: block;
    font-weight: var(--font-weight-medium);
  }
  .invitation-name {
    display: block;
    font-size: 0.875rem;
    color: var(--color-muted, #666);
  }
  .status-badge {
    display: inline-block;
    margin-top: 0.25rem;
    font-size: 0.75rem;
    padding: 0.25rem 0.5rem;
    border-radius: var(--radius-sm);
    background: var(--color-surface-muted, #eee);
  }
  .invitation-actions {
    display: flex;
    gap: var(--space-sm);
  }
  .inline-form {
    display: inline;
  }
  .cancel-btn :global(button) {
    font-size: 0.875rem;
    padding: 0.25rem 0.5rem;
  }
</style>
