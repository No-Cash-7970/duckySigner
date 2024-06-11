import {render, screen} from '@testing-library/svelte';
import { describe, it, expect, vi } from 'vitest';
import { writable } from 'svelte/store';
import userEvent from '@testing-library/user-event';


import WalletInfoPage from './+page.svelte';

const walletsList = [
  'H3PFTYORQCTLIN7PEPDCYI4ALUHNE4CE5GJIPLZA3ZBKWG23TWND4IP47A',
  'V3NC4VRDRP33OI2R5AQXEOOXFXXRYHWDKJOCGB64C7QRCF2IWNWHPFZ4QU',
];

vi.mock('$lib/wails-bindings/duckysigner/services/kmdservice', () => ({
  GetWalletInfo: async () => ({
    ID: 'MTIz', // '123'
    Name: 'Rm9vYmFy', // 'Foobar'
  }),
  ListAccountsInWallet: async () => walletsList,
  CheckWalletPassword: async (id: string, pw: string) => {
    if (pw !== 'badpassword') throw Error;
  },
  GenerateWalletAccount: async () => {
    const newWallet = 'RMAZSNHVLAMY5AUWWTSDON4S2HIUV7AYY6MWWEMKYH63YLHAKLZNHQIL3A';
    walletsList.push(newWallet);
    return newWallet;
  },
}));

vi.mock('$app/stores', () => ({
  page: writable({url: {searchParams: { get: () => '123' }}})
}))

describe('Wallet Information Page', () => {

  it('has wallet name as heading', async () => {
		render(WalletInfoPage);
    const link = await screen.findByRole('heading', { level: 1 });
    expect(link).toHaveTextContent('Foobar');
	});

  it('has back button', () => {
		render(WalletInfoPage);
    expect(screen.getByText('Back')).toBeInTheDocument();
  });

  it('asks for wallet password', async () => {
		render(WalletInfoPage);

    await userEvent.click(screen.getByLabelText(/Wallet password/));
    await userEvent.paste('badpassword');
    await userEvent.click(screen.getByText('Unlock wallet'));

    expect(screen.queryByText('Unlock Wallet')).not.toBeInTheDocument();
  });

  it('shows error message when given wrong wallet password', async () => {
		render(WalletInfoPage);

    await userEvent.click(screen.getByLabelText(/Wallet password/));
    await userEvent.paste('goodpassword');
    await userEvent.click(screen.getByText('Unlock wallet'));

    expect(await screen.findByText(/Incorrect password/)).toBeInTheDocument();
  });

  it('can add account', async () => {
		render(WalletInfoPage);

    // Unlock wallet
    await userEvent.click(screen.getByLabelText(/Wallet password/));
    await userEvent.paste('badpassword');
    await userEvent.click(screen.getByText('Unlock wallet'));
    // Add account
    await userEvent.click(screen.getByText(/Add account/));

    expect(await screen.findByText('H3PFTYORQCTLIN7PEPDCYI4ALUHNE4CE5GJIPLZA3ZBKWG23TWND4IP47A')).toBeInTheDocument();
    expect(await screen.findByText('V3NC4VRDRP33OI2R5AQXEOOXFXXRYHWDKJOCGB64C7QRCF2IWNWHPFZ4QU')).toBeInTheDocument();
    expect(await screen.findByText('RMAZSNHVLAMY5AUWWTSDON4S2HIUV7AYY6MWWEMKYH63YLHAKLZNHQIL3A')).toBeInTheDocument();
  });

  it('shows accounts in wallet', async () => {
		render(WalletInfoPage);

    // Unlock wallet
    await userEvent.click(screen.getByLabelText(/Wallet password/));
    await userEvent.paste('badpassword');
    await userEvent.click(screen.getByText('Unlock wallet'));

    expect(await screen.findByText('H3PFTYORQCTLIN7PEPDCYI4ALUHNE4CE5GJIPLZA3ZBKWG23TWND4IP47A')).toBeInTheDocument();
    expect(await screen.findByText('V3NC4VRDRP33OI2R5AQXEOOXFXXRYHWDKJOCGB64C7QRCF2IWNWHPFZ4QU')).toBeInTheDocument();
  });

});
