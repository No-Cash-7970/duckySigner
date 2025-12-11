import { describe, it, expect, vi, beforeEach } from 'vitest'
import 'fake-indexeddb/auto'
import * as dc from './ducky-connect'
import algosdk from 'algosdk'
import { clear as idbClear, get as idbGet, set as idbSet } from 'idb-keyval'
import hawk from 'hawk'

const consoleWarnSpy = vi.spyOn(globalThis.console, 'warn')
const fetchSpy = vi.spyOn(globalThis, 'fetch')
const hawkClientAuthSpy = vi.spyOn(hawk.client, 'authenticate')

describe('Ducky Connect class', () => {

  beforeEach(async () => {
    consoleWarnSpy.mockReset()
    fetchSpy.mockReset()
    hawkClientAuthSpy.mockReset()
    await idbClear()
  })

  describe('establishSession()', () => {
    it('creates a new session and returns the session data', async () => {
      // Mock the responses to connect server requests
      fetchSpy.mockImplementation(async url => {
        if (url === `${dc.DEFAULT_SERVER_BASE_URL}${dc.SESSION_INIT_ENDPOINT}`) {
          return new Response(
            JSON.stringify({
              id: 'K9/uJiX2miksx2Bp+X3L9QQFz6xX+J0icHnGsDSCcm8=',
              code: '88048',
              token: 'v4.local.kotGdyU87V83S8E2qtvnMchNcTjmq50u2C17-mlBDzvVf_bw3xtW0vQQZ-X8jfkDO326nz7_Z9BinJIQre3rQqrJD2BLSM9mzJxCUnjpZKxcgKvr5Bj4pphCanuonR7gwtb03_jYRcfH8PJaOHsRIz24KuhO7GJ8sJS4k-jVMhvuPquqi0VJDAAy7Cn8OMD871mKn7vfE7zYpjjul1AGXgTvd_SapSbYqd3K4PeP3_Y-9_gagK-eXqvoxZ4pNAfcEb4eGP_beR_QjP0X0a7Eq8dwr-bONpW-VzhN3gcj',
              exp: 1760003247,
            }),
            { headers: { 'Content-Type': 'application/json' } }
          )
        }

        if (url === `${dc.DEFAULT_SERVER_BASE_URL}${dc.SESSION_CONFIRM_ENDPOINT}`) {
          hawkClientAuthSpy.mockReturnValue({ headers: {'server-authorization': 'fake_header'}})
          return new Response(
            JSON.stringify({
              id: 'XN/2YQP/uAdTsa3946CvbicxbwZGFPqAdep7g47UyyQ=',
              exp: 1760591204,
              addrs: ['RMAZSNHVLAMY5AUWWTSDON4S2HIUV7AYY6MWWEMKYH63YLHAKLZNHQIL3A'],
            }),
            {
              headers: {
                'Content-Type': 'application/json',
                'Server-Authorization': 'some_faked_auth_header',
              }
            }
          )
        }

        return new Response
      })

      const confirmCodeDisplayFn = vi.fn()
      const duckconn = await (new dc.DuckyConnect({
        dapp: { name: 'Test DApp'},
        confirmCodeDisplayFn
      })).init()
      const session = await duckconn.establishSession()

      expect(confirmCodeDisplayFn).toHaveBeenCalledOnce()
      expect(session.id).toBe('XN/2YQP/uAdTsa3946CvbicxbwZGFPqAdep7g47UyyQ=')
      expect(Math.floor(session.exp.getTime() / 1000)).toBe(1760591204)
      expect(session.addrs).toStrictEqual(['RMAZSNHVLAMY5AUWWTSDON4S2HIUV7AYY6MWWEMKYH63YLHAKLZNHQIL3A'])
      expect(await duckconn.retrieveSession()).not.toBeNull()
    })

    it('throws error when session initialization fails', async () => {
      // Mock the responses to connect server requests
      fetchSpy.mockImplementation(async url => {
        if (url === `${dc.DEFAULT_SERVER_BASE_URL}${dc.SESSION_INIT_ENDPOINT}`) {
          return new Response(
            JSON.stringify({ name: 'oops',  message: 'Something went wrong!' }),
            { status: 500, headers: { 'Content-Type': 'application/json' } }
          )
        }

        if (url === `${dc.DEFAULT_SERVER_BASE_URL}${dc.SESSION_CONFIRM_ENDPOINT}`) {
          return new Response(
            JSON.stringify({
              id: 'XN/2YQP/uAdTsa3946CvbicxbwZGFPqAdep7g47UyyQ=',
              exp: 1760591204,
              addrs: ['RMAZSNHVLAMY5AUWWTSDON4S2HIUV7AYY6MWWEMKYH63YLHAKLZNHQIL3A'],
            }),
            { headers: { 'Content-Type': 'application/json' } }
          )
        }

        return new Response
      })

      const confirmCodeDisplayFn = vi.fn()
      const duckconn = await (new dc.DuckyConnect({
        dapp: { name: 'Test DApp'},
        confirmCodeDisplayFn
      })).init()

      await expect(() => duckconn.establishSession()).rejects.toThrowError()
      expect(confirmCodeDisplayFn).not.toBeCalled()
    })

    it('throws error when session confirmation fails', async () => {
      // Mock the responses to connect server requests
      fetchSpy.mockImplementation(async url => {
        if (url === `${dc.DEFAULT_SERVER_BASE_URL}${dc.SESSION_INIT_ENDPOINT}`) {
          return new Response(
            JSON.stringify({
              id: 'K9/uJiX2miksx2Bp+X3L9QQFz6xX+J0icHnGsDSCcm8=',
              code: '88048',
              token: 'v4.local.kotGdyU87V83S8E2qtvnMchNcTjmq50u2C17-mlBDzvVf_bw3xtW0vQQZ-X8jfkDO326nz7_Z9BinJIQre3rQqrJD2BLSM9mzJxCUnjpZKxcgKvr5Bj4pphCanuonR7gwtb03_jYRcfH8PJaOHsRIz24KuhO7GJ8sJS4k-jVMhvuPquqi0VJDAAy7Cn8OMD871mKn7vfE7zYpjjul1AGXgTvd_SapSbYqd3K4PeP3_Y-9_gagK-eXqvoxZ4pNAfcEb4eGP_beR_QjP0X0a7Eq8dwr-bONpW-VzhN3gcj',
              exp: 1760003247,
            }),
            { headers: { 'Content-Type': 'application/json' } }
          )
        }

        if (url === `${dc.DEFAULT_SERVER_BASE_URL}${dc.SESSION_CONFIRM_ENDPOINT}`) {
          return new Response(
            JSON.stringify({ name: 'oops',  message: 'Something went wrong!' }),
            { status: 500, headers: { 'Content-Type': 'application/json' } }
          )
        }

        return new Response
      })

      const confirmCodeDisplayFn = vi.fn()
      const duckconn = await (new dc.DuckyConnect({
        dapp: { name: 'Test DApp'},
        confirmCodeDisplayFn
      })).init()

      await expect(() => duckconn.establishSession()).rejects.toThrowError()
      expect(confirmCodeDisplayFn).toHaveBeenCalledOnce()
    })
  })

  describe('retrieveSession()', () => {
    it('returns stored session information, if it exists', async () => {
      const sessionInfoToBeStored: dc.StoredSessionInfo = {
        connectId: '7v/yMHo8iYIvnDvq5ObjgSjTX88/PIdpxkTA+zRM/Xo=',
        session: {
          id: 'XN/2YQP/uAdTsa3946CvbicxbwZGFPqAdep7g47UyyQ=',
          exp: new Date(1760591204 * 1000),
          addrs: ['RMAZSNHVLAMY5AUWWTSDON4S2HIUV7AYY6MWWEMKYH63YLHAKLZNHQIL3A'],
        },
        dapp: { name: 'Test DApp' },
        serverURL: dc.DEFAULT_SERVER_BASE_URL,
      }
      // Put session data into storage
      idbSet(dc.DEFAULT_SESSION_DATA_NAME, sessionInfoToBeStored)

      const duckconn = await (new dc.DuckyConnect({dapp: {name: ''}})).init()
      const storedSession = await duckconn.retrieveSession()

      expect(storedSession?.connectId).toBe(sessionInfoToBeStored.connectId)
      expect(storedSession?.session.id).toBe(sessionInfoToBeStored.session.id)
      expect(storedSession?.session.exp).toStrictEqual(sessionInfoToBeStored.session.exp)
      expect(storedSession?.session.addrs).toStrictEqual(sessionInfoToBeStored.session.addrs)
      expect(storedSession?.dapp.name).toBe(sessionInfoToBeStored.dapp.name)
    })

    it('returns null if there is no stored session information', async () => {
      const duckconn = await (new dc.DuckyConnect({dapp: {name: ''}})).init()
      expect(await duckconn.retrieveSession()).toBeNull()
    })
  })

  describe('endSession()', () => {
    const sessionInfoToBeStored: dc.StoredSessionInfo = {
      connectId: '7v/yMHo8iYIvnDvq5ObjgSjTX88/PIdpxkTA+zRM/Xo=',
      session: {
        id: 'XN/2YQP/uAdTsa3946CvbicxbwZGFPqAdep7g47UyyQ=',
        exp: new Date(1760591204 * 1000),
        addrs: ['RMAZSNHVLAMY5AUWWTSDON4S2HIUV7AYY6MWWEMKYH63YLHAKLZNHQIL3A'],
      },
      dapp: { name: 'Test DApp' },
      serverURL: dc.DEFAULT_SERVER_BASE_URL,
    }

    it('removes stored session after successfully contacting server', async () => {
      // Mock the responses to connect server requests
      fetchSpy.mockImplementation(async url => {
        if (url === `${dc.DEFAULT_SERVER_BASE_URL}${dc.SESSION_END_ENDPOINT}`) {
          return new Response('OK', { headers: { 'Content-Type': 'application/json' } })
        }
        return new Response
      })
      // Put connect ("dapp") key pair into storage
      idbSet(dc.DEFAULT_CONNECT_KEY_PAIR_NAME,
        await globalThis.crypto.subtle.generateKey(dc.KEY_ALGORITHM, false, ['deriveBits'])
      )
      // Put session data into storage
      idbSet(dc.DEFAULT_SESSION_DATA_NAME, sessionInfoToBeStored)

      const duckconn = await (new dc.DuckyConnect({dapp: { name: 'Test DApp'}})).init()
      await duckconn.endSession()

      expect(await duckconn.retrieveSession()).toBeNull()
      expect(fetchSpy).toHaveBeenCalledOnce()
    })

    it('still removes stored session information after contacting server fails', async () => {
      // Silence the console warning output that will appear when running the test
      consoleWarnSpy.mockImplementation(() => {})

      // Mock the responses to connect server requests
      fetchSpy.mockImplementation(async url => {
        if (url === `${dc.DEFAULT_SERVER_BASE_URL}${dc.SESSION_END_ENDPOINT}`) {
          throw Error('Uh oh!')
        }
        return new Response
      })
      // Put connect ("dapp") key pair into storage
      idbSet(dc.DEFAULT_CONNECT_KEY_PAIR_NAME,
        await globalThis.crypto.subtle.generateKey(dc.KEY_ALGORITHM, false, ['deriveBits'])
      )
      // Put session data into storage
      idbSet(dc.DEFAULT_SESSION_DATA_NAME, sessionInfoToBeStored)

      const duckconn = await (new dc.DuckyConnect({dapp: { name: 'Test DApp'}})).init()
      await duckconn.endSession()

      expect(await duckconn.retrieveSession()).toBeNull()
      expect(consoleWarnSpy).toHaveBeenCalledOnce()
    })

    it('does not fail if there is no session', async () => {
      const duckconn = await (new dc.DuckyConnect({dapp: { name: 'Test DApp'}})).init()
      await duckconn.endSession()

      expect(await duckconn.retrieveSession()).toBeNull()
    })

    it('does not attempt to contact the server if specified not to do so', async () => {
      // Mock the responses to connect server requests
      fetchSpy.mockImplementation(async url => {
        if (url === `${dc.DEFAULT_SERVER_BASE_URL}${dc.SESSION_END_ENDPOINT}`) {
          return new Response('OK', { headers: { 'Content-Type': 'application/json' } })
        }
        return new Response
      })
      // Put connect ("dapp") key pair into storage
      idbSet(dc.DEFAULT_CONNECT_KEY_PAIR_NAME,
        await globalThis.crypto.subtle.generateKey(dc.KEY_ALGORITHM, false, ['deriveBits'])
      )
      // Put session data into storage
      idbSet(dc.DEFAULT_SESSION_DATA_NAME, sessionInfoToBeStored)

      const duckconn = await (new dc.DuckyConnect({dapp: { name: 'Test DApp'}})).init()
      await duckconn.endSession(false) // Specify not to contact the server

      expect(await duckconn.retrieveSession()).toBeNull()
      expect(fetchSpy).not.toHaveBeenCalled()
    })
  })

  describe('refreshConnectKeyPair()', () => {
    const sessionInfoToBeStored: dc.StoredSessionInfo = {
      connectId: '7v/yMHo8iYIvnDvq5ObjgSjTX88/PIdpxkTA+zRM/Xo=',
      session: {
        id: 'XN/2YQP/uAdTsa3946CvbicxbwZGFPqAdep7g47UyyQ=',
        exp: new Date(1760591204 * 1000),
        addrs: ['RMAZSNHVLAMY5AUWWTSDON4S2HIUV7AYY6MWWEMKYH63YLHAKLZNHQIL3A'],
      },
      dapp: { name: 'Test DApp' },
      serverURL: dc.DEFAULT_SERVER_BASE_URL,
    }

    beforeEach(() => {
      // Mock the responses to connect server requests
      fetchSpy.mockImplementation(async url => {
        if (url === `${dc.DEFAULT_SERVER_BASE_URL}${dc.SESSION_END_ENDPOINT}`) {
          return new Response('OK', { headers: { 'Content-Type': 'application/json' } })
        }
        return new Response
      })
    })

    it('replaces the key pair in storage with a newly generated pair', async () => {
      // Put session data into storage
      idbSet(dc.DEFAULT_SESSION_DATA_NAME, sessionInfoToBeStored)
      // Put connect ("dapp") key pair into storage
      idbSet(dc.DEFAULT_CONNECT_KEY_PAIR_NAME,
        await globalThis.crypto.subtle.generateKey(dc.KEY_ALGORITHM, false, ['deriveBits'])
      )

      const duckconn = await (new dc.DuckyConnect({dapp: { name: 'Test DApp'}})).init()
      const oldKeyPair = await idbGet<CryptoKeyPair>(dc.DEFAULT_CONNECT_KEY_PAIR_NAME)
      await duckconn.refreshConnectKeyPair()
      const newKeyPair = await idbGet<CryptoKeyPair>(dc.DEFAULT_CONNECT_KEY_PAIR_NAME)

      expect(oldKeyPair?.publicKey).not.toBe(newKeyPair?.publicKey)
      expect(await duckconn.retrieveSession()).toBeNull()
      expect(fetchSpy).toHaveBeenCalledOnce()
    })

    it('does not attempt to contact the server if specified not to do so', async () => {
      // Put session data into storage
      idbSet(dc.DEFAULT_SESSION_DATA_NAME, sessionInfoToBeStored)
      // Put connect ("dapp") key pair into storage
      idbSet(dc.DEFAULT_CONNECT_KEY_PAIR_NAME,
        await globalThis.crypto.subtle.generateKey(dc.KEY_ALGORITHM, false, ['deriveBits'])
      )

      const duckconn = await (new dc.DuckyConnect({dapp: { name: 'Test DApp'}})).init()
      const oldKeyPair = await idbGet<CryptoKeyPair>(dc.DEFAULT_CONNECT_KEY_PAIR_NAME)
      await duckconn.refreshConnectKeyPair(false)
      const newKeyPair = await idbGet<CryptoKeyPair>(dc.DEFAULT_CONNECT_KEY_PAIR_NAME)

      expect(oldKeyPair?.publicKey).not.toBe(newKeyPair?.publicKey)
      expect(await duckconn.retrieveSession()).toBeNull()
      expect(fetchSpy).not.toHaveBeenCalled()
    })
  })

  describe('signTransaction()', () => {
    beforeEach(async () => {
      // Put session data into storage
      idbSet(dc.DEFAULT_SESSION_DATA_NAME, {
        connectId: '7v/yMHo8iYIvnDvq5ObjgSjTX88/PIdpxkTA+zRM/Xo=',
        session: {
          id: 'XN/2YQP/uAdTsa3946CvbicxbwZGFPqAdep7g47UyyQ=',
          exp: new Date(1760591204 * 1000),
          addrs: ['RMAZSNHVLAMY5AUWWTSDON4S2HIUV7AYY6MWWEMKYH63YLHAKLZNHQIL3A'],
        },
        dapp: { name: 'Test DApp' },
        serverURL: dc.DEFAULT_SERVER_BASE_URL,
      })
      // Put connect ("dapp") key pair into storage
      idbSet(dc.DEFAULT_CONNECT_KEY_PAIR_NAME,
        await globalThis.crypto.subtle.generateKey(dc.KEY_ALGORITHM, false, ['deriveBits'])
      )
    })

    it('signs transaction WITHOUT signer address given', async () => {
      // Mock the responses to connect server requests
      fetchSpy.mockImplementation(async url => {
        hawkClientAuthSpy.mockReturnValue({ headers: {'server-authorization': 'fake_header'}})
        if (url === `${dc.DEFAULT_SERVER_BASE_URL}${dc.SIGN_TXN_ENDPOINT}`) {
          return new Response(
            JSON.stringify({
              signed_transaction: 'gqNzaWfEQHOy8+zozpBTp3wOA1ZzANbN2LXeHTUTFre5xg0WpsPiKTm9Eto4Kq+XuutVHvaTMa9v7KxpWB+tZ79iOeCqDgyjdHhuiKNmZWXNA+iiZnbOAnvsNaNnZW6sdGVzdG5ldC12MS4womdoxCBIY7UYpLPITsgQ8i1PEIHLD3HwWaesIN7GL39w5Qk6IqJsds4Ce/Ado3JjdsQgiwGZNPVYGY6ClrTkNzeS0dFK/BjHmWsRisH9vCzgUvKjc25kxCCLAZk09VgZjoKWtOQ3N5LR0Ur8GMeZaxGKwf28LOBS8qR0eXBlo3BheQ==',
            }),
            {
              headers: {
                'Content-Type': 'application/json',
                'Server-Authorization': 'some_faked_auth_header',
              }
            }
          )
        }
        return new Response
      })

      const testTxn = algosdk.makePaymentTxnWithSuggestedParamsFromObject({
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
      })
      const duckconn = await (new dc.DuckyConnect({dapp: { name: 'Test DApp'}})).init()
      const signedTxn = await duckconn.signTransaction(testTxn, '', vi.fn())

      // Check if the correct transaction was signed
      expect(signedTxn.txn.sender.toString()).toBe(testTxn.sender.toString())
      expect(signedTxn.txn.payment?.amount).toBe(0n)
    })

    it('signs transaction WITH signer address given', async () => {
      // Mock the responses to connect server requests
      fetchSpy.mockImplementation(async url => {
        hawkClientAuthSpy.mockReturnValue({ headers: {'server-authorization': 'fake_header'}})
        if (url === `${dc.DEFAULT_SERVER_BASE_URL}${dc.SIGN_TXN_ENDPOINT}`) {
          return new Response(
            JSON.stringify({
              signed_transaction: 'gqNzaWfEQHOy8+zozpBTp3wOA1ZzANbN2LXeHTUTFre5xg0WpsPiKTm9Eto4Kq+XuutVHvaTMa9v7KxpWB+tZ79iOeCqDgyjdHhuiKNmZWXNA+iiZnbOAnvsNaNnZW6sdGVzdG5ldC12MS4womdoxCBIY7UYpLPITsgQ8i1PEIHLD3HwWaesIN7GL39w5Qk6IqJsds4Ce/Ado3JjdsQgiwGZNPVYGY6ClrTkNzeS0dFK/BjHmWsRisH9vCzgUvKjc25kxCCLAZk09VgZjoKWtOQ3N5LR0Ur8GMeZaxGKwf28LOBS8qR0eXBlo3BheQ==',
            }),
            {
              headers: {
                'Content-Type': 'application/json',
                'Server-Authorization': 'some_faked_auth_header',
              }
            }
          )
        }
        return new Response
      })

      const testTxn = algosdk.makePaymentTxnWithSuggestedParamsFromObject({
        sender: 'RMAZSNHVLAMY5AUWWTSDON4S2HIUV7AYY6MWWEMKYH63YLHAKLZNHQIL3A',
        receiver: 'RMAZSNHVLAMY5AUWWTSDON4S2HIUV7AYY6MWWEMKYH63YLHAKLZNHQIL3A',
        amount: 0,
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
      const duckconn = await (new dc.DuckyConnect({dapp: { name: 'Test DApp'}})).init()
      const signedTxn = await duckconn.signTransaction(
        testTxn,
        'VCMJKWOY5P5P7SKMZFFOCEROPJCZOTIJMNIYNUCKH7LRO45JMJP6UYBIJA',
        vi.fn(),
      )

      // Check if the correct transaction was signed
      expect(signedTxn.txn.sender.toString()).toBe(testTxn.sender.toString())
      expect(signedTxn.txn.payment?.amount).toBe(0n)
    })

    it('runs user prompt function when starting to wait for response', async () => {
      // Mock the responses to connect server requests
      fetchSpy.mockImplementation(async url => {
        hawkClientAuthSpy.mockReturnValue({ headers: {'server-authorization': 'fake_header'}})
        if (url === `${dc.DEFAULT_SERVER_BASE_URL}${dc.SIGN_TXN_ENDPOINT}`) {
          return new Response(
            JSON.stringify({
              signed_transaction: 'gqNzaWfEQHOy8+zozpBTp3wOA1ZzANbN2LXeHTUTFre5xg0WpsPiKTm9Eto4Kq+XuutVHvaTMa9v7KxpWB+tZ79iOeCqDgyjdHhuiKNmZWXNA+iiZnbOAnvsNaNnZW6sdGVzdG5ldC12MS4womdoxCBIY7UYpLPITsgQ8i1PEIHLD3HwWaesIN7GL39w5Qk6IqJsds4Ce/Ado3JjdsQgiwGZNPVYGY6ClrTkNzeS0dFK/BjHmWsRisH9vCzgUvKjc25kxCCLAZk09VgZjoKWtOQ3N5LR0Ur8GMeZaxGKwf28LOBS8qR0eXBlo3BheQ==',
            }),
            {
              headers: {
                'Content-Type': 'application/json',
                'Server-Authorization': 'some_faked_auth_header',
              }
            }
          )
        }
        return new Response
      })

      const testTxn = algosdk.makePaymentTxnWithSuggestedParamsFromObject({
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
      })
      const duckconn = await (new dc.DuckyConnect({dapp: { name: 'Test DApp'}})).init()
      const promptUserFn = vi.fn()
      const signedTxn = await duckconn.signTransaction(testTxn, '', promptUserFn)

      // Check if the correct transaction was signed
      expect(signedTxn.txn.sender.toString()).toBe(testTxn.sender.toString())
      expect(signedTxn.txn.payment?.amount).toBe(0n)
      expect(promptUserFn).toHaveBeenCalledOnce()
    })

    it('throws error if transaction signing fails', async () => {
      // Mock the responses to connect server requests
      fetchSpy.mockImplementation(async url => {
        hawkClientAuthSpy.mockReturnValue({ headers: {'server-authorization': 'fake_header'}})
        if (url === `${dc.DEFAULT_SERVER_BASE_URL}${dc.SIGN_TXN_ENDPOINT}`) {
          return new Response(
            JSON.stringify({ name: 'oops',  message: 'Something went wrong!' }),
            {
              status: 500,
              headers: {
                'Content-Type': 'application/json',
                'Server-Authorization': 'some_faked_auth_header',
              }
            }
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
      const duckconn = await (new dc.DuckyConnect({dapp: { name: 'Test DApp'}})).init()

      await expect(() => duckconn.signTransaction(testTxn)).rejects.toThrowError()
    })

    it('throws error if no session has been established', async () => {
      // Mock the responses to connect server requests to prevent an error from being thrown from a
      // request failing
      fetchSpy.mockImplementation(async url => {
        hawkClientAuthSpy.mockReturnValue({ headers: {'server-authorization': 'fake_header'}})
        if (url === `${dc.DEFAULT_SERVER_BASE_URL}${dc.SIGN_TXN_ENDPOINT}`) {
          return new Response(
            JSON.stringify({ signed_transaction: 'some base64 encoded stuff' }),
            {
              headers: {
                'Content-Type': 'application/json',
                'Server-Authorization': 'some_faked_auth_header',
              }
            }
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
      const duckconn = await (new dc.DuckyConnect({dapp: { name: 'Test DApp'}})).init()

      await expect(() => duckconn.signTransaction(testTxn)).rejects.toThrowError()
    })
  })
})
