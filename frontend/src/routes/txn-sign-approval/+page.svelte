<script lang="ts">
  import { Events, Window } from '@wailsio/runtime';
  import algosdk from "algosdk";

  let txn: algosdk.Transaction|null = null;

  Events.On('txn_sign_prompt_load', async (e) => {
    // Extract and parse transaction data
    const parsedEvtData: {data: {transaction: string, signer: string}} = JSON.parse(`${e.data}`)
    const txnByteData = algosdk.base64ToBytes(parsedEvtData.data.transaction);
    txn = algosdk.decodeUnsignedTransaction(txnByteData)
  })

  async function sendTxnApproval(approved: boolean) {
    Events.Emit('txn_sign_response', JSON.stringify({approved}))
    Window.Close()
  }
</script>

<div class="h-full">
  <h1 class="mt-4">Approve Transaction</h1>
  <p class="mb-2">Transaction information:</p>

  <code class="card font-mono bg-neutral text-neutral-content whitespace-pre overflow-x-auto p-4">
    {#if txn}
      {algosdk.encodeJSON(txn, {space: 2})}
    {/if}
  </code>
  <div class="mt-6 grid grid-cols-2 gap-2">
    <button type="button" class="btn btn-primary btn-block" on:click={() => sendTxnApproval(true)}>
      Approve
    </button>
    <button type="button" class="btn btn-block" on:click={() => sendTxnApproval(false)}>
      Reject
    </button>
  </div>
</div>
