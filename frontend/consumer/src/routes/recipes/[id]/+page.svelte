<script lang="ts">
  import { page } from '$app/stores';
  import { Link, PageContainer } from '@primandproper/platform-ui';

  const _id = $derived($page.params.id);
</script>

<PageContainer>
  <h1>Recipe Created</h1>
  <p>Your recipe has been created successfully.</p>
  <p><Link href="/meal_plans">Back to Meal Plans</Link></p>
</PageContainer>
