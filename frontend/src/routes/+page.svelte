<script lang="ts">
  import { DappConnectService, KMDService } from "$lib/wails-bindings/duckysigner/services";
  import { onMount } from "svelte";

  const walletsList = KMDService.ListWallets();
  let dappConnectOn = false;

  onMount(async () => {
    await startServer()
  });

  async function startServer() {
    dappConnectOn = await DappConnectService.Start()
  }

  async function stopServer() {
    dappConnectOn = await DappConnectService.Stop();
  }
</script>

<div class="grid content-center h-full">
  <a class="btn btn-secondary mb-2" href="/wallets/import">Import wallet</a>
  <a class="btn btn-primary mb-2" href="/wallets/create">Create wallet</a>
  {#if dappConnectOn}
    <button class="btn" on:click={stopServer}>⏹️ Turn off dApp connect server</button>
  {:else}
    <button class="btn" on:click={startServer}>▶️ Turn on dApp connect server</button>
  {/if}

  {#await walletsList then wallets}
    <ul class="menu menu-lg bg-base-200">
      {#each wallets as wallet}
        <li><a href="/wallets?id={atob(wallet.ID)}" class="no-underline">{atob(wallet.Name)}</a></li>
      {/each}
    </ul>
  {/await}
</div>
