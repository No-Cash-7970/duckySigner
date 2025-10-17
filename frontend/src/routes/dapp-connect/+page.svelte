<script lang="ts">
  import { Events, Window } from '@wailsio/runtime';

  let dappData = { dapp: { name: "", uri: "", desc: "", icon: "" } };
  let confirmCode = '';

  Events.On('session_confirm_prompt_load', (e) => {
    dappData = JSON.parse(`${e.data}`)
  })

  async function confirmConnect() {
    Events.Emit(
      'session_confirm_response',
      JSON.stringify([{ code: confirmCode, addrs: [] }]),
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
    <form on:submit|preventDefault={confirmConnect} autocomplete="off">
      <fieldset class="fieldset bg-base-200 border-base-300 rounded-box w-xs border p-4 mx-4">
        <legend class="fieldset-legend">Enter confirmation code</legend>

        <label class="label" for="confirm-code-input">Confirmation code</label>
        <input type="text" bind:value={confirmCode} class="input input-bordered w-full" id="confirm-code-input" required />

        <button type='submit' class="btn btn-block btn-primary mt-4">Confirm connection</button>
      </fieldset>
    </form>
  {/if}
</div>
