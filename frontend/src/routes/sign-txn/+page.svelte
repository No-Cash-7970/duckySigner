<script lang="ts">
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import { KMDService } from '$lib/wails-bindings/duckysigner/services';
  import * as algosdk from 'algosdk';
  import { Dialog } from "bits-ui";
  import { onMount } from 'svelte';

  let walletId = '';
  let acctAddr = '';
  let backLink = '/';
  let txnFileList: FileList;
  let txnData = '';
  let txnFileIsSigned = false;
  let txnFileInput: HTMLInputElement;

  // Adapted from https://svelte.dev/repl/file-inputs?version=4.2.18
  $: if (txnFileList) {
    txnData = '';
    txnFileIsSigned = false;
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
    let txn: algosdk.Transaction;

    // Try decoding file into a `Transaction` object
    try {
      txn = algosdk.decodeUnsignedTransaction(txnByteData);
    } catch (error) {
      // Decoding the transaction as an unsigned transaction did not work, so try to decode
      // it as a signed transaction because it may be a signed transaction
      txn = algosdk.decodeSignedTransaction(txnByteData).txn;
      txnFileIsSigned = true;
    }

    txnData = JSON.stringify(JSON.parse(txn.toString()), null, 2);
  };

  async function signTxn() {
    // TODO: Sign transaction using KMD service
  }

  function cancelSignTxn() {
    // Reset things
    txnFileInput.value = "";
    txnData = '';
    txnFileIsSigned = false;
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
        <button on:click={signTxn} class="btn btn-wide btn-secondary" type="button">Yes</button>
        <button on:click={cancelSignTxn} class="btn btn-wide" type="button">No</button>
      </div>
    </div>

    {#if txnFileIsSigned}
      <div class="alert alert-warning mt-2">This transaction has already been signed.</div>
    {/if}

    <code class="card font-mono bg-neutral text-neutral-content whitespace-pre overflow-x-auto p-4">
      {txnData}
    </code>
  {/if}
</div>
