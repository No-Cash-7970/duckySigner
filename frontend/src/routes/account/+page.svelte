<script lang="ts">
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import { KMDService } from '$lib/wails-bindings/duckysigner/services';
  import { Dialog } from "bits-ui";
  import { onMount } from 'svelte';
  import { Algodv2, microalgosToAlgos } from 'algosdk';

  let walletId = '';
  let acctAddr = '';
  let backLink = '/';
  let deleteAcctDialogOpen = false;
  let askPassForMnemonicDialogOpen = false;
  let mnemonicDialogOpen = false;
  let passwordWrong = false;
  let walletPassword = '';
  let mnemonicParts: string[] = [];
  const algodClient = new Algodv2('', 'https://testnet-api.algonode.cloud', 443);
  let acctInfo: any;

  onMount(async () => {
    walletId = $page.url.searchParams.get('id') ?? '';
    acctAddr = $page.url.searchParams.get('addr') ?? '';
    backLink = walletId ? `/wallets?id=${walletId}` : '/';
    acctInfo = await algodClient.accountInformation(acctAddr).do();
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

  async function showMnemonic() {
    askPassForMnemonicDialogOpen = false;
    mnemonicParts = (await KMDService.ExportWalletMnemonic(walletId, walletPassword)).split(' ');
    mnemonicDialogOpen = true;
  }
</script>

<a href="{backLink}" class="btn">Back</a>

<h1 class="text-center text-4xl mb-8 truncate">{acctAddr}</h1>

{#if acctInfo}
  <div class="stats shadow">
    <div class="stat">
      <div class="stat-title">Balance</div>
      <div class="stat-value">{microalgosToAlgos(acctInfo.amount)}</div>
      <div class="stat-desc">Algos</div>
    </div>
    <div class="stat">
      <div class="stat-title">Minimum balance</div>
      <div class="stat-value">{microalgosToAlgos(acctInfo['min-balance'])}</div>
      <div class="stat-desc">Algos</div>
    </div>
    <div class="stat">
      <div class="stat-title">Number of assets</div>
      <div class="stat-value">{acctInfo.assets.length}</div>
    </div>
  </div>
{/if}

<div class="mt-6">
  <button class="btn" on:click={() => askPassForMnemonicDialogOpen = true}>See mnemonic</button>
  <button class="btn btn-error" on:click={() => deleteAcctDialogOpen = true}>Remove from wallet</button>
  <!-- TODO: Sign txn file -->
  <button class="btn btn-secondary">Sign transaction file</button>
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
            <Dialog.Close class="btn">Close</Dialog.Close>
          </div>
        </form>
      </div>
    </Dialog.Content>
  </Dialog.Portal>
</Dialog.Root>

<Dialog.Root bind:open={askPassForMnemonicDialogOpen}>
  <Dialog.Portal>
    <Dialog.Overlay />
    <Dialog.Content class="modal prose modal-open">
      <div class="modal-box">
        <Dialog.Title class="mt-0">Enter wallet password</Dialog.Title>
        <Dialog.Description>Enter wallet password to see this account's mnemonic.</Dialog.Description>
        <form id="unlock-wallet-form" on:submit|preventDefault={showMnemonic} autocomplete="off">
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
            <button type='submit' class="btn btn-primary">Continue</button>
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
