<script lang="ts">
  import { enhance } from '$app/forms';
  import { PageContainer, FormField, Input, Button, Alert, Link } from '@primandproper/platform-ui';

  let { form } = $props();
  const success = form?.success ?? false;
  const error = form?.error;
</script>

<PageContainer narrow>
  <h1>Forgot Password</h1>
  <p>Enter your email address and we'll send you a link to reset your password.</p>

  {#if success}
    <Alert variant="info"
      >If an account exists with that email, we've sent a password reset link. Check your inbox.</Alert
    >
    <p><Link href="/forgot_password">Try again</Link> with a different email.</p>
  {:else}
    <form method="POST" use:enhance class="forgot-form">
      {#if error}
        <Alert variant="error">{error}</Alert>
      {/if}
      <FormField id="email" label="Email Address" required>
        <Input id="email" name="email" type="email" autocomplete="email" required dataTestId="forgot-password-email" />
      </FormField>
      <Button type="submit">Send Reset Link</Button>
    </form>
  {/if}

  <p><Link href="/login">Back to Sign In</Link></p>
</PageContainer>

<style>
  .forgot-form {
    display: flex;
    flex-direction: column;
    gap: var(--pp-space-md);
    margin: var(--pp-space-lg) 0;
  }
</style>
