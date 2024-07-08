<script lang="ts">
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import { KMDService } from '$lib/wails-bindings/duckysigner/services';
  import { Dialog } from "bits-ui";
  import { onMount } from 'svelte';

  let walletId = '';
  let acctAddr = '';
  let backLink = '/';

  onMount(async () => {
    walletId = $page.url.searchParams.get('id') ?? '';
    acctAddr = $page.url.searchParams.get('addr') ?? '';
    backLink = (walletId && acctAddr) ? `/account?id=${walletId}&addr=${acctAddr}` : '/';
  });
</script>

<a href="{backLink}" class="btn">Back</a>

<h1 class="text-center text-4xl mb-8">Sign Transaction</h1>

<p>Wallet: {walletId}</p>
<p>Address: {acctAddr}</p>
