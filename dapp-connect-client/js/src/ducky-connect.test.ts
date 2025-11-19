import { describe, it, expect, vi, afterEach } from 'vitest'
import { DuckyConnect, DEFAULT_SERVER_BASE_URL, type StoredSessionInfo } from './ducky-connect'
import algosdk from 'algosdk'

const fetchSpy = vi.spyOn(globalThis, 'fetch')

describe('Ducky Connect class', () => {

  afterEach(() => {
    fetchSpy.mockReset()
    // TODO: Clear local storage
  })

  describe('establishSession()', () => {
    it.skip('creates a new session and returns the session data', async () => {
      // Mock the responses to connect server requests
      fetchSpy.mockImplementation(async url => {
        if (url === `${DEFAULT_SERVER_BASE_URL}/session/init`) {
          return new Response(
            JSON.stringify({
              id: 'K9/uJiX2miksx2Bp+X3L9QQFz6xX+J0icHnGsDSCcm8=',
              code: '88048',
              token: 'v4.local.kotGdyU87V83S8E2qtvnMchNcTjmq50u2C17-mlBDzvVf_bw3xtW0vQQZ-X8jfkDO326nz7_Z9BinJIQre3rQqrJD2BLSM9mzJxCUnjpZKxcgKvr5Bj4pphCanuonR7gwtb03_jYRcfH8PJaOHsRIz24KuhO7GJ8sJS4k-jVMhvuPquqi0VJDAAy7Cn8OMD871mKn7vfE7zYpjjul1AGXgTvd_SapSbYqd3K4PeP3_Y-9_gagK-eXqvoxZ4pNAfcEb4eGP_beR_QjP0X0a7Eq8dwr-bONpW-VzhN3gcj',
              exp: 1760003247,
            }),
            { headers: new Headers({'Content-Type': 'application/json'}) }
          )
        }

        if (url === `${DEFAULT_SERVER_BASE_URL}/session/confirm`) {
          return new Response(
            JSON.stringify({
              id: 'XN/2YQP/uAdTsa3946CvbicxbwZGFPqAdep7g47UyyQ=',
              exp: 1760591204,
              addrs: ['RMAZSNHVLAMY5AUWWTSDON4S2HIUV7AYY6MWWEMKYH63YLHAKLZNHQIL3A'],
            }),
            { headers: new Headers({'Content-Type': 'application/json'}) }
          )
        }

        return new Response
      })

      const dc = new DuckyConnect({dapp: { name: 'Test DApp'}})
      const session = await dc.establishSession()
      expect(session.id).toBe('XN/2YQP/uAdTsa3946CvbicxbwZGFPqAdep7g47UyyQ=')
      expect(Math.floor(session.exp.getTime() / 1000)).toBe(1760591204)
      expect(session.addrs).toBe(['RMAZSNHVLAMY5AUWWTSDON4S2HIUV7AYY6MWWEMKYH63YLHAKLZNHQIL3A'])
      expect(dc.retrieveSession()).not.toBeNull()
    })

    it.skip('throws error when session initialization fails', async () => {
      // Mock the responses to connect server requests
      fetchSpy.mockImplementation(async url => {
        if (url === `${DEFAULT_SERVER_BASE_URL}/session/init`) {
          return new Response(
            JSON.stringify({ name: 'oops',  message: 'Something went wrong!' }),
            { status: 500, headers: new Headers({'Content-Type': 'application/json'}) }
          )
        }

        if (url === `${DEFAULT_SERVER_BASE_URL}/session/confirm`) {
          return new Response(
            JSON.stringify({
              id: 'XN/2YQP/uAdTsa3946CvbicxbwZGFPqAdep7g47UyyQ=',
              exp: 1760591204,
              addrs: ['RMAZSNHVLAMY5AUWWTSDON4S2HIUV7AYY6MWWEMKYH63YLHAKLZNHQIL3A'],
            }),
            { headers: new Headers({'Content-Type': 'application/json'}) }
          )
        }

        return new Response
      })

      const dc = new DuckyConnect({dapp: { name: 'Test DApp'}})
      await expect(() => dc.establishSession()).rejects.toThrowError()
    })

    it.skip('throws error when session confirmation fails', async () => {
      // Mock the responses to connect server requests
      fetchSpy.mockImplementation(async url => {
        if (url === `${DEFAULT_SERVER_BASE_URL}/session/init`) {
          return new Response(
            JSON.stringify({
              id: 'K9/uJiX2miksx2Bp+X3L9QQFz6xX+J0icHnGsDSCcm8=',
              code: '88048',
              token: 'v4.local.kotGdyU87V83S8E2qtvnMchNcTjmq50u2C17-mlBDzvVf_bw3xtW0vQQZ-X8jfkDO326nz7_Z9BinJIQre3rQqrJD2BLSM9mzJxCUnjpZKxcgKvr5Bj4pphCanuonR7gwtb03_jYRcfH8PJaOHsRIz24KuhO7GJ8sJS4k-jVMhvuPquqi0VJDAAy7Cn8OMD871mKn7vfE7zYpjjul1AGXgTvd_SapSbYqd3K4PeP3_Y-9_gagK-eXqvoxZ4pNAfcEb4eGP_beR_QjP0X0a7Eq8dwr-bONpW-VzhN3gcj',
              exp: 1760003247,
            }),
            { headers: new Headers({'Content-Type': 'application/json'}) }
          )
        }

        if (url === `${DEFAULT_SERVER_BASE_URL}/session/confirm`) {
          return new Response(
            JSON.stringify({ name: 'oops',  message: 'Something went wrong!' }),
            { status: 500, headers: new Headers({'Content-Type': 'application/json'}) }
          )
        }

        return new Response
      })

      const dc = new DuckyConnect({dapp: { name: 'Test DApp'}})
      await expect(() => dc.establishSession()).rejects.toThrowError()
    })
  })

  describe('retrieveSession()', () => {
    it.skip('returns stored session information, if it exists', () => {
      const sessionInfoToBeStored: StoredSessionInfo = {
        connectId: '7v/yMHo8iYIvnDvq5ObjgSjTX88/PIdpxkTA+zRM/Xo=',
        session: {
          id: 'XN/2YQP/uAdTsa3946CvbicxbwZGFPqAdep7g47UyyQ=',
          exp: new Date(1760591204 * 1000),
          addrs: ['RMAZSNHVLAMY5AUWWTSDON4S2HIUV7AYY6MWWEMKYH63YLHAKLZNHQIL3A'],
        },
        dapp: { name: 'Test DApp' },
      }

      // TODO: Set stored session info in local storage

      const dc = new DuckyConnect({dapp: {name: ''}})
      const storedSession = dc.retrieveSession()
      expect(storedSession?.session.id).toBe(sessionInfoToBeStored.session.id)
      expect(storedSession?.session.exp).toBe(sessionInfoToBeStored.session.exp)
      expect(storedSession?.session.addrs).toBe(sessionInfoToBeStored.session.addrs)
      expect(storedSession?.dapp.name).toBe(sessionInfoToBeStored.dapp.name)
    })

    it.skip('returns null if there is no stored session information', () => {
      const dc = new DuckyConnect({dapp: {name: ''}})
      expect(dc.retrieveSession()).toBeNull()
    })
  })

  describe('endSession()', () => {
    it.skip('removes stored session after successfully contacting server', async () => {
      // Mock the responses to connect server requests
      fetchSpy.mockImplementation(async url => {
        if (url === `${DEFAULT_SERVER_BASE_URL}/session/end`) {
          return new Response('OK', { headers: new Headers({'Content-Type': 'application/json'}) })
        }
        return new Response
      })

      // TODO: Set stored session info in local storage

      const dc = new DuckyConnect({dapp: { name: 'Test DApp'}})
      await dc.endSession()
      expect(dc.retrieveSession()).toBeNull()
    })

    it.skip('still removes stored session information after contacting server fails', async () => {
      // Mock the responses to connect server requests
      fetchSpy.mockImplementation(async url => {
        if (url === `${DEFAULT_SERVER_BASE_URL}/session/end`) {
          throw Error('Uh oh!')
        }
        return new Response
      })

      // TODO: Set stored session info in local storage

      const dc = new DuckyConnect({dapp: { name: 'Test DApp'}})
      await dc.endSession()
      expect(dc.retrieveSession()).toBeNull()
    })
  })

  describe('signTransaction()', () => {
    it.skip('signs transaction WITHOUT signer address given', async () => {
      // Mock the responses to connect server requests
      fetchSpy.mockImplementation(async url => {
        if (url === `${DEFAULT_SERVER_BASE_URL}/transaction/sign`) {
          return new Response(
            JSON.stringify({
              signed_transaction: 'gqNzaWfEQMSfjRLM8S/j4At47sdxr8GSV+Yy//7Srs9iJlpReFs719ibxEiU+ZIpE2NJ2kJYpvPswnSx+8eIa0Jm6wJ+ZwijdHhuiaNhbXTOAA9CQKNmZWXNA+iiZnbOA1+J/aNnZW6sdGVzdG5ldC12MS4womdoxCBIY7UYpLPITsgQ8i1PEIHLD3HwWaesIN7GL39w5Qk6IqJsds4DX43lo3JjdsQgVTpfwudWgk+SmzvrmbFS1Xh2IAM+amjAWnhX5FsIJzajc25kxCBVOl/C51aCT5KbO+uZsVLVeHYgAz5qaMBaeFfkWwgnNqR0eXBlo3BheQ==',
            }),
            { headers: new Headers({'Content-Type': 'application/json'}) }
          )
        }
        return new Response
      })

      // TODO: Set stored session info in local storage

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
      const dc = new DuckyConnect({dapp: { name: 'Test DApp'}})
      const signedTxn = await dc.signTransaction(testTxn)
      // Check if the correct transaction was signed
      expect(signedTxn.txn.sender).toBe('RMAZSNHVLAMY5AUWWTSDON4S2HIUV7AYY6MWWEMKYH63YLHAKLZNHQIL3A')
      expect(signedTxn.txn.payment?.amount).toBe(5_000_000)
    })

    it.skip('signs transaction WITH signer address given', async () => {
      // Mock the responses to connect server requests
      fetchSpy.mockImplementation(async url => {
        if (url === `${DEFAULT_SERVER_BASE_URL}/transaction/sign`) {
          return new Response(
            JSON.stringify({
              'signed_transaction': 'gqNzaWfEQMSfjRLM8S/j4At47sdxr8GSV+Yy//7Srs9iJlpReFs719ibxEiU+ZIpE2NJ2kJYpvPswnSx+8eIa0Jm6wJ+ZwijdHhuiaNhbXTOAA9CQKNmZWXNA+iiZnbOA1+J/aNnZW6sdGVzdG5ldC12MS4womdoxCBIY7UYpLPITsgQ8i1PEIHLD3HwWaesIN7GL39w5Qk6IqJsds4DX43lo3JjdsQgVTpfwudWgk+SmzvrmbFS1Xh2IAM+amjAWnhX5FsIJzajc25kxCBVOl/C51aCT5KbO+uZsVLVeHYgAz5qaMBaeFfkWwgnNqR0eXBlo3BheQ==',
            }),
            { headers: new Headers({'Content-Type': 'application/json'}) }
          )
        }
        return new Response
      })

      // TODO: Set stored session info in local storage

      const testTxn = algosdk.makePaymentTxnWithSuggestedParamsFromObject({
        sender: 'RMAZSNHVLAMY5AUWWTSDON4S2HIUV7AYY6MWWEMKYH63YLHAKLZNHQIL3A',
        receiver: 'RMAZSNHVLAMY5AUWWTSDON4S2HIUV7AYY6MWWEMKYH63YLHAKLZNHQIL3A',
        amount: 5_000_000, // 5 Algos
        suggestedParams: {
          fee: 1000, // 0.001 Algos
          flatFee: true,
          firstValid: 6000000,
          lastValid: 6001000,
          genesisHash: algosdk.base64ToBytes('SGO1GKSzyE7IEPItTxCByw9x8FmnrCDexi9/cOUJOiI='),
          genesisID: 'testnet-v1.0',
          minFee: 1000,
        }
      })
      const dc = new DuckyConnect({dapp: { name: 'Test DApp'}})
      const signedTxn = await dc.signTransaction(
        testTxn,
        'RMAZSNHVLAMY5AUWWTSDON4S2HIUV7AYY6MWWEMKYH63YLHAKLZNHQIL3A'
      )
      // Check if the correct transaction was signed
      expect(signedTxn.txn.sender).toBe('RMAZSNHVLAMY5AUWWTSDON4S2HIUV7AYY6MWWEMKYH63YLHAKLZNHQIL3A')
      expect(signedTxn.txn.payment?.amount).toBe(5_000_000)
    })

    it.skip('throws error if transaction signing fails', async () => {
      // Mock the responses to connect server requests
      fetchSpy.mockImplementation(async url => {
        if (url === `${DEFAULT_SERVER_BASE_URL}/transaction/sign`) {
          return new Response(
            JSON.stringify({ name: 'oops',  message: 'Something went wrong!' }),
            { status: 500, headers: new Headers({'Content-Type': 'application/json'}) }
          )
        }

        return new Response
      })

      // TODO: Set stored session info in local storage

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

      const dc = new DuckyConnect({dapp: { name: 'Test DApp'}})
      await expect(() => dc.signTransaction(testTxn)).rejects.toThrowError()
    })

    it.skip('throws error if no session has been established', async () => {
      // Mock the responses to connect server requests to prevent an error from being thrown from a
      // request failing
      fetchSpy.mockImplementation(async url => {
        if (url === `${DEFAULT_SERVER_BASE_URL}/transaction/sign`) {
          return new Response(
            JSON.stringify({ signed_transaction: 'some base64 encoded stuff' }),
            { headers: new Headers({'Content-Type': 'application/json'}) }
          )
        }

        return new Response
      })

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

      const dc = new DuckyConnect({dapp: { name: 'Test DApp'}})
      await expect(() => dc.signTransaction(testTxn)).rejects.toThrowError()
    })
  })
})
