<script lang="ts">
  import { PageContainer, Link, Alert } from '@primandproper/platform-ui';
  import { COMPANY_NAME } from '@dinnerdonebetter/branding';

  let { data } = $props();
</script>

<PageContainer>
  <h1>Welcome to {COMPANY_NAME}</h1>
  <p>Manage your meal plans and grocery lists.</p>

  {#if data?.flash}
    <Alert variant={data.flash.type === 'error' ? 'error' : 'info'}>{data.flash.message}</Alert>
  {/if}

  {#if data?.user}
    <p>Hello, {data.user.username ?? data.user.firstName ?? 'there'}!</p>
    {#if data.account}
      <p>Account: {data.account.name}</p>
    {/if}
    <p><Link href="/logout">Sign out</Link></p>
  {:else}
    <p><Link href="/login">Sign in</Link></p>
  {/if}
</PageContainer>
