<script lang="ts">
  import { Heading, Text, Link, Card } from '@dinnerdonebetter/ui';

  let { data } = $props();
  const user = data?.user as Record<string, unknown> | null;
</script>

{#if data?.passwordChangeUpdated}
  <p class="success">Password change requirement updated.</p>
{/if}

{#if data?.error}
  <p class="error">{data.error}</p>
{:else if user}
  <Heading level={1}>User details</Heading>
  <p class="subtitle">ID: {user.id ?? '-'}</p>

  <Card>
    <h2 class="card-title">Profile</h2>
    <dl class="detail-list">
      <dt>Username</dt>
      <dd>{user.username ?? '-'}</dd>
      <dt>First name</dt>
      <dd>{user.firstName ?? '-'}</dd>
      <dt>Last name</dt>
      <dd>{user.lastName ?? '-'}</dd>
      <dt>Email</dt>
      <dd>{user.emailAddress ?? '-'}</dd>
      <dt>Account status</dt>
      <dd>{user.accountStatus ?? '-'}</dd>
      <dt>Requires password change</dt>
      <dd>{user.requiresPasswordChange ? 'Yes' : 'No'}</dd>
    </dl>
  </Card>

  {#if data?.accounts && (data.accounts as unknown[]).length > 0}
    <Card>
      <h2 class="card-title">Accounts</h2>
      <ul>
        {#each data.accounts as acc ((acc as Record<string, unknown>).id)}
          <li>
            <Link href="/accounts/{(acc as Record<string, unknown>).id}">
              {(acc as Record<string, unknown>).name ?? (acc as Record<string, unknown>).id}
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
              <th>Action</th>
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

  <Card>
    <h2 class="card-title">Actions</h2>
    <div class="actions-list">
      <Link href="/users/{user.id}/sessions">View &amp; manage sessions</Link>
      <!--
        The two password-change buttons are gone with the RPC behind them. The flag is
        still readable above, and still honoured at sign-in; what is missing is a way to
        set it, because the directory's gRPC surface exposes three operator writes and
        this is not one of them. See the page's server file.
      -->
    </div>
  </Card>

  <p><Link href="/users">Back to users</Link></p>
{:else}
  <Text>User not found.</Text>
{/if}

<style>
  .subtitle {
    color: var(--color-text-muted);
    margin-bottom: var(--space-lg);
  }
  .error {
    color: var(--color-error, #b91c1c);
  }
  .success {
    color: var(--color-success, #15803d);
    margin-bottom: var(--space-md);
  }
  .actions-list {
    display: flex;
    flex-direction: column;
    gap: var(--space-md);
    align-items: flex-start;
  }
  .btn {
    padding: var(--space-xs) var(--space-md);
    border: none;
    border-radius: 4px;
    cursor: pointer;
    font-size: 0.875rem;
  }
  .btn-warning {
    background: var(--color-warning, #d97706);
    color: white;
  }
  .btn-secondary {
    background: var(--color-secondary, #6b7280);
    color: white;
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
</style>
