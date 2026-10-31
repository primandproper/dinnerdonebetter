<script lang="ts">
  import { PageContainer, Alert, Link } from '@primandproper/platform-ui';

  let { data } = $props();
  const success = data?.success ?? false;
  const message = data?.message ?? 'Verification complete.';
</script>

<PageContainer narrow>
  <h1>Verify Email</h1>
  {#if success}
    <Alert variant="info">{message}</Alert>
  {:else}
    <Alert variant="error">{message}</Alert>
  {/if}
  <p><Link href="/login">Sign In</Link></p>
</PageContainer>
