import {render, screen} from '@testing-library/svelte';
import { describe, it, expect, vi } from 'vitest';

import HomePage from './+page.svelte';

vi.mock('$lib/wails-bindings/duckysigner/services/kmdservice', () => ({
  ListWallets: async () => [
    {
      Name: 'Rm9vIFdhbGxldA==', // 'Foo Wallet'
      ID: 'MQ==', // '1'
    },
    {
      Name: 'QmFyIFdhbGxldA==', // 'Bar Wallet'
      ID: 'Mg==', // '2'
    },
    {
      Name: 'QmF6IFdhbGxldA==', // 'Baz Wallet'
      ID: 'Mw==', // '3'
    },
  ]
}));

describe('Home', () => {

  it('has "import wallet" link', () => {
		render(HomePage);
    const link = screen.getByText('Import wallet');
    expect(link).toHaveRole('link');
	});

  it('has "create wallet" link', () => {
		render(HomePage);
    const link = screen.getByText('Create wallet');
    expect(link).toHaveRole('link');
	});

  it('lists wallets', async () => {
		render(HomePage);
    expect(await screen.findByText('Foo Wallet')).toHaveRole('link');
    expect(await screen.findByText('Bar Wallet')).toHaveRole('link');
    expect(await screen.findByText('Baz Wallet')).toHaveRole('link');
  });

});
