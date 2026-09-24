<script lang="ts">
  import { Heading, Text, Link, Card } from '@dinnerdonebetter/ui';

  let { data } = $props();
  const account = data?.account as Record<string, unknown> | null;
</script>

{#if data?.error}
  <p class="error">{data.error}</p>
{:else if account}
  <Heading level={1}>Account details</Heading>
  <p class="subtitle">ID: {account.id ?? '-'}</p>

  <Card>
    <h2 class="card-title">Profile</h2>
    <dl class="detail-list">
      <dt>Name</dt>
      <dd>{account.name ?? '-'}</dd>
      <dt>City</dt>
      <dd>{account.city ?? '-'}</dd>
      <dt>State</dt>
      <dd>{account.state ?? '-'}</dd>
    </dl>
  </Card>

  {#if data?.users && (data.users as unknown[]).length > 0}
    <Card>
      <h2 class="card-title">Users</h2>
      <ul>
        {#each data.users as u ((u as Record<string, unknown>).id)}
          <li>
            <Link href="/users/{(u as Record<string, unknown>).id}">
              {(u as Record<string, unknown>).username ?? (u as Record<string, unknown>).id}
            </Link>
          </li>
        {/each}
      </ul>
    </Card>
  {/if}

  {#if data?.auditLog && (data.auditLog as unknown[]).length > 0}
    <Card>
      <h2 class="card-title">Audit log</h2>
      <div class="table-wrap">
        <table class="data-table">
          <thead>
            <tr>
              <th>Time</th>
              <th>Event</th>
            </tr>
          </thead>
          <tbody>
            {#each data.auditLog as entry, i ((entry as Record<string, unknown>).id ?? i)}
              <tr>
                <td>{(entry as Record<string, unknown>).recordedAt ?? '-'}</td>
                <td>{(entry as Record<string, unknown>).eventType ?? '-'}</td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    </Card>
  {/if}

  <p><Link href="/accounts">Back to accounts</Link></p>
{:else}
  <Text>Account not found.</Text>
{/if}

<style>
  .subtitle {
    color: var(--color-text-muted);
    margin-bottom: var(--space-lg);
  }
  .error {
    color: var(--color-error, #b91c1c);
  }
  .card-title {
    font-size: 1rem;
    margin-bottom: var(--space-md);
  }
  .detail-list {
    display: grid;
    grid-template-columns: auto 1fr;
    gap: var(--space-xs) var(--space-lg);
  }
  .detail-list dt {
    color: var(--color-text-muted);
  }
  .table-wrap {
    overflow-x: auto;
  }
  .data-table th,
  .data-table td {
    padding: var(--space-sm) var(--space-md);
    text-align: left;
    border-bottom: 1px solid var(--color-border);
  }
</style>
