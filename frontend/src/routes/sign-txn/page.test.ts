import {render, screen} from '@testing-library/svelte';
import { describe, it, expect, vi } from 'vitest';
import { writable } from 'svelte/store';
import userEvent from '@testing-library/user-event';

import SignTxnPage from './+page.svelte';

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

vi.mock('algosdk', async (importOriginal) => {
  const actual = await importOriginal<typeof import('algosdk')>();
  return {
    microalgosToAlgos: actual.microalgosToAlgos,
    Algodv2: class {
      constructor () {
        return {
          accountInformation: () => ({do: async () => ({
            amount: 5_000_000,
            'min-balance': 100_000,
            assets: [{},{},{},{},{},{},{},]
          })}),
        }
      }
    },
  }
});

describe('Sign Transaction Page', () => {

  it('has heading', async () => {
		render(SignTxnPage);
    expect(await screen.findByText('Sign Transaction')).toHaveRole('heading');
	});

  it('has back button', () => {
		render(SignTxnPage);
    expect(screen.getByText('Back')).toBeInTheDocument();
  });

});
