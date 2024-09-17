<script lang="ts">
  import { page } from '$app/stores';
  import { KMDService } from '$lib/wails-bindings/duckysigner/services';
  import { Events } from '@wailsio/runtime';
  import * as algosdk from 'algosdk';
  import { Dialog } from "bits-ui";
  import { onMount } from 'svelte';

  let walletId = '';
  let acctAddr = '';
  let backLink = '/';
  let txn: algosdk.Transaction;
  let txnFileList: FileList;
  let txnData = '';
  let importedTxnFileIsSigned = false;
  let txnFileInput: HTMLInputElement;
  let walletPasswordDialogOpen = false;
  let passwordWrong = false;
  let walletPassword = '';
  let signedTxnB64 = '';

  // Adapted from https://svelte.dev/repl/file-inputs?version=4.2.18
  $: if (txnFileList) {
    txnData = '';
    importedTxnFileIsSigned = false;
    const file = txnFileList.item(0);
    if (file) {
      processTxnFile(file);
    }

  }

  onMount(async () => {
    walletId = $page.url.searchParams.get('id') ?? '';
    acctAddr = $page.url.searchParams.get('addr') ?? '';
    backLink = (walletId && acctAddr) ? `/account?id=${walletId}&addr=${acctAddr}` : '/';
  });

  /** Converts the given file's contents to bytes as a Uint8Array buffer
   * @param file The file to convert to bytes
   * @returns The file's contents as a Uint8Array buffer
   */
  async function fileToBytes(file: File) {
    const fileData = await new Promise((resolve, reject) => {
      const reader = Object.assign(new FileReader(), {
        onload: () => resolve(reader.result),
        onerror: () => reject(reader.error),
      });
      reader.readAsArrayBuffer(file);
    });
    return new Uint8Array(fileData as ArrayBuffer);
  };

  /** Processes the given file as a signed or unsigned transaction file
   * @param file File to process
   */
  async function processTxnFile(file: File) {
    const txnByteData = await fileToBytes(file);

    // Try decoding file into a `Transaction` object
    try {
      txn = algosdk.decodeUnsignedTransaction(txnByteData);
      // Reset relevant things
      signedTxnB64 = '';
    } catch (error) {
      // Decoding the transaction as an unsigned transaction did not work, so try to decode
      // it as a signed transaction because it may be a signed transaction
      txn = algosdk.decodeSignedTransaction(txnByteData).txn;
      importedTxnFileIsSigned = true;
    }

    txnData = algosdk.encodeJSON(txn, { space: 2 });
  };

  /** Converts bytes as a Uint8Array buffer to data URL.
   *
   * Adapted from:
   * https://developer.mozilla.org/en-US/docs/Glossary/Base64#converting_arbitrary_binary_data
   *
   * @param bytes The bytes to convert to a data URL
   * @param type The MIME type of the data
   * @returns The bytes in the form of a data URL
   */
  export const bytesToDataUrl = async (
    bytes: Uint8Array,
    type = 'application/octet-stream'
  ): Promise<string> => {
    return await new Promise((resolve, reject) => {
      const reader = Object.assign(new FileReader(), {
        onload: () => resolve(reader.result as string),
        onerror: () => reject(reader.error),
      });
      reader.readAsDataURL(new File([bytes], '', { type }));
    });
  };

  /** Converts bytes as a Uint8Array buffer to a Base64-encoded string
   * @param bytes The bytes to convert to a Base64-encoded string
   * @returns The bytes in the form of a Base64-encoded string
   */
  export const bytesToBase64 = async (bytes: Uint8Array) => {
    const dataUrl = await bytesToDataUrl(bytes);
    return dataUrl.slice(dataUrl.indexOf(',') + 1);
  };

  async function signTxn() {
    try {
      const unsignedTxnB64 = await bytesToBase64(txn.toByte());
      signedTxnB64 = await KMDService.SignTransaction(walletId, walletPassword, unsignedTxnB64, acctAddr);
      // Reset wallet password dialog
      passwordWrong = false;
      walletPasswordDialogOpen = false;
      // Reset relevant things
      txnData = '';
    } catch (error) {
      passwordWrong = true;
      console.log(error)
    }
  }

  function resetThings() {
    // Reset things
    txnFileInput.value = "";
    txnData = '';
    importedTxnFileIsSigned = false;
    signedTxnB64 = '';
  }

  function saveTxnFile() {
    Events.Emit({ name: 'saveFile', data: signedTxnB64 })
    resetThings();
  }
</script>

<a href="{backLink}" class="btn">Back</a>

<h1 class="text-center text-4xl mb-8">Sign Transaction</h1>

<div class="pb-4">
  <p>Signing Address: {acctAddr}</p>

  <div class="mt-6">
    <label class="label" for="txn-upload-input">Choose transaction file</label>
    <input bind:files={txnFileList} bind:this={txnFileInput} class="file-input file-input-bordered file-input-primary" type="file" name="txn-upload" id="txn-upload-input" />
  </div>

  {#if txnData}
    <div class="grid place-content-center mb-6">
      <p class="mb-2 text-center">Sign the following transaction?</p>
      <div>
        <button on:click={() => walletPasswordDialogOpen = true} class="btn btn-wide btn-secondary" type="button">Yes</button>
        <button on:click={resetThings} class="btn btn-wide" type="button">No</button>
      </div>
    </div>

    {#if importedTxnFileIsSigned}
      <div class="alert alert-warning mt-2">This transaction has already been signed.</div>
    {/if}

    <code class="card font-mono bg-neutral text-neutral-content whitespace-pre overflow-x-auto p-4">
      {txnData}
    </code>
  {/if}

  {#if signedTxnB64}
    <button type="button" class="btn btn-primary btn-block mt-4" on:click={saveTxnFile}>
      Save signed transaction
    </button>
  {/if}
</div>

<Dialog.Root bind:open={walletPasswordDialogOpen}>
  <Dialog.Portal>
    <Dialog.Overlay />
    <Dialog.Content class="modal prose modal-open">
      <div class="modal-box">
        <Dialog.Title class="mt-0">Sign Transaction</Dialog.Title>
        <form id="unlock-wallet-form" on:submit|preventDefault={signTxn} autocomplete="off">
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
            <button type='submit' class="btn btn-primary">Sign</button>
            <button type="button" class="btn" on:click ={() => walletPasswordDialogOpen = false}>Cancel</button>
          </div>
        </form>
      </div>
    </Dialog.Content>
  </Dialog.Portal>
</Dialog.Root>
