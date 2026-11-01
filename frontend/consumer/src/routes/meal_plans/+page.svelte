<script lang="ts">
  import { browser } from '$app/environment';
  import { PageContainer } from '@primandproper/platform-ui';
  import { COMPANY_NAME, APP_STORE_URL } from '@dinnerdonebetter/branding';

  let isIOS = $state(false);
  if (browser) {
    const ua = navigator.userAgent.toLowerCase();
    isIOS = ua.includes('iphone') || ua.includes('ipad');
  }
</script>

<PageContainer>
  <h1>View Meal Plan</h1>
  {#if isIOS}
    <p>
      To view this meal plan, open the {COMPANY_NAME} app. If you don't have the app, download it from the App Store.
    </p>
    <!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- external App Store URL, not an internal route -->
    <a href={APP_STORE_URL} class="app-link">Open in App Store</a>
  {:else}
    <p>
      This meal plan link is for the {COMPANY_NAME} mobile app. Open this link on your iPhone or iPad to view the meal plan.
    </p>
    <!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- external App Store URL, not an internal route -->
    <a href={APP_STORE_URL} class="app-link">Get the App</a>
  {/if}
</PageContainer>

<style>
  .app-link {
    display: inline-block;
    margin-top: var(--pp-space-md);
    padding: var(--pp-space-sm) var(--pp-space-lg);
    background: var(--pp-color-text);
    color: var(--pp-color-surface);
    text-decoration: none;
    border-radius: var(--pp-radius-sm);
    font-weight: var(--pp-font-weight-medium);
  }
  .app-link:hover {
    opacity: 0.9;
  }
</style>
