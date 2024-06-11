<script lang="ts">
  import { page } from '$app/stores';
  import type { Metadata } from '$lib/wails-bindings/duckysigner/kmd/wallet';
  import { KMDService } from '$lib/wails-bindings/duckysigner/services';
  import { Dialog } from "bits-ui";
  import { onMount } from 'svelte';

  let walletId = '';
  let walletInfo: Metadata;
  // TODO: Find a more secure way to deal with the wallet password that allows the password to be temporarily stored for a short period of time
  let walletPassword = '';
  let passwordWrong = false;
  let passwordCorrect = false;
  let dialogOpen = true;
  let accounts: string[] = [];

  onMount(async () => {
    walletId = $page.url.searchParams.get('id') ?? '';
    walletInfo = await KMDService.GetWalletInfo(walletId);
    accounts = await KMDService.ListAccountsInWallet(walletId);
  });

  async function unlockWallet() {
    try {
      await KMDService.CheckWalletPassword(walletId, walletPassword);
      passwordCorrect = true;
      passwordWrong = false;
      dialogOpen = false;
    } catch (error) {
      passwordWrong = true;
      passwordCorrect = false;
    }
  }

  async function addAccount() {
    await KMDService.GenerateWalletAccount(walletId, walletPassword);
    accounts = await KMDService.ListAccountsInWallet(walletId);
  }
</script>

<a href="/" class="btn">Back</a>

{#if walletInfo}
  <h1 class="text-center text-4xl mb-8">{atob(walletInfo.Name)}</h1>

  <p>ID: {atob(walletInfo.ID)}</p>

  {#if passwordCorrect}
    <div>
      <!-- TODO: Rename wallet -->
      <button type='submit' class="btn btn-primary" disabled>Rename</button>
      <button type='submit' class="btn btn-primary" on:click={addAccount}>Add account</button>
    </div>
    {#if accounts.length > 0}
      <ul class="menu menu-lg bg-base-200">
        {#each accounts as address}
          <li><a href="/account?id={walletId}&addr={address}" class="no-underline">{address}</a></li>
        {/each}
      </ul>
    {:else}
      <p class="text-center italic">No accounts</p>
    {/if}
  {/if}
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
            {#if passwordWrong}
              <div class="label bg-error px-2">
                <span class="label-text-alt text-error-content">Incorrect password.</span>
              </div>
            {/if}
          </div>
          <div class="modal-action">
            <button type='submit' class="btn btn-primary">Unlock wallet</button>
          </div>
        </form>
      </div>
    </Dialog.Content>
  </Dialog.Portal>
</Dialog.Root>
