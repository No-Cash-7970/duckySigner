import {render, screen} from '@testing-library/svelte';
import { describe, it, expect, vi } from 'vitest';
import { writable } from 'svelte/store';
import userEvent from '@testing-library/user-event';
import * as fs from "node:fs";

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
    ...actual,
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

  it('displays transaction information of valid UNSIGNED transaction file', async () => {
    const data = fs.readFileSync('src/testing/test_unsigned.txn.msgpack');
    const file = new File([data], 'unsigned.txn.msgpack', { type: 'application/octet-stream' });

		render(SignTxnPage);

    const fileInput = screen.getByLabelText('Choose transaction file');
    await userEvent.upload(fileInput, file);

    expect(await screen.findByRole('code')).not.toBeEmptyDOMElement();
  });

  it('displays transaction information of valid SIGNED transaction file', async () => {
    const data = fs.readFileSync('src/testing/test_signed.txn.msgpack');
    const file = new File([data], 'signed.txn.msgpack', { type: 'application/octet-stream' });

		render(SignTxnPage);

    const fileInput = screen.getByLabelText('Choose transaction file');
    await userEvent.upload(fileInput, file);

    expect(await screen.findByRole('code')).not.toBeEmptyDOMElement();
    expect(await screen.findByText('This transaction has already been signed.')).toBeInTheDocument();
  });

  it('clears file input and displayed transaction data when "No" button is clicked', async () => {
    const data = fs.readFileSync('src/testing/test_signed.txn.msgpack');
    const file = new File([data], 'signed.txn.msgpack', { type: 'application/octet-stream' });

		render(SignTxnPage);

    await userEvent.upload(screen.getByLabelText('Choose transaction file'), file);
    await userEvent.click(await screen.findByText('No'));

    // TODO: Test if file input is empty
    expect(screen.queryByRole('code')).not.toBeInTheDocument();
  });

  it.skip('signs transaction when "Yes" button is clicked', async () => {
    const data = fs.readFileSync('src/testing/test_signed.txn.msgpack');
    const file = new File([data], 'signed.txn.msgpack', { type: 'application/octet-stream' });

		render(SignTxnPage);

    await userEvent.upload(screen.getByLabelText('Choose transaction file'), file);
    await userEvent.click(await screen.findByText('Yes'));

    // TODO: Test if file input is empty
    expect(screen.queryByRole('code')).not.toBeInTheDocument();
    // TODO: Test if transaction is signed and can be saved as a file
  });

});
