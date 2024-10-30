<script lang="ts">
  import { page } from '$app/stores';
  import type { Metadata } from '$lib/wails-bindings/duckysigner/kmd/wallet';
  import { KMDService } from '$lib/wails-bindings/duckysigner/services';
  import { Dialog } from "bits-ui";
  import { onMount } from 'svelte';

  let walletId = '';
  let walletInfo: Metadata;
  let passwordWrong = false;
  let passwordCorrect = false;
  let unlockWalletDialogOpen = false;
  let renameWalletDialogOpen = false;
  let passwordForMnemonicDialogOpen = false;
  let mnemonicDialogOpen = false;
  let importAccountDialogOpen = false;
  let renameFail = false;
  let importFail = false;
  let accounts: string[] = [];
  let mnemonicParts: string[] = [];
  let sessionStarted = false;

  onMount(async () => {
    walletId = $page.url.searchParams.get('id') ?? '';
    walletInfo = await KMDService.GetWalletInfo(walletId);
    sessionStarted = await KMDService.SessionIsForWallet(walletId);

    if (!sessionStarted) {
      // There is no session. Ask for wallet password
      unlockWalletDialogOpen = true;
      return
    }

    try {
      await KMDService.SessionCheck();
    } catch (error) {
      // Session has expired. Ask for wallet password
      unlockWalletDialogOpen = true;
      return
    }

    // Getting to this point means we have a valid wallet session, so load the wallet's accounts
    accounts = await KMDService.SessionListAccounts();
  });

  async function unlockWallet(e: SubmitEvent) {
    const formData = new FormData(e.target as HTMLFormElement);
    try {
      await KMDService.StartSession(walletId, formData.get('walletPassword')?.toString() ?? '');
      sessionStarted = true;
      accounts = await KMDService.SessionListAccounts();
      passwordCorrect = true;
      passwordWrong = false;
      unlockWalletDialogOpen = false;
    } catch (error) {
      passwordWrong = true;
      passwordCorrect = false;
      sessionStarted = false;
    }
  }

  async function renameWallet(e: SubmitEvent) {
    const formData = new FormData(e.target as HTMLFormElement);
    const newName = formData.get('newName')?.toString() ?? '';
    const walletPassword = formData.get('walletPassword')?.toString() ?? '';

    try {
      await KMDService.RenameWallet(walletId, newName, walletPassword)
      walletInfo = await KMDService.GetWalletInfo(walletId);
      renameWalletDialogOpen = false;
      renameFail = false;
      passwordWrong = false;
    } catch (e: unknown) {
      // Wrong password
      if (typeof e === 'string' && e.indexOf('wrong password') !== -1) {
        renameFail = false;
        passwordWrong = true;
        return
      }
      // Session expired
      if (typeof e === 'string' && e.indexOf('session') !== -1) {
        // Create new session
        renameFail = false;
        passwordWrong = false;
        unlockWalletDialogOpen = true;
        renameWalletDialogOpen = false;
        return
      }
      // Renaming failed
      renameFail = true;
      passwordWrong = false;
    }
  }

  async function generateAccount() {
    await KMDService.SessionGenerateAccount();
    accounts = await KMDService.SessionListAccounts();
  }

  async function showMnemonic(e: SubmitEvent) {
    const formData = new FormData(e.target as HTMLFormElement);
    const walletPassword = formData.get('walletPassword')?.toString() ?? '';
    let mnemonic = '';

    try {
      mnemonic = await KMDService.SessionExportWallet(walletPassword)
    } catch (e: unknown) {
      // Wrong password
      if (typeof e === 'string' && e.indexOf('wrong password') !== -1) {
        passwordWrong = true;
        return
      }
      // Session expired
      if (typeof e === 'string' && e.indexOf('session') !== -1) {
        // Create new session
        passwordWrong = false;
        unlockWalletDialogOpen = true;
        passwordForMnemonicDialogOpen = false;
        return
      }
    }

    // No errors, so show mnemonic
    passwordWrong = false;
    passwordForMnemonicDialogOpen = false;
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
  <div>
    <button class="btn" on:click={() => passwordForMnemonicDialogOpen = true}>
      See mnemonic
    </button>
    <button class="btn" on:click={() => renameWalletDialogOpen = true}>
      Rename
    </button>
    <button class="btn btn-primary" on:click={generateAccount}>
      Generate new account
    </button>
    <button class="btn btn-primary" on:click={() => importAccountDialogOpen = true}>
      Import account
    </button>
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

<Dialog.Root bind:open={unlockWalletDialogOpen}>
  <Dialog.Portal>
    <Dialog.Overlay />
    <Dialog.Content class="modal prose modal-open">
      <div class="modal-box">
        <Dialog.Title class="mt-0">Unlock Wallet</Dialog.Title>
        <form id="unlock-wallet-form" on:submit|preventDefault={unlockWallet} autocomplete="off">
          <div>
            <label class="label" for="wallet-password-input">Wallet password</label>
            <input type="password" name="walletPassword" class="input input-bordered w-full" id="wallet-password-input" required />
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
            <label class="label" for="wallet-password-input">Wallet password</label>
            <input type="password" name="walletPassword" class="input input-bordered w-full" id="wallet-password-input" required />
            {#if passwordWrong}
              <div class="label bg-error px-2">
                <span class="label-text-alt text-error-content">Incorrect password.</span>
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

<Dialog.Root bind:open={passwordForMnemonicDialogOpen}>
  <Dialog.Portal>
    <Dialog.Overlay />
    <Dialog.Content class="modal prose modal-open">
      <div class="modal-box">
        <Dialog.Title class="mt-0">Enter password</Dialog.Title>
        <form id="unlock-wallet-form" on:submit|preventDefault={showMnemonic} autocomplete="off">
          <div>
            <label class="label" for="wallet-password-input">Password</label>
            <input type="password" name="walletPassword" class="input input-bordered w-full" id="wallet-password-input" required />
            {#if passwordWrong}
              <div class="label bg-error px-2">
                <span class="label-text-alt text-error-content">Incorrect password.</span>
              </div>
            {/if}
          </div>
          <div class="modal-action">
            <button type='submit' class="btn btn-primary">Submit</button>
            <a href="/" class="btn">Cancel</a>
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
