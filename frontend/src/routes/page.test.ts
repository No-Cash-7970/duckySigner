import {render, screen} from '@testing-library/svelte';
import { describe, it, expect, vi } from 'vitest';

import HomePage from './+page.svelte';
import userEvent from '@testing-library/user-event';

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

const startServerFnMock = vi.fn();
const stopServerFnMock = vi.fn();
vi.mock('$lib/wails-bindings/duckysigner/services/dappconnectservice', () => ({
  Start: () => startServerFnMock(),
  Stop: () => stopServerFnMock(),
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

  it('turns on the server when page is first loaded', async () => {
    startServerFnMock.mockClear();
    startServerFnMock.mockResolvedValue(true);
		render(HomePage);

    expect(await screen.findByText(/Turn off dApp/)).toBeInTheDocument();
    expect(startServerFnMock).toHaveBeenCalledOnce();
  });

  it('turns off the server when "turn server off" button is clicked', async () => {
    startServerFnMock.mockClear();
    startServerFnMock.mockResolvedValue(true);
		render(HomePage);

    stopServerFnMock.mockClear();
    stopServerFnMock.mockResolvedValue(false);
    await userEvent.click(await screen.findByText(/Turn off dApp/));

    expect(await screen.findByText(/Turn on dApp/)).toBeInTheDocument();
    expect(screen.queryByText(/Turn off dApp/)).not.toBeInTheDocument();
    expect(stopServerFnMock).toHaveBeenCalledOnce();
  });

  it('turns the server on when "turn server on" button is clicked', async () => {
    startServerFnMock.mockClear();
    startServerFnMock.mockResolvedValue(false);
		render(HomePage);

    startServerFnMock.mockResolvedValue(true);
    await userEvent.click(await screen.findByText(/Turn on dApp/));

    expect(await screen.findByText(/Turn off dApp/)).toBeInTheDocument();
    expect(screen.queryByText(/Turn on dApp/)).not.toBeInTheDocument();
    expect(startServerFnMock).toHaveBeenCalledTimes(2);
  });

  it('still shows "turn server on" button after it was clicked if server failed to turn on',
  async () => {
    startServerFnMock.mockClear();
    startServerFnMock.mockResolvedValue(false);
		render(HomePage);

    await userEvent.click(await screen.findByText(/Turn on dApp/));

    expect(await screen.findByText(/Turn on dApp/)).toBeInTheDocument();
    expect(screen.queryByText(/Turn off dApp/)).not.toBeInTheDocument();
    expect(startServerFnMock).toHaveBeenCalledTimes(2);
  });

  it('still shows "turn server off" button after it was clicked if server failed to turn off',
  async () => {
    startServerFnMock.mockResolvedValue(true);
		render(HomePage);

    stopServerFnMock.mockClear();
    stopServerFnMock.mockResolvedValue(true);
    await userEvent.click(await screen.findByText(/Turn off dApp/));

    expect(await screen.findByText(/Turn off dApp/)).toBeInTheDocument();
    expect(screen.queryByText(/Turn on dApp/)).not.toBeInTheDocument();
    expect(stopServerFnMock).toHaveBeenCalledOnce();
  });

});
