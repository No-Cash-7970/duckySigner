import {render, screen} from '@testing-library/svelte';
import { describe, it, expect, vi } from 'vitest';

vi.mock('$lib/wails-bindings/duckysigner/services/kmdservice', () => ({
  ImportWallet: async () => ({
    ID: "NzMxN2MxNmE0MGRjMmMzZjY0MzBkMzYzZDY5NDE3MzY=",
    Name: "dGVzdA==",
    DriverName: "sqlite",
    DriverVersion: 1,
    SupportsMnemonicUX: false,
    SupportsMasterKey: true,
    SupportedTransactions:["pay","keyreg"]
  })
}));

vi.mock('$lib/wails-bindings/duckysigner/services/dappconnectservice', () => ({}));

vi.mock('@wailsio/runtime', () => ({}));

import ImportWalletPage from './+page.svelte';

describe('Import Wallet Page', () => {

  it('has heading', () => {
		render(ImportWalletPage);
    const heading = screen.getByRole('heading', { level: 1 });
    expect(heading).toHaveTextContent('Import Wallet');
	});

  it('has back button', () => {
		render(ImportWalletPage);
    expect(screen.getByText('Back')).toBeInTheDocument();
  });

  it('has "wallet name" input', () => {
		render(ImportWalletPage);
    expect(screen.getByLabelText('Wallet name')).toBeInTheDocument();
	});

  it('has "wallet password" input', () => {
		render(ImportWalletPage);
    expect(screen.getByLabelText('Wallet password')).toBeInTheDocument();
	});

  it('has "import wallet" button', () => {
		render(ImportWalletPage);
    expect(screen.getByRole('button')).toHaveTextContent('Import wallet')
	});

});
