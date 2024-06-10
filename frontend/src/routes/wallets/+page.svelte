<script lang="ts">
  import { page } from '$app/stores';
  import type { Metadata } from '$lib/wails-bindings/duckysigner/kmd/wallet';
  import { KMDService } from '$lib/wails-bindings/duckysigner/services';
  import { Dialog } from "bits-ui";
  import { onMount } from 'svelte';

  let walletId = $page.url.searchParams.get('id') ?? '';
  let walletInfo: Metadata;
  // TODO: Find a more secure way to deal with the wallet password that allows the password to be temporarily stored for a short period of time
  let walletPassword = '';
  let dialogOpen = false;
  let accounts: string[] = [];

  KMDService.GetWalletInfo(walletId).then(info => walletInfo = info);
  KMDService.ListAccountsInWallet(walletId).then(accts => accounts = accts);

  function unlockWallet() {
    console.log(walletPassword);
    dialogOpen = false;
  }

</script>

{#if walletInfo}
  <h1 class="text-center text-4xl mb-8">{atob(walletInfo.Name)}</h1>
  <p>ID: {atob(walletInfo.ID)}</p>

  <div>
    <!-- TODO: Unlock wallet -->
    <button type='submit' class="btn btn-primary" disabled>Unlock</button>
    <!-- TODO: Rename wallet -->
    <button type='submit' class="btn btn-primary" disabled>Rename</button>
    <!-- TODO: Add account -->
    <button type='submit' class="btn btn-primary">Add account</button>
  </div>

  {#if accounts.length > 0}
    <ul class="menu menu-lg bg-base-200">
      {#each accounts as account}
        <!-- TODO: List accounts -->
        <li><a href="#foo" class="no-underline">{'foo'}</a></li>
      {/each}
    </ul>
  {:else}
    <p class="text-center italic">No accounts</p>
  {/if}

  <Dialog.Root bind:open={dialogOpen}>
    <Dialog.Portal>
      <Dialog.Overlay />
      <Dialog.Content class="modal prose modal-open">
        <div class="modal-box">
          <Dialog.Title class="mt-0">Unlock Wallet</Dialog.Title>
          <Dialog.Description></Dialog.Description>
          <form id="unlock-wallet-form" on:submit|preventDefault={unlockWallet} autocomplete="off">
            <div>
              <label class="label" for="wallet-password-input">Wallet password</label>
              <input type="password" bind:value={walletPassword} class="input input-bordered w-full" id="wallet-password-input" required />
            </div>

            <div class="modal-action">
              <button type='submit' class="btn btn-primary">Unlock wallet</button>
            </div>
          </form>
        </div>
      </Dialog.Content>
    </Dialog.Portal>
  </Dialog.Root>
{/if}
