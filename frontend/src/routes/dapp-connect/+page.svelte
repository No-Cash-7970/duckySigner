<script lang="ts">
  import { KMDService } from '$lib/wails-bindings/duckysigner/services'
  import { Events, Window } from '@wailsio/runtime'
  import { onMount } from 'svelte'

  let dappData = { dapp: { name: "", uri: "", desc: "", icon: "" } }
  let confirmCode = ''
  let accounts: string[] = []
  let selectedAccounts: string[] = []

  onMount(async () => {
    accounts = await KMDService.SessionListAccounts()
  })

  Events.On('session_confirm_prompt_load', (e) => {
    dappData = JSON.parse(`${e.data}`)
  })

  async function confirmConnect() {
    Events.Emit(
      'session_confirm_response',
      JSON.stringify({ code: confirmCode, addrs: selectedAccounts }),
    )
    Window.Close()
  }
</script>

<div class="h-full">
  <h1 class="mt-4">Connect to DApp</h1>
  <p class="mb-2">DApp information:</p>

  <ul class="mt-0 mb-8">
    <li>Name: <span>{dappData.dapp.name}</span></li>
    <li>URI:
      {#if dappData.dapp.uri}
        <div>{dappData.dapp.uri}</div>
      {:else}
        <i>None</i>
      {/if}
    </li>
    <li>Description:
      {#if dappData.dapp.desc}
        <div>{dappData.dapp.desc}</div>
      {:else}
        <i>None</i>
      {/if}
    </li>
    <li>Icon:
      {#if dappData.dapp.icon}
        <div>{dappData.dapp.icon}</div>
      {:else}
        <i>None</i>
      {/if}
    </li>
  </ul>

  {#if dappData}
    <form on:submit|preventDefault={confirmConnect} autocomplete="off" class="pb-24">
      <fieldset class="fieldset bg-base-200 border-base-300 rounded-box w-xs border p-4 mx-4">
        <legend class="fieldset-legend">Enter confirmation code</legend>
        <label class="label" for="confirm-code-input">Confirmation code</label>
        <input type="text" bind:value={confirmCode} class="input input-bordered w-full" id="confirm-code-input" required />
      </fieldset>

      <fieldset class="fieldset bg-base-200 border-base-300 rounded-box w-xs border p-4 mx-4">
        <legend class="fieldset-legend">Choose accounts to connect</legend>
        {#if accounts.length > 0}
          {#each accounts as address}
            <label class="label place-content-start align-middle">
              <input type="checkbox" class="checkbox me-2" bind:group={selectedAccounts} value={address} />
              <span class="break-all">{address}</span>
            </label>
          {/each}
        {:else}
          <p class="text-center italic">No accounts</p>
        {/if}
      </fieldset>

      <div class="p-4 bg-base-100 fixed bottom-0 start-0 w-full">
        <button type='submit' class="btn btn-primary" disabled={selectedAccounts.length === 0 || confirmCode === ''}>
          Confirm connection
        </button>
        <button type='button' class="btn" on:click={() => Window.Close()}>
          Cancel
        </button>
      </div>
    </form>
  {/if}
</div>
