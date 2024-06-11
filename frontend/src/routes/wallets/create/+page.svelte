<script lang="ts">
  import { goto } from "$app/navigation";
  import { KMDService } from "$lib/wails-bindings/duckysigner/services"

  let walletName = '';
  let walletPassword = '';

  function submitForm() {
    console.log('wallet name: ', walletName);
    console.log('wallet password: ', walletPassword);

    KMDService.CreateWallet(walletName, walletPassword).then(walletInfo => {
      console.log(JSON.stringify(walletInfo));
      goto('/');
    });
  }
</script>

<a href="/" class="btn">Back</a>

<h1 class="text-center text-4xl mb-8">Create New Wallet</h1>
<form on:submit|preventDefault={submitForm} autocomplete="off">
  <div>
    <label class="label" for="wallet-name-input">Wallet name</label>
    <input type="text" bind:value={walletName} class="input input-bordered" id="wallet-name-input" required />

    <label class="label" for="wallet-password-input">Wallet password</label>
    <input type="password" bind:value={walletPassword} class="input input-bordered" id="wallet-password-input" required />
  </div>

  <button type='submit' class="btn btn-primary btn-wide mt-8">Create wallet</button>
</form>
