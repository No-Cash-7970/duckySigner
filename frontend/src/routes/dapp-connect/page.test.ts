import {render, screen} from '@testing-library/svelte';
import { describe, it, expect, vi } from 'vitest';
import userEvent from '@testing-library/user-event';

import DappConnectPage from './+page.svelte';

const eventEmitFunc = vi.fn();
const windowCloseFunc = vi.fn();
vi.mock('@wailsio/runtime', () => ({
  Events: {
    Emit: () => eventEmitFunc(),
    On: vi.fn().mockImplementation((evtName: string, cb: Function) => {
      cb({data: '{"dapp":{"name":"Foo DApp","uri":"http://example.com","desc":"Foobar","icon":""}}'});
    }),
  },
  Window: { Close: () => windowCloseFunc() }
}));

describe('DApp Connect Confirmation Page', () => {
  it('has dApp information', async () => {
		render(DappConnectPage);
    expect(await screen.findByText('Foo DApp')).toBeInTheDocument()
    expect(await screen.findByText('http://example.com')).toBeInTheDocument()
    expect(await screen.findByText('Foobar')).toBeInTheDocument()
  });

  it('has list of accounts', async () => {
		render(DappConnectPage);
    expect(await screen.findByText('Choose accounts to connect')).toBeInTheDocument()
    expect(await screen.findByText('No accounts')).toBeInTheDocument()
  });

  it('responds to backend & closes window when confirmation code submitted', async () => {
		render(DappConnectPage);

    // Enter and submit confirmation code
    await userEvent.click(await screen.findByLabelText(/Confirmation code/))
    await userEvent.paste('000000');
    await userEvent.click(screen.getByText('Confirm connection'));

    expect(eventEmitFunc).toHaveBeenCalledOnce()
    expect(windowCloseFunc).toHaveBeenCalledOnce()
  });
});
