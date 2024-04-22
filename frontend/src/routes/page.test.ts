import {render, screen} from '@testing-library/svelte';
import { describe, it, expect, vi } from 'vitest';

import Page from './+page.svelte';

vi.mock('@wailsio/runtime', () => ({
  Events: { On: () => (() => {}) }
}));

vi.mock('../lib/wails-bindings/services/GreetService', () => ({
  Greet: async () => 'Hello user!'
}));

describe('Home', () => {

  it('has "wallets list" link', () => {
		render(Page);
    const link = screen.getByText('Wallets List');
    expect(link).toHaveRole('link');
	});

});
