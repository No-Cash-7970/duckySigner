import type { CustomProvider, WalletAccount } from '@txnlab/use-wallet'
import algosdk from 'algosdk'
import { DuckyConnect, type ConnectOptions } from './ducky-connect'

/** A wallet provider for DuckySigner to be used with `@txnlab/use-wallet`.
 * See <https://txnlab.gitbook.io/use-wallet/guides/custom-provider> for more information.
 */
export default class DuckyConnectProvider implements CustomProvider {

  #dc: DuckyConnect

  constructor(options: ConnectOptions) {
    this.#dc = new DuckyConnect(options)
  }

  // Required: Connect to wallet and return accounts
  async connect(): Promise<WalletAccount[]> {
    // TODO
    return []
  }

  // Optional: Clean up when disconnecting
  async disconnect(): Promise<void> {
    // TODO
  }

  // Optional: Restore previous session
  async resumeSession(): Promise<WalletAccount[] | void> {
    // TODO
  }

  // Optional: Sign transactions (implement at least one signing method)
  async signTransactions(
    txnGroup: algosdk.Transaction[] | Uint8Array[] | (algosdk.Transaction[] | Uint8Array[])[],
    indexesToSign?: number[]
  ): Promise<(Uint8Array | null)[]> {
    // TODO
    return []
  }

  // // Optional: Sign with ATC-compatible signer
  // async transactionSigner(
  //   txnGroup: algosdk.Transaction[],
  //   indexesToSign: number[]
  // ): Promise<Uint8Array[]> {
  //   // TODO
  //   return []
  // }
}
