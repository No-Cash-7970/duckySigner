import {render, screen} from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import { describe, it, expect, vi } from 'vitest';
import algosdk from 'algosdk';

import TxnSignApprovalPage from './+page.svelte';

// Code (with modifications) from https://vitest.dev/api/vi.html#vi-hoisted
const { eventEmitFunc, windowCloseFunc } = vi.hoisted(() => {
  return {
    eventEmitFunc: vi.fn(),
    windowCloseFunc: vi.fn(),
  }
});
vi.mock('@wailsio/runtime', () => ({
  Events: {
    Emit: eventEmitFunc,
    On: vi.fn().mockImplementation((evtName: string, cb: Function) => {
      const txnB64 = Buffer.from(algosdk.encodeUnsignedTransaction(
        algosdk.makePaymentTxnWithSuggestedParamsFromObject({
          sender: 'EW64GC6F24M7NDSC5R3ES4YUVE3ZXXNMARJHDCCCLIHZU6TBEOC7XRSBG4',
          receiver: 'GD64YIY3TWGDMCNPP553DZPPR6LDUSFQOIJVFDPPXWEG3FVOJCCDBBHU5A',
          amount: 5_000_000,
          suggestedParams: { fee: 1000, firstValid: 6000000, lastValid: 6001000, minFee: 1000 }
        })
      )).toString('base64');
      cb({data: `{"data":{"transaction":"${txnB64}","signer":""}}`});
    }),
  },
  Window: { Close: windowCloseFunc }
}));

describe('Transaction Signing Approval Page', () => {

  it('has heading', async () => {
		render(TxnSignApprovalPage);
    expect(await screen.findByText('Approve Transaction')).toHaveRole('heading');
	});

  it('has transaction information', async () => {
		render(TxnSignApprovalPage);
    expect(await screen.findByText(/EW64GC6F24M7NDSC5R3ES4YUVE3ZXXNMARJHDCCCLIHZU6TBEOC7XRSBG4/))
      .toBeInTheDocument()
  })

  it('responds to backend & closes window when user approves transaction', async() => {
    eventEmitFunc.mockClear()
    windowCloseFunc.mockClear()

		render(TxnSignApprovalPage);
    await userEvent.click(await screen.findByText("Approve"))

    expect(eventEmitFunc).toHaveBeenCalledOnce()
    expect(eventEmitFunc).toHaveBeenCalledWith('txn_sign_response', '[{"approved":true}]')
    expect(windowCloseFunc).toHaveBeenCalledOnce()
  });

  it('responds to backend & closes window when user rejects transaction', async() => {
    eventEmitFunc.mockClear()
    windowCloseFunc.mockClear()

    render(TxnSignApprovalPage);
    await userEvent.click(await screen.findByText("Reject"))

    expect(eventEmitFunc).toHaveBeenCalledOnce()
    expect(eventEmitFunc).toHaveBeenCalledWith('txn_sign_response', '[{"approved":false}]')
    expect(windowCloseFunc).toHaveBeenCalledOnce()
  });

});
