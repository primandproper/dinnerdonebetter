<script lang="ts">
  import { enhance } from '$app/forms';
  import { PageContainer, FormField, Input, Button, Alert, Link } from '@dinnerdonebetter/ui';

  let { data, form } = $props();
  const token = data?.token ?? form?.token ?? '';
  const missingToken = data?.missingToken ?? false;
  const deadLink = $derived(data?.deadLink ?? false);
  const error = form?.error;
</script>

<PageContainer narrow>
  <h1>Reset Password</h1>

  {#if missingToken && !token}
    <Alert variant="error">Missing reset token. Please use the link from your email.</Alert>
    <p><Link href="/forgot_password">Request a new reset link</Link></p>
  {:else if deadLink}
    <Alert variant="error">This reset link is invalid or has expired. Please request a new one.</Alert>
    <p><Link href="/forgot_password">Request a new reset link</Link></p>
  {:else}
    <div class="reset-form">
      {#if error}
        <Alert variant="error">{error}</Alert>
      {/if}
      <form method="POST" use:enhance>
        <input type="hidden" name="token" value={token} data-testid="reset-password-token" />
        <FormField id="new_password" label="New Password" required>
          <Input
            id="new_password"
            name="new_password"
            type="password"
            autocomplete="new-password"
            required
            minlength={8}
            dataTestId="reset-password-new"
          />
        </FormField>
        <FormField id="confirm_password" label="Confirm Password" required>
          <Input
            id="confirm_password"
            name="confirm_password"
            type="password"
            autocomplete="new-password"
            required
            minlength={8}
            dataTestId="reset-password-confirm"
          />
        </FormField>
        <Button type="submit">Reset Password</Button>
      </form>
    </div>
    <p><Link href="/login">Back to Sign In</Link></p>
  {/if}
</PageContainer>

<style>
  .reset-form {
    margin: var(--space-lg) 0;
  }
  .reset-form form {
    display: flex;
    flex-direction: column;
    gap: var(--space-md);
  }
</style>
