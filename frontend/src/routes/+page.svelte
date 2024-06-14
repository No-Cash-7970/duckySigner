<script lang="ts">
  import { KMDService } from "$lib/wails-bindings/duckysigner/services";

  const walletsList = KMDService.ListWallets();
</script>

<div class="grid content-center h-full">
  <a class="btn btn-secondary mb-2" href="/wallets/import">Import wallet</a>
  <a class="btn btn-primary mb-2" href="/wallets/create">Create wallet</a>

  {#await walletsList then wallets}
    <ul class="menu menu-lg bg-base-200">
      {#each wallets as wallet}
        <li><a href="/wallets?id={atob(wallet.ID)}" class="no-underline">{atob(wallet.Name)}</a></li>
      {/each}
    </ul>
  {/await}
</div>
