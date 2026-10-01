<script lang="ts">
  import { browser } from '$app/environment';
  import { PageContainer, LoginForm, Button, Alert, Link } from '@dinnerdonebetter/ui';

  let { data, form } = $props();
  const supportsPasskey = browser && typeof PublicKeyCredential !== 'undefined';
  const resetSuccess = data?.resetSuccess ?? false;

  async function signInWithPasskey() {
    if (!window.PublicKeyCredential) {
      alert('Passkeys are not supported in this browser.');
      return;
    }
    const usernameInput = document.getElementById('username') as HTMLInputElement;
    const username = usernameInput?.value?.trim() ?? '';

    function b64enc(buf: ArrayBuffer): string {
      const b = new Uint8Array(buf);
      let s = '';
      for (let i = 0; i < b.length; i++) s += String.fromCharCode(b[i]);
      return btoa(s).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
    }
    function b64dec(s: string): ArrayBuffer {
      const padded = s.replace(/-/g, '+').replace(/_/g, '/');
      const padded2 = padded + '==='.slice((padded.length + 3) % 4);
      const binary = atob(padded2);
      return Uint8Array.from(binary, (c) => c.charCodeAt(0)).buffer;
    }

    // One ceremony: options from the server, the key, and the assertion back. A key tapped
    // with no PIN or biometric is one factor, so a person with a second factor is asked for
    // their code and taps again: the first assertion's challenge is spent.
    async function ceremony(totpCode: string): Promise<{ totpRequired?: boolean; redirect?: string }> {
      const optsRes = await fetch('/auth/passkey/authentication/options', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username }),
        credentials: 'include',
      });
      if (!optsRes.ok) throw new Error('Failed to get options');
      const opts = await optsRes.json();

      const obj = JSON.parse(atob(opts.options));
      const pk = obj.publicKey || obj;
      if (typeof pk.challenge === 'string') pk.challenge = b64dec(pk.challenge);
      for (const c of pk.allowCredentials ?? []) {
        if (typeof c.id === 'string') c.id = b64dec(c.id);
      }

      const cred = await navigator.credentials.get({ publicKey: pk });
      if (!cred) throw new Error('No credential');

      const pkCred = cred as PublicKeyCredential;
      const r = pkCred.response as AuthenticatorAssertionResponse;
      const assertion = {
        id: pkCred.id,
        rawId: b64enc(pkCred.rawId),
        type: pkCred.type,
        response: {
          clientDataJSON: b64enc(r.clientDataJSON),
          authenticatorData: b64enc(r.authenticatorData),
          signature: b64enc(r.signature),
          userHandle: r.userHandle ? b64enc(r.userHandle) : null,
        },
      };

      const verifyRes = await fetch('/auth/passkey/authentication/verify', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, assertionResponse: assertion, totpCode }),
        credentials: 'include',
      });
      const result = await verifyRes.json().catch(() => ({}));
      if (result.totpRequired) return { totpRequired: true };
      if (!verifyRes.ok) throw new Error(result.error ?? 'Authentication failed');
      return result;
    }

    try {
      let result = await ceremony('');
      if (result.totpRequired) {
        const code = prompt('Enter the code from your authenticator app, then use your passkey again.')?.trim();
        if (!code) return;
        result = await ceremony(code);
        if (result.totpRequired) throw new Error('That code was not accepted');
      }
      window.location.href = result.redirect ?? '/';
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Passkey sign-in failed');
    }
  }
</script>

<PageContainer narrow>
  <h1>Sign In</h1>
  {#if resetSuccess}
    <Alert variant="info">Your password has been reset. Sign in with your new password.</Alert>
  {/if}
  <LoginForm
    action="?/login"
    username={form?.username ?? ''}
    error={form?.error}
    showTotp={(form as { totpRequired?: boolean } | undefined)?.totpRequired ?? false}
  >
    {#snippet passkeySlot()}
      {#if supportsPasskey}
        <div class="passkey-section">
          <p class="divider">or</p>
          <Button type="button" variant="default" onclick={signInWithPasskey}>Sign in with passkey</Button>
        </div>
      {/if}
    {/snippet}
  </LoginForm>

  <p><Link href="/forgot_password">Forgot password?</Link></p>
</PageContainer>

<style>
  .passkey-section {
    margin-top: var(--space-lg);
  }
  .divider {
    margin: var(--space-md) 0;
    color: var(--color-text-muted);
    font-size: var(--font-size-sm);
  }
</style>
