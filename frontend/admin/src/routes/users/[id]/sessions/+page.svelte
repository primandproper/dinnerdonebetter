<script lang="ts">
  import { enhance } from '$app/forms';
  import { Heading, Button, Alert, Link, Card } from '@dinnerdonebetter/ui';
  import type { ActiveSignIn } from '@primandproper/platform-client/signin/v1';

  let { data } = $props();
  const signIns = $derived((data?.signIns ?? []) as ActiveSignIn[]);
  const error = $derived(data?.error as string | null | undefined);
  const revoked = $derived(data?.revoked ?? false);
  const revokedAll = $derived(data?.revokedAll ?? false);
  const userId = $derived(data?.userId as string);

  const errorMessages: Record<string, string> = {
    invalid: 'Invalid request.',
    revoke_failed: 'Failed to revoke session. Please try again.',
    revoke_all_failed: 'Failed to revoke sessions. Please try again.',
    server: 'Something went wrong. Please try again.',
  };
  const displayError = $derived(error ? (errorMessages[error] ?? 'Something went wrong.') : null);

  function formatDateTime(d: Date | undefined): string {
    if (!d) return '-';
    return new Date(d).toLocaleDateString(undefined, {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
      hour: 'numeric',
      minute: '2-digit',
    });
  }
</script>

<Heading level={1}>Sessions</Heading>
<p class="subtitle">User ID: {userId}</p>

{#if revoked}
  <Alert variant="info">Session revoked successfully.</Alert>
{/if}
{#if revokedAll}
  <Alert variant="info">All sessions have been revoked.</Alert>
{/if}
{#if displayError}
  <Alert variant="error">{displayError}</Alert>
{/if}

{#if signIns.length > 0}
  <div class="actions-bar">
    <form method="POST" action="?/revoke-all" use:enhance>
      <Button type="submit" variant="default" class="revoke-all-btn">Revoke all sessions</Button>
    </form>
  </div>

  <Card>
    <div class="table-wrap">
      <table class="data-table">
        <thead>
          <tr>
            <th>Login</th>
            <th>Opened by</th>
            <th>Account</th>
            <th>Signed in</th>
            <th>Last refreshed</th>
            <th>Expires</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          {#each signIns as signIn (signIn.familyId)}
            <tr>
              <td>{signIn.credentialKind || '-'}{signIn.administrative ? ' (admin)' : ''}</td>
              <td>{signIn.actorId || '-'}</td>
              <td>{signIn.activeAccountId || '-'}</td>
              <td>{formatDateTime(signIn.signedInAt)}</td>
              <td>{formatDateTime(signIn.lastRefreshedAt)}</td>
              <td>{formatDateTime(signIn.expiresAt)}</td>
              <td>
                <form method="POST" action="?/revoke" use:enhance class="inline-form">
                  <input type="hidden" name="family_id" value={signIn.familyId} />
                  <Button type="submit" variant="default" class="revoke-btn">Revoke</Button>
                </form>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  </Card>
{:else if !displayError}
  <Card><p class="muted">No active sessions found.</p></Card>
{/if}

<p><Link href="/users/{userId}">Back to user</Link></p>

<style>
  .subtitle {
    color: var(--color-text-muted);
    margin-bottom: var(--space-lg);
  }
  .actions-bar {
    margin-bottom: var(--space-md);
  }
  .table-wrap {
    overflow-x: auto;
  }
  .data-table {
    width: 100%;
    border-collapse: collapse;
  }
  .data-table th,
  .data-table td {
    padding: var(--space-sm) var(--space-md);
    text-align: left;
    border-bottom: 1px solid var(--color-border);
  }
  .inline-form {
    display: inline;
  }
  .muted {
    color: var(--color-text-muted);
  }
</style>
