import {render, screen} from '@testing-library/svelte';
import { describe, it, expect, vi } from 'vitest';

vi.mock('$lib/wails-bindings/duckysigner/services/kmdservice', () => ({
  CreateWallet: async () => ({
    ID: "NzMxN2MxNmE0MGRjMmMzZjY0MzBkMzYzZDY5NDE3MzY=",
    Name: "dGVzdA==",
    DriverName: "sqlite",
    DriverVersion: 1,
    SupportsMnemonicUX: false,
    SupportsMasterKey: true,
    SupportedTransactions:["pay","keyreg"]
  })
}));

import CreateWalletPage from './+page.svelte';

describe('Create Wallet Page', () => {

  it('has heading', () => {
		render(CreateWalletPage);
    const heading = screen.getByRole('heading', { level: 1 });
    expect(heading).toHaveTextContent('Create New Wallet');
	});

  it('has back button', () => {
		render(CreateWalletPage);
    expect(screen.getByText('Back')).toBeInTheDocument();
  });

  it('has "wallet name" input', () => {
		render(CreateWalletPage);
    expect(screen.getByLabelText('Wallet name')).toBeInTheDocument();
	});

  it('has "wallet password" input', () => {
		render(CreateWalletPage);
    expect(screen.getByLabelText('Wallet password')).toBeInTheDocument();
	});

  it('has "create wallet" button', () => {
		render(CreateWalletPage);
    expect(screen.getByRole('button')).toHaveTextContent('Create wallet')
	});

});
