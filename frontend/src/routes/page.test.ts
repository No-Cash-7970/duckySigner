import {render, screen} from '@testing-library/svelte';
import { describe, it, expect } from 'vitest';

import Page from './+page.svelte';

describe('Home', () => {

  it('has link', () => {
		render(Page);
    const link = screen.getByRole('link');
    expect(link).toBeInTheDocument();
	});

});
