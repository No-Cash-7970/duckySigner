<script lang="ts">
  import { page } from '$app/stores';
  import { onMount } from 'svelte';

  let walletId = '';
  let acctAddr = '';
  let backLink = '/';

  onMount(() => {
    walletId = $page.url.searchParams.get('id') ?? '';
    acctAddr = $page.url.searchParams.get('addr') ?? '';
    backLink = walletId ? `/wallets?id=${walletId}` : '/';
  });
</script>

<a href="{backLink}" class="btn">Back</a>

<h1 class="text-center text-4xl mb-8 truncate">{acctAddr}</h1>

<div class="text-2xl mb-4">
  <!-- TODO: Add wallet amount -->
  ??? Algos
</div>
<div>
  <!-- TODO: See account mnemonic -->
  <button class="btn">See mnemonic</button>
  <!-- TODO: Remove account -->
  <button class="btn">Remove from wallet</button>
  <!-- TODO: Sign txn file -->
  <button class="btn">Sign transaction file</button>
</div>
