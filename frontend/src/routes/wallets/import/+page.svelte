<script lang="ts">
  import { goto } from "$app/navigation";
  import { KMDService } from "$lib/wails-bindings/duckysigner/services";

  let importFail = false;

  async function importWallet(e: SubmitEvent) {
    const formData = new FormData(e.target as HTMLFormElement);
    const mnemonic = formData.get('mnemonic')?.toString() ?? '';
    const walletName = formData.get('walletName')?.toString() ?? '';
    const walletPassword = formData.get('walletPassword')?.toString() ?? '';

    try {
      await KMDService.ImportWalletMnemonic(mnemonic, walletName, walletPassword);
      importFail = false;
      goto('/');
    } catch (error) {
      importFail = true;
    }
  }
</script>

<a href="/" class="btn">Back</a>

<h1 class="text-center text-4xl mb-8">Import Wallet</h1>
<form on:submit|preventDefault={importWallet} autocomplete="off">
  {#if importFail}
    <div class="label bg-error px-2">
      <span class="label-text-alt text-error-content">Cannot import wallet.</span>
    </div>
  {/if}
  <div>
    <div>
      <label class="label" for="mnemonic-input">Mnemonic</label>
      <textarea name="mnemonic" class="textarea textarea-bordered w-full" id="mnemonic-input" required></textarea>
    </div>
    <label class="label" for="wallet-name-input">Wallet name</label>
    <input type="text" name="walletName" class="input input-bordered" id="wallet-name-input" required />

    <label class="label" for="wallet-password-input">Wallet password</label>
    <input type="password" name="walletPassword" class="input input-bordered" id="wallet-password-input" required />
  </div>

  <button type='submit' class="btn btn-primary btn-wide mt-8">Import wallet</button>
</form>
