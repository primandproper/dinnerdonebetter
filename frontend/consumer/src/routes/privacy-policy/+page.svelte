<script lang="ts">
  import { PageContainer, Link } from '@primandproper/platform-ui';

  let { data } = $props();
  const html = data?.html ?? '';
</script>

<PageContainer>
  <div class="legal-content">
    <!-- eslint-disable-next-line svelte/no-at-html-tags -- server-rendered legal content -->
    {@html html}
  </div>
  <p><Link href="/">Back to Home</Link></p>
</PageContainer>

<style>
  .legal-content {
    max-width: 65ch;
  }
  .legal-content :global(h1) {
    font-size: 1.5rem;
    margin-bottom: var(--pp-space-md);
  }
  .legal-content :global(h2) {
    font-size: 1.125rem;
    margin-top: var(--pp-space-lg);
    margin-bottom: var(--pp-space-sm);
  }
  .legal-content :global(p) {
    margin-bottom: var(--pp-space-md);
  }
  .legal-content :global(ul) {
    margin-bottom: var(--pp-space-md);
    padding-left: 1.5rem;
  }
</style>
