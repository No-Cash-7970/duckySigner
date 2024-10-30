import {render, screen} from '@testing-library/svelte';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { writable } from 'svelte/store';
import userEvent from '@testing-library/user-event';

import WalletInfoPage from './+page.svelte';

const { walletsList } = vi.hoisted(() => ({
  walletsList: [
    'H3PFTYORQCTLIN7PEPDCYI4ALUHNE4CE5GJIPLZA3ZBKWG23TWND4IP47A',
    'V3NC4VRDRP33OI2R5AQXEOOXFXXRYHWDKJOCGB64C7QRCF2IWNWHPFZ4QU',
  ]
}));
const { walletInfo } = vi.hoisted(() => ({
  walletInfo: {
    ID: 'MTIz', // '123'
    Name: 'Rm9vYmFy', // 'Foobar'
  }
}));
const { kmdServiceMock } = vi.hoisted(() => ({
  kmdServiceMock: {
    GetWalletInfo: vi.fn().mockResolvedValue(walletInfo),
    SessionListAccounts: vi.fn().mockResolvedValue(walletsList),
    StartSession: async (id: string, pw: string) => {
      if (pw !== 'badpassword') throw Error;
    },
    SessionGenerateAccount: vi.fn().mockImplementation(async () => {
      const newWallet = 'RMAZSNHVLAMY5AUWWTSDON4S2HIUV7AYY6MWWEMKYH63YLHAKLZNHQIL3A';
      walletsList.push(newWallet);
      return newWallet;
    }),
    RenameWallet: vi.fn().mockImplementation(
      async (id: string, name: string, pw: string) => walletInfo.Name = btoa(name)
    ),
    SessionExportWallet: async () => 'abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon',
    SessionImportAccount: vi.fn().mockImplementation(async () => {
      const newWallet = '7JDB2I2R4ZXN4BAGZMRKYPZGKOTABRAG4KN2R7TWOAGMBCLUZXIMVLMA2M';
      walletsList.push(newWallet);
      return newWallet;
    }),
    SessionIsForWallet: vi.fn(),
    SessionCheck: vi.fn(),
  }
}));
vi.mock('$lib/wails-bindings/duckysigner/services/kmdservice', () => kmdServiceMock);

vi.mock('$lib/wails-bindings/duckysigner/services/dappconnectservice', () => ({}));

vi.mock('$app/stores', () => ({
  page: writable({url: {searchParams: { get: () => '123' }}})
}))

vi.mock('@wailsio/runtime', () => ({}));

describe('Wallet Information Page', () => {
  beforeEach(() => {
    kmdServiceMock.SessionIsForWallet.mockResolvedValue(true);
  })

  it('has wallet name as heading', async () => {
		render(WalletInfoPage);
    expect(await screen.findByText(/Foobar/)).toHaveRole('heading');
	});

  it('has back button', () => {
		render(WalletInfoPage);
    expect(screen.getByText('Back')).toBeInTheDocument();
  });

  it('asks for & accepts wallet password when wallet session is NOT for the current wallet', async () => {
    kmdServiceMock.SessionIsForWallet.mockResolvedValue(false);
		render(WalletInfoPage);

    await userEvent.click(await screen.findByLabelText(/Wallet password/));
    await userEvent.paste('badpassword');
    await userEvent.click(screen.getByText('Unlock wallet'));

    expect(screen.queryByText('Unlock Wallet')).not.toBeInTheDocument();
	});

  it('shows error message when given wrong wallet password', async () => {
    kmdServiceMock.SessionIsForWallet.mockResolvedValue(false);
		render(WalletInfoPage);

    await userEvent.click(await screen.findByLabelText(/Wallet password/));
    await userEvent.paste('goodpassword');
    await userEvent.click(screen.getByText('Unlock wallet'));

    expect(await screen.findByText(/Incorrect password/)).toBeInTheDocument();
  });

  it('can generate a new account', async () => {
		render(WalletInfoPage);

    await userEvent.click(await screen.findByText(/Generate new account/));

    expect(await screen.findByText('H3PFTYORQCTLIN7PEPDCYI4ALUHNE4CE5GJIPLZA3ZBKWG23TWND4IP47A')).toBeInTheDocument();
    expect(await screen.findByText('V3NC4VRDRP33OI2R5AQXEOOXFXXRYHWDKJOCGB64C7QRCF2IWNWHPFZ4QU')).toBeInTheDocument();
    expect(await screen.findByText('RMAZSNHVLAMY5AUWWTSDON4S2HIUV7AYY6MWWEMKYH63YLHAKLZNHQIL3A')).toBeInTheDocument();
  });

  it('can import an account', async () => {
		render(WalletInfoPage);

    await userEvent.click(await screen.findByText('Import account'));
    await userEvent.click(screen.getByLabelText('Mnemonic'));
    await userEvent.paste('abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon');
    await userEvent.click(screen.getByText('Import'));

    expect(await screen.findByText('7JDB2I2R4ZXN4BAGZMRKYPZGKOTABRAG4KN2R7TWOAGMBCLUZXIMVLMA2M')).toBeInTheDocument();
  });

  it('can rename wallet', async () => {
		render(WalletInfoPage);

    // Click Button
    await userEvent.click(await screen.findByText(/Rename/));
    // Fill out form
    await userEvent.click(await screen.findByLabelText(/New wallet name/));
    await userEvent.paste('Baz Qux');
    await userEvent.click(screen.getByLabelText(/Wallet password/));
    await userEvent.paste('badpassword');
    await userEvent.click(screen.getByText('Rename wallet'));

    expect(await screen.findByText(/Baz Qux/)).toHaveRole('heading');
  });

  it('shows accounts in wallet', async () => {
		render(WalletInfoPage);

    expect(await screen.findByText(/H3PFTYORQCTLIN7PEPDCYI4ALUHNE4CE5GJIPLZA3ZBKWG23TWND4IP47A/)).toBeInTheDocument();
    expect(await screen.findByText(/V3NC4VRDRP33OI2R5AQXEOOXFXXRYHWDKJOCGB64C7QRCF2IWNWHPFZ4QU/)).toBeInTheDocument();
  });

  it('can show mnemonic', async () => {
		render(WalletInfoPage);

    // Try to see mnemonic
    await userEvent.click(await screen.findByText(/See mnemonic/));
    // Enter password
    await userEvent.click(await screen.findByLabelText(/Password/));
    await userEvent.paste('badpassword');
    await userEvent.click(screen.getByText('Submit'));

    expect((await screen.findAllByText('abandon')).length).toBe(25);
  });

});
