<script lang="ts">
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import { KMDService } from '$lib/wails-bindings/duckysigner/services';
  import { Dialog } from "bits-ui";
  import { onMount } from 'svelte';

  let walletId = '';
  let acctAddr = '';
  let backLink = '/';
  let deleteAcctDialogOpen = false;
  let passwordWrong = false;
  let walletPassword = '';

  onMount(() => {
    walletId = $page.url.searchParams.get('id') ?? '';
    acctAddr = $page.url.searchParams.get('addr') ?? '';
    backLink = walletId ? `/wallets?id=${walletId}` : '/';
  });

  async function removeAcct() {
    try {
      await KMDService.CheckWalletPassword(walletId, walletPassword);
      passwordWrong = false;
      deleteAcctDialogOpen = false;
    } catch (error) {
      passwordWrong = true;
    }

    await KMDService.RemoveAccountFromWallet(acctAddr, walletId, walletPassword);
    goto(backLink);
  }
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
  <button class="btn btn-error" on:click={() => deleteAcctDialogOpen = true}>Remove from wallet</button>
  <!-- TODO: Sign txn file -->
  <button class="btn">Sign transaction file</button>
</div>

<Dialog.Root bind:open={deleteAcctDialogOpen}>
  <Dialog.Portal>
    <Dialog.Overlay />
    <Dialog.Content class="modal prose modal-open">
      <div class="modal-box">
        <Dialog.Title class="mt-0">Remove account from wallet</Dialog.Title>
        <Dialog.Description>Enter wallet password to remove this account.</Dialog.Description>
        <form id="unlock-wallet-form" on:submit|preventDefault={removeAcct} autocomplete="off">
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
            <button type='submit' class="btn btn-primary">Remove this account</button>
          </div>
        </form>
      </div>
    </Dialog.Content>
  </Dialog.Portal>
</Dialog.Root>
