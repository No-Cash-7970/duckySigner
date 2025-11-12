import { describe, it, expect, vi } from 'vitest'
import DuckyConnectProvider from './DuckyConnectProvider'
import type { WalletAccount } from '@txnlab/use-wallet'
import algosdk from 'algosdk'

vi.mock('./ducky-connect', () => {
  const mockSession = {
    id: 'XN/2YQP/uAdTsa3946CvbicxbwZGFPqAdep7g47UyyQ=',
    exp: new Date(1760591204),
    addrs: [
      'RMAZSNHVLAMY5AUWWTSDON4S2HIUV7AYY6MWWEMKYH63YLHAKLZNHQIL3A',
      'EW64GC6F24M7NDSC5R3ES4YUVE3ZXXNMARJHDCCCLIHZU6TBEOC7XRSBG4',
      'GD64YIY3TWGDMCNPP553DZPPR6LDUSFQOIJVFDPPXWEG3FVOJCCDBBHU5A',
    ],
  }
  return ({
    DuckyConnect: class {
      #options: any
      constructor(options: any = {}) {this.#options = options}
      async establishSession() { return mockSession }
      async retrieveSessionData() {
        return this.#options.dapp.name === 'Unestablished Session DApp'
          ? null
          : {...mockSession, dapp: { name: 'Test DApp' }}
      }
    }
  })
})

describe('Ducky Connect use-wallet provider', () => {

  describe('connect()', () => {
    it.skip('connects to DuckySigner by establishing a session', async () => {
      const dcProvider = new DuckyConnectProvider({dapp: { name: 'Test DApp'}})
      const connectedAccts = await dcProvider.connect()
      expect(connectedAccts).toHaveLength(3)
      expect(connectedAccts[0].address).toBe('RMAZSNHVLAMY5AUWWTSDON4S2HIUV7AYY6MWWEMKYH63YLHAKLZNHQIL3A')
      expect(connectedAccts[1].address).toBe('EW64GC6F24M7NDSC5R3ES4YUVE3ZXXNMARJHDCCCLIHZU6TBEOC7XRSBG4')
      expect(connectedAccts[2].address).toBe('GD64YIY3TWGDMCNPP553DZPPR6LDUSFQOIJVFDPPXWEG3FVOJCCDBBHU5A')
    })
  })

  describe('resumeSession()', () => {
    it.skip('returns connected addresses if a session has been established', async () => {
      const dcProvider = new DuckyConnectProvider({dapp: { name: 'Test DApp'}})
      const connectedAccts = await dcProvider.resumeSession() as WalletAccount[]
      expect(connectedAccts).toHaveLength(3)
      expect(connectedAccts[0].address).toBe('RMAZSNHVLAMY5AUWWTSDON4S2HIUV7AYY6MWWEMKYH63YLHAKLZNHQIL3A')
      expect(connectedAccts[1].address).toBe('EW64GC6F24M7NDSC5R3ES4YUVE3ZXXNMARJHDCCCLIHZU6TBEOC7XRSBG4')
      expect(connectedAccts[2].address).toBe('GD64YIY3TWGDMCNPP553DZPPR6LDUSFQOIJVFDPPXWEG3FVOJCCDBBHU5A')
    })

    it.skip('returns `void` (`undefined`) if there is no established session', async () => {
      const dcProvider = new DuckyConnectProvider({dapp: { name: 'Unestablished Session DApp'}})
      expect(typeof (await dcProvider.resumeSession())).toBe('undefined')
    })
  })

  describe('signTransactions()', () => {
    it.skip('signs given transaction', async () => {
      const dcProvider = new DuckyConnectProvider({dapp: { name: 'Test DApp'}})
      const testTxn = algosdk.makePaymentTxnWithSuggestedParamsFromObject({
        sender: 'RMAZSNHVLAMY5AUWWTSDON4S2HIUV7AYY6MWWEMKYH63YLHAKLZNHQIL3A',
        receiver: 'RMAZSNHVLAMY5AUWWTSDON4S2HIUV7AYY6MWWEMKYH63YLHAKLZNHQIL3A',
        amount: 5_000_000, // 5 Algos
        suggestedParams: {
          fee: 1000, // 0.001 Algos
          firstValid: 6000000,
          lastValid: 6001000,
          genesisHash: algosdk.base64ToBytes('SGO1GKSzyE7IEPItTxCByw9x8FmnrCDexi9/cOUJOiI='),
          genesisID: 'testnet-v1.0',
          minFee: 1000,
        }
      })
      const signedTxns = await dcProvider.signTransactions([testTxn])
      expect(signedTxns).toHaveLength(1)
      expect(signedTxns[0]?.length).toBeGreaterThan(0)
    })
  })

})
