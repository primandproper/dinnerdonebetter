<script lang="ts">
  import { Heading, Text, Link, Card } from '@dinnerdonebetter/ui';

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
            <td><code>{(setting as Record<string, unknown>).id ?? '-'}</code></td>
            <td>{(setting as Record<string, unknown>).name ?? '-'}</td>
            <td><Link href="/settings/{(setting as Record<string, unknown>).id}">View</Link></td>
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
    color: var(--color-text-muted);
    margin-bottom: var(--space-lg);
  }
  .error {
    color: var(--color-error, #b91c1c);
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
  .search-form {
    display: flex;
    align-items: center;
    gap: var(--space-sm);
    margin-bottom: var(--space-lg);
  }
  .search-form input {
    padding: var(--space-xs) var(--space-sm);
    min-width: 12rem;
  }
</style>
