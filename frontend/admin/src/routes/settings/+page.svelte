<script lang="ts">
  import { Heading, Text, Link, Card } from '@primandproper/platform-ui';

  let { data } = $props();
</script>

<Heading level={1}>Settings</Heading>
<p class="subtitle">Service settings</p>

<form method="get" action="/settings" class="search-form">
  <label for="q">Search</label>
  <input id="q" name="q" type="search" placeholder="Filter by name…" value={data?.query ?? ''} />
  <button type="submit">Search</button>
</form>

{#if data?.error}
  <p class="error">{data.error}</p>
{:else if data?.settings && data.settings.length > 0}
  <div class="table-wrap">
    <table class="data-table">
      <thead>
        <tr>
          <th>ID</th>
          <th>Name</th>
          <th></th>
        </tr>
      </thead>
      <tbody>
        {#each data.settings as setting (setting.id)}
          <tr>
            <td><code>{setting.id || '-'}</code></td>
            <td>{setting.name || '-'}</td>
            <td><Link href="/settings/{setting.id}">View</Link></td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
{:else}
  <Card>
    <Text>No settings found.</Text>
  </Card>
{/if}

<style>
  .subtitle {
    color: var(--pp-color-text-muted);
    margin-bottom: var(--pp-space-lg);
  }
  .error {
    color: var(--pp-color-error, #b91c1c);
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
    padding: var(--pp-space-sm) var(--pp-space-md);
    text-align: left;
    border-bottom: 1px solid var(--pp-color-border);
  }
  .search-form {
    display: flex;
    align-items: center;
    gap: var(--pp-space-sm);
    margin-bottom: var(--pp-space-lg);
  }
  .search-form input {
    padding: var(--pp-space-xs) var(--pp-space-sm);
    min-width: 12rem;
  }
</style>
