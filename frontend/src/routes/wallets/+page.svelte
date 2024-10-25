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
  let walletPasswordDialogOpen = true;
  let renameWalletDialogOpen = false;
  let mnemonicDialogOpen = false;
  let importAccountDialogOpen = false;
  let renameFail = false;
  let importFail = false;
  let accounts: string[] = [];
  let mnemonicParts: string[] = [];

  onMount(async () => {
    walletId = $page.url.searchParams.get('id') ?? '';
    walletInfo = await KMDService.GetWalletInfo(walletId);
    accounts = await KMDService.SessionListAccounts()
  });

  async function unlockWallet() {
    try {
      await KMDService.StartSession(walletId, walletPassword);
      passwordCorrect = true;
      passwordWrong = false;
      walletPasswordDialogOpen = false;
    } catch (error) {
      passwordWrong = true;
      passwordCorrect = false;
    }
  }

  async function renameWallet(e: SubmitEvent) {
    const formData = new FormData(e.target as HTMLFormElement);
    const newName = formData.get('newName')?.toString();

    try {
      await KMDService.RenameWallet(walletId, newName ?? '', walletPassword)
      walletInfo = await KMDService.GetWalletInfo(walletId);
      renameWalletDialogOpen = false;
      renameFail = false;
    } catch (error) {
      renameFail = true;
    }
  }

  async function generateAccount() {
    await KMDService.SessionGenerateAccount();
    accounts = await KMDService.SessionListAccounts();
  }

  async function showMnemonic() {
    mnemonicParts = (await KMDService.SessionExportWallet(walletPassword)).split(' ');
    mnemonicDialogOpen = true;
  }

  async function importAccount(e: SubmitEvent) {
    const formData = new FormData(e.target as HTMLFormElement);
    const importedMnemonic = formData.get('mnemonic')?.toString();

    try {
      await KMDService.SessionImportAccount(importedMnemonic  ?? '');
      importFail = false;
      importAccountDialogOpen = false;
      accounts = await KMDService.SessionListAccounts();
    } catch (error: any) {
      importFail = true;
    }

  }

</script>

<a href="/" class="btn">Back</a>

{#if walletInfo}
  <h1 class="text-center text-4xl mb-8">{atob(walletInfo.Name)}</h1>

  {#if passwordCorrect}
    <div>
      <button class="btn" on:click={showMnemonic}>See mnemonic</button>
      <button class="btn" on:click={() => renameWalletDialogOpen = true}>Rename</button>
      <button class="btn btn-primary" on:click={generateAccount}>Generate new account</button>
      <button class="btn btn-primary" on:click={() => importAccountDialogOpen = true}>Import account</button>
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

<Dialog.Root bind:open={walletPasswordDialogOpen}>
  <Dialog.Portal>
    <Dialog.Overlay />
    <Dialog.Content class="modal prose modal-open">
      <div class="modal-box">
        <Dialog.Title class="mt-0">Unlock Wallet</Dialog.Title>
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
            <a href="/" class="btn">Cancel</a>
          </div>
        </form>
      </div>
    </Dialog.Content>
  </Dialog.Portal>
</Dialog.Root>

<Dialog.Root bind:open={renameWalletDialogOpen}>
  <Dialog.Portal>
    <Dialog.Overlay />
    <Dialog.Content class="modal prose modal-open">
      <div class="modal-box">
        <Dialog.Title class="mt-0">Rename Wallet</Dialog.Title>
        <form id="rename-wallet-form" on:submit|preventDefault={renameWallet} autocomplete="off">
          <div>
            <label class="label" for="rename-wallet-input">New wallet name</label>
            <input type="text" name="newName" class="input input-bordered w-full" id="rename-wallet-input" required />
            {#if renameFail}
              <div class="label bg-error px-2">
                <span class="label-text-alt text-error-content">A wallet with this name already exists.</span>
              </div>
            {/if}
          </div>
          <div class="modal-action">
            <button type='submit' class="btn btn-primary">Rename wallet</button>
            <Dialog.Close class="btn">Close</Dialog.Close>
          </div>
        </form>
      </div>
    </Dialog.Content>
  </Dialog.Portal>
</Dialog.Root>

<Dialog.Root bind:open={mnemonicDialogOpen}>
  <Dialog.Portal>
    <Dialog.Overlay />
    <Dialog.Content class="modal prose modal-open">
      <div class="modal-box">
        <Dialog.Title class="mt-0">Account Mnemonic</Dialog.Title>
        <table class="table">
          <tbody>
            {#each mnemonicParts as part, index}
              <tr>
                <td>{index + 1}</td>
                <td>{part}</td>
              </tr>
            {/each}
          </tbody>
        </table>
        <div class="modal-action">
          <Dialog.Close class="btn">Close</Dialog.Close>
        </div>
      </div>
    </Dialog.Content>
  </Dialog.Portal>
</Dialog.Root>

<Dialog.Root bind:open={importAccountDialogOpen}>
  <Dialog.Portal>
    <Dialog.Overlay />
    <Dialog.Content class="modal prose modal-open">
      <div class="modal-box">
        <Dialog.Title class="mt-0">Import Account</Dialog.Title>
        <form id="rename-wallet-form" on:submit|preventDefault={importAccount} autocomplete="off">
          <div>
            <label class="label" for="mnemonic-input">Mnemonic</label>
            <textarea name="mnemonic" class="textarea textarea-bordered w-full" id="mnemonic-input" required></textarea>
          </div>
          {#if importFail}
            <div class="label bg-error px-2">
              <span class="label-text-alt text-error-content">Cannot import account.</span>
            </div>
          {/if}
          <div class="modal-action">
            <button type='submit' class="btn btn-primary">Import</button>
            <Dialog.Close class="btn">Cancel</Dialog.Close>
          </div>
        </form>
      </div>
    </Dialog.Content>
  </Dialog.Portal>
</Dialog.Root>
