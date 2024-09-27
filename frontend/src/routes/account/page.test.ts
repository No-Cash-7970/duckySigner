import {render, screen} from '@testing-library/svelte';
import { describe, it, expect, vi } from 'vitest';
import { writable } from 'svelte/store';
import userEvent from '@testing-library/user-event';

import AccountInfoPage from './+page.svelte';

vi.mock('$app/stores', () => ({
  page: writable({url: { searchParams: {
    get: (p: string) => {
      if (p === 'id') return '123' // Wallet ID
      if (p === 'addr') return 'H3PFTYORQCTLIN7PEPDCYI4ALUHNE4CE5GJIPLZA3ZBKWG23TWND4IP47A' // Account address
    }
  }}})
}));

const gotoMockFunc = vi.fn();
vi.mock('$app/navigation', () => ({ goto: () => gotoMockFunc() }));

vi.mock('$lib/wails-bindings/duckysigner/services/kmdservice', () => ({
  RemoveAccountFromWallet: async () => { return },
  CheckWalletPassword: async (id: string, pw: string) => {
    if (pw !== 'badpassword') throw Error;
  },
  ExportAccountInWallet: async () => 'abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon',
}));

vi.mock('$lib/wails-bindings/duckysigner/services/dappconnectservice', () => ({}));

vi.mock('algosdk', async (importOriginal) => {
  const actual = await importOriginal<typeof import('algosdk')>();
  return {
    microalgosToAlgos: actual.microalgosToAlgos,
    Algodv2: class {
      constructor () {
        return {
          accountInformation: () => ({
            do: async () => ({
              amount: 5_000_000n,
              minBalance: 100_000n,
              assets: [{},{},{},{},{},{},{},]
            })
          }),
        }
      }
    },
  }
});

describe('Account Information Page', () => {

  it('has address as heading', async () => {
		render(AccountInfoPage);
    expect(await screen.findByText('H3PFTYORQCTLIN7PEPDCYI4ALUHNE4CE5GJIPLZA3ZBKWG23TWND4IP47A')).toHaveRole('heading');
	});

  it('has back button', () => {
		render(AccountInfoPage);
    expect(screen.getByText('Back')).toBeInTheDocument();
  });

  it('shows account information', async () => {
		render(AccountInfoPage);
    // Balance
    expect(await screen.findByText('5')).toBeInTheDocument();
    // Min balance
    expect(await screen.findByText('0.1')).toBeInTheDocument();
    // Number of assets
    expect(await screen.findByText('7')).toBeInTheDocument();
  });

  it('can remove account from wallet', async () => {
		render(AccountInfoPage);

    await userEvent.click(screen.getByText(/Remove from wallet/));
    await userEvent.click(screen.getByLabelText(/Wallet password/));
    await userEvent.paste('badpassword');
    await userEvent.click(screen.getByText('Remove this account'));

    expect(gotoMockFunc).toHaveBeenCalled();
  });

  it('can show mnemonic', async () => {
		render(AccountInfoPage);

    await userEvent.click(screen.getByText(/See mnemonic/));
    // Enter password
    await userEvent.click(screen.getByLabelText(/Wallet password/));
    await userEvent.paste('badpassword');
    await userEvent.click(screen.getByText(/Continue/));

    expect((await screen.findAllByText('abandon')).length).toBe(25);
  });

});
