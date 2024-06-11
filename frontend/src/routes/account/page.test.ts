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

describe('Account Information Page', () => {

  it('has address as heading', async () => {
		render(AccountInfoPage);

    const heading = await screen.findByRole('heading', { level: 1 });
    expect(heading).toHaveTextContent('H3PFTYORQCTLIN7PEPDCYI4ALUHNE4CE5GJIPLZA3ZBKWG23TWND4IP47A');
	});

  it('has back button', () => {
		render(AccountInfoPage);
    expect(screen.getByText('Back')).toBeInTheDocument();
  });

});
