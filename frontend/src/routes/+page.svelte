<script lang="ts">
  import { KMDService } from "$lib/wails-bindings/duckysigner/services";

  const walletsList = KMDService.ListWallets();
</script>

<div class="grid content-center h-[100vh]">
  <a class="btn btn-sm btn-primary" href="/wallets/create">Create wallet</a>

  {#await walletsList then wallets}
    <ul class="menu menu-lg bg-base-200">
      {#each wallets as wallet}
        <li><a href="/wallets?id={atob(wallet.ID)}" class="no-underline">{atob(wallet.Name)}</a></li>
      {/each}
    </ul>
  {/await}
</div>
