import {render, screen} from '@testing-library/svelte';
import { describe, it, expect, vi } from 'vitest';
import { writable } from 'svelte/store';


import ShowWalletPage from './+page.svelte';

vi.mock('$lib/wails-bindings/duckysigner/services/kmdservice', () => ({
  GetWalletInfo: async () => ({
    ID: 'MTIz', // '123'
    Name: 'Rm9vYmFy', // 'Foobar'
  }),
  ListAccountsInWallet: async () => [
    'H3PFTYORQCTLIN7PEPDCYI4ALUHNE4CE5GJIPLZA3ZBKWG23TWND4IP47A',
    'V3NC4VRDRP33OI2R5AQXEOOXFXXRYHWDKJOCGB64C7QRCF2IWNWHPFZ4QU',
    'RMAZSNHVLAMY5AUWWTSDON4S2HIUV7AYY6MWWEMKYH63YLHAKLZNHQIL3A',
  ],
}));

vi.mock('$app/stores', () => ({
  page: writable({url: {searchParams: { get: () => '123' }}})
}))

describe('Show Wallet', () => {

  it('has wallet name as heading', async () => {
		render(ShowWalletPage);
    const link = await screen.findByRole('heading', { level: 1 });
    expect(link).toHaveTextContent('Foobar');
	});

  it('has back button', () => {
		render(ShowWalletPage);
    expect(screen.getByText('Back')).toBeInTheDocument();
  });

  // TODO: Test for buttons

  // TODO: Test accounts list

});
