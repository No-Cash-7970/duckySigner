import { describe, it, expect, vi } from 'vitest'
import DuckyConnectProvider from './DuckyConnectProvider'
import type { WalletAccount } from '@txnlab/use-wallet'
import algosdk from 'algosdk'

vi.mock('./ducky-connect', () => {
  const mockSessionData = {
    dapp: { name: 'Test DApp' },
    session: {
      id: 'XN/2YQP/uAdTsa3946CvbicxbwZGFPqAdep7g47UyyQ=',
      exp: new Date(1760591204),
      addrs: [
        'RMAZSNHVLAMY5AUWWTSDON4S2HIUV7AYY6MWWEMKYH63YLHAKLZNHQIL3A',
        'EW64GC6F24M7NDSC5R3ES4YUVE3ZXXNMARJHDCCCLIHZU6TBEOC7XRSBG4',
        'GD64YIY3TWGDMCNPP553DZPPR6LDUSFQOIJVFDPPXWEG3FVOJCCDBBHU5A',
      ],
    }
  }
  return ({
    DuckyConnect: vi.fn(class {
      #options: any
      constructor(options: any = {}) {this.#options = options}
      setup = async () => {}
      establishSession = async () => mockSessionData
      retrieveSession = async () => (
        this.#options.dapp.name === 'Unestablished Session DApp' ? null : mockSessionData
      )
      endSession = vi.fn()
      signTransaction = async () => {
        const stxnB64 = 'gqNzaWfEQHOy8+zozpBTp3wOA1ZzANbN2LXeHTUTFre5xg0WpsPiKTm9Eto4Kq+XuutVHvaTMa9v7KxpWB+tZ79iOeCqDgyjdHhuiKNmZWXNA+iiZnbOAnvsNaNnZW6sdGVzdG5ldC12MS4womdoxCBIY7UYpLPITsgQ8i1PEIHLD3HwWaesIN7GL39w5Qk6IqJsds4Ce/Ado3JjdsQgiwGZNPVYGY6ClrTkNzeS0dFK/BjHmWsRisH9vCzgUvKjc25kxCCLAZk09VgZjoKWtOQ3N5LR0Ur8GMeZaxGKwf28LOBS8qR0eXBlo3BheQ=='
        return algosdk.decodeSignedTransaction(algosdk.base64ToBytes(stxnB64))
      }
    })
  })
})

describe('Ducky Connect use-wallet provider', () => {

  describe('connect()', () => {
    it('connects to DuckySigner by establishing a session', async () => {
      const dcProvider = new DuckyConnectProvider({dapp: { name: 'Test DApp'}})
      const connectedAccts = await dcProvider.connect()
      expect(connectedAccts).toHaveLength(3)
      expect(connectedAccts[0].name).toBe('Ducky Account 1')
      expect(connectedAccts[1].name).toBe('Ducky Account 2')
      expect(connectedAccts[2].name).toBe('Ducky Account 3')
      expect(connectedAccts[0].address).toBe('RMAZSNHVLAMY5AUWWTSDON4S2HIUV7AYY6MWWEMKYH63YLHAKLZNHQIL3A')
      expect(connectedAccts[1].address).toBe('EW64GC6F24M7NDSC5R3ES4YUVE3ZXXNMARJHDCCCLIHZU6TBEOC7XRSBG4')
      expect(connectedAccts[2].address).toBe('GD64YIY3TWGDMCNPP553DZPPR6LDUSFQOIJVFDPPXWEG3FVOJCCDBBHU5A')
    })
  })

  describe('resumeSession()', () => {
    it('returns connected addresses if a session has been established', async () => {
      const dcProvider = new DuckyConnectProvider({dapp: { name: 'Test DApp'}})
      const connectedAccts = await dcProvider.resumeSession() as WalletAccount[]
      expect(connectedAccts).toHaveLength(3)
      expect(connectedAccts[0].name).toBe('Ducky Account 1')
      expect(connectedAccts[1].name).toBe('Ducky Account 2')
      expect(connectedAccts[2].name).toBe('Ducky Account 3')
      expect(connectedAccts[0].address).toBe('RMAZSNHVLAMY5AUWWTSDON4S2HIUV7AYY6MWWEMKYH63YLHAKLZNHQIL3A')
      expect(connectedAccts[1].address).toBe('EW64GC6F24M7NDSC5R3ES4YUVE3ZXXNMARJHDCCCLIHZU6TBEOC7XRSBG4')
      expect(connectedAccts[2].address).toBe('GD64YIY3TWGDMCNPP553DZPPR6LDUSFQOIJVFDPPXWEG3FVOJCCDBBHU5A')
    })

    it('returns `void` (`undefined`) if there is no established session', async () => {
      const dcProvider = new DuckyConnectProvider({dapp: { name: 'Unestablished Session DApp'}})
      expect(typeof (await dcProvider.resumeSession())).toBe('undefined')
    })
  })

  describe('signTransactions()', () => {
    const testTxns = [
      // Transaction #1
      algosdk.makePaymentTxnWithSuggestedParamsFromObject({
        sender: 'RMAZSNHVLAMY5AUWWTSDON4S2HIUV7AYY6MWWEMKYH63YLHAKLZNHQIL3A',
        receiver: 'RMAZSNHVLAMY5AUWWTSDON4S2HIUV7AYY6MWWEMKYH63YLHAKLZNHQIL3A',
        amount: 0,
        suggestedParams: {
          fee: 1000, // 0.001 Algos
          firstValid: 6000000,
          lastValid: 6001000,
          genesisHash: algosdk.base64ToBytes('SGO1GKSzyE7IEPItTxCByw9x8FmnrCDexi9/cOUJOiI='),
          genesisID: 'testnet-v1.0',
          minFee: 1000,
        }
      }),
      // Transaction #2
      algosdk.makePaymentTxnWithSuggestedParamsFromObject({
        sender: 'RMAZSNHVLAMY5AUWWTSDON4S2HIUV7AYY6MWWEMKYH63YLHAKLZNHQIL3A',
        receiver: 'RMAZSNHVLAMY5AUWWTSDON4S2HIUV7AYY6MWWEMKYH63YLHAKLZNHQIL3A',
        amount: 1_000_000, // 1 Algo
        suggestedParams: {
          fee: 1000, // 0.001 Algos
          firstValid: 6000000,
          lastValid: 6001000,
          genesisHash: algosdk.base64ToBytes('SGO1GKSzyE7IEPItTxCByw9x8FmnrCDexi9/cOUJOiI='),
          genesisID: 'testnet-v1.0',
          minFee: 1000,
        }
      }),
      // Transaction #3
      algosdk.makePaymentTxnWithSuggestedParamsFromObject({
        sender: 'RMAZSNHVLAMY5AUWWTSDON4S2HIUV7AYY6MWWEMKYH63YLHAKLZNHQIL3A',
        receiver: 'RMAZSNHVLAMY5AUWWTSDON4S2HIUV7AYY6MWWEMKYH63YLHAKLZNHQIL3A',
        amount: 2_000_000, // 2 Algos
        suggestedParams: {
          fee: 1000, // 0.001 Algos
          firstValid: 6000000,
          lastValid: 6001000,
          genesisHash: algosdk.base64ToBytes('SGO1GKSzyE7IEPItTxCByw9x8FmnrCDexi9/cOUJOiI='),
          genesisID: 'testnet-v1.0',
          minFee: 1000,
        }
      })
    ]

    it('can sign one transaction', async () => {
      const dcProvider = new DuckyConnectProvider({dapp: { name: 'Test DApp'}})
      const signedTxns = await dcProvider.signTransactions([testTxns[0]])
      expect(signedTxns).toHaveLength(1)
      expect(signedTxns[0]?.length).toBeGreaterThan(0)
    })

    it('can sign multiple transactions', async () => {
      const dcProvider = new DuckyConnectProvider({dapp: { name: 'Test DApp'}})
      const signedTxns = await dcProvider.signTransactions(testTxns)
      expect(signedTxns).toHaveLength(3)
      expect(signedTxns[0]?.length).toBeGreaterThan(0)
      expect(signedTxns[1]?.length).toBeGreaterThan(0)
      expect(signedTxns[2]?.length).toBeGreaterThan(0)
    })

    it('signs only the transactions specified by the `indexesToSign` argument', async () => {
      const dcProvider = new DuckyConnectProvider({dapp: { name: 'Test DApp'}})
      const signedTxns = await dcProvider.signTransactions(testTxns, [1])
      expect(signedTxns).toHaveLength(3)
      expect(signedTxns[0]).toBeNull()
      expect(signedTxns[1]?.length).toBeGreaterThan(0)
      expect(signedTxns[2]).toBeNull()
    })
  })

})
