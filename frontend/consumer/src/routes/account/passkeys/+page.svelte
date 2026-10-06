<script lang="ts">
  import { enhance } from '$app/forms';
  import { PageContainer, Button, Alert, Link } from '@dinnerdonebetter/ui';
  import type { Passkey } from '@primandproper/platform-client/passkeys/v1';
  import { parseRegistrationOptions, serializeRegistration } from '@primandproper/platform-client/webauthn';

  let { data } = $props();
  const passkeys = $derived((data?.passkeys ?? []) as Passkey[]);
  const error = $derived(data?.error as string | null | undefined);
  const deleted = $derived(data?.deleted ?? false);

  const errorMessages: Record<string, string> = {
    invalid: 'Invalid request.',
    delete_failed: 'Failed to remove passkey. Please try again.',
    last_passkey: 'This passkey is your only way to sign in. Add a password or another passkey before removing it.',
    server: 'Something went wrong. Please try again.',
  };
  const displayError = $derived(error ? (errorMessages[error] ?? 'Something went wrong.') : null);

  function formatDate(d: Date | undefined): string {
    if (!d) return '';
    return new Date(d).toLocaleDateString(undefined, {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
    });
  }

  // friendlyName is what the list calls a new passkey: the browser and platform it was made on.
  function friendlyName(): string {
    const ua = navigator.userAgent;
    const browserName = /Edg\//.test(ua)
      ? 'Edge'
      : /Firefox\//.test(ua)
        ? 'Firefox'
        : /Chrome\//.test(ua)
          ? 'Chrome'
          : /Safari\//.test(ua)
            ? 'Safari'
            : 'Browser';
    const platform = /iPhone|iPad/.test(ua)
      ? 'iOS'
      : /Android/.test(ua)
        ? 'Android'
        : /Mac/.test(ua)
          ? 'macOS'
          : /Windows/.test(ua)
            ? 'Windows'
            : /Linux/.test(ua)
              ? 'Linux'
              : '';
    return platform ? `${browserName} on ${platform}` : browserName;
  }

  async function addPasskey() {
    const btn = document.getElementById('add-passkey-btn');
    if (!btn || !window.PublicKeyCredential) return;
    btn.setAttribute('disabled', 'true');

    try {
      const optsRes = await fetch('/auth/passkey/registration/options', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({}),
        credentials: 'include',
      });
      if (!optsRes.ok) throw new Error('Failed to get options');
      const publicKey = parseRegistrationOptions(new Uint8Array(await optsRes.arrayBuffer()));

      const cred = await navigator.credentials.create({ publicKey });
      if (!cred) throw new Error('No credential');
      const attestationResponse = new TextDecoder().decode(serializeRegistration(cred as PublicKeyCredential));

      const verifyRes = await fetch('/auth/passkey/registration/verify', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          attestationResponse,
          friendlyName: friendlyName(),
        }),
        credentials: 'include',
      });
      if (!verifyRes.ok) throw new Error((await verifyRes.json().catch(() => ({}))).error ?? 'Registration failed');
      window.location.reload();
    } catch (err) {
      btn.removeAttribute('disabled');
      alert(err instanceof Error ? err.message : 'Passkey registration failed');
    }
  }
</script>

<PageContainer>
  <h1>Passkeys</h1>
  <p><Link href="/account/settings" class="back-link">Back to Account Settings</Link></p>

  {#if deleted}
    <Alert variant="info">Passkey removed successfully.</Alert>
  {/if}
  {#if displayError}
    <Alert variant="error">{displayError}</Alert>
  {/if}

  <div class="passkeys-content">
    <div class="add-section">
      <h2>Add passkey</h2>
      <p class="muted">Add a passkey to sign in quickly without a password.</p>
      <Button id="add-passkey-btn" type="button" variant="default" onclick={addPasskey}>Add passkey</Button>
    </div>

    {#if passkeys.length > 0}
      <div class="list-section">
        <h2>Your passkeys</h2>
        <div class="passkey-list">
          {#each passkeys as pk (pk.id)}
            <div class="passkey-card">
              <div class="passkey-info">
                <span class="passkey-name">{pk.friendlyName || 'Passkey'}</span>
                <span class="passkey-details">
                  {formatDate(pk.createdAt)}
                  {#if pk.lastUsedAt}
                    · Last used {formatDate(pk.lastUsedAt)}
                  {/if}
                </span>
              </div>
              <form method="POST" action="?/delete" use:enhance class="delete-form">
                <input type="hidden" name="credential_id" value={pk.id} data-testid="passkey-credential-id" />
                <Button type="submit" variant="default" class="remove-btn">Remove</Button>
              </form>
            </div>
          {/each}
        </div>
      </div>
    {:else}
      <div class="empty-section">
        <p class="muted">No passkeys yet. Add one to sign in quickly without a password.</p>
      </div>
    {/if}
  </div>
</PageContainer>

<style>
  .back-link {
    font-size: 0.875rem;
  }
  .passkeys-content {
    display: flex;
    flex-direction: column;
    gap: var(--space-xl);
    margin-top: var(--space-lg);
  }
  .add-section .muted,
  .empty-section .muted {
    font-size: 0.875rem;
    color: var(--color-muted, #666);
    margin: 0 0 var(--space-md);
  }
  .list-section h2,
  .add-section h2 {
    font-size: 1.125rem;
    font-weight: var(--font-weight-medium);
    margin: 0 0 var(--space-sm);
  }
  .passkey-list {
    display: flex;
    flex-direction: column;
    gap: var(--space-sm);
  }
  .passkey-card {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-md);
    padding: var(--space-md);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-md);
    background: var(--color-surface);
  }
  .passkey-info {
    flex: 1;
    min-width: 0;
  }
  .passkey-name {
    display: block;
    font-weight: var(--font-weight-medium);
  }
  .passkey-details {
    font-size: 0.875rem;
    color: var(--color-muted, #666);
    margin-top: 0.25rem;
  }
  .delete-form {
    flex-shrink: 0;
  }
  .remove-btn {
    font-size: 0.875rem;
    padding: 0.25rem 0.5rem;
    color: var(--color-error, #c00);
    border-color: var(--color-error, #c00);
  }
</style>
