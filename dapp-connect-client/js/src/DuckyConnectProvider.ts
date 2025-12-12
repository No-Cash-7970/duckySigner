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

  /** Connect to wallet and return accounts
   * @returns Wallet accounts that are accessible in this connection
   */
  async connect(): Promise<WalletAccount[]> {
    await this.#dc.setup()
    const storedSessionData = await this.#dc.establishSession()
    return this.#addrsToWalletAccounts(storedSessionData.session.addrs)
  }

  // Optional: Clean up when disconnecting
  async disconnect(): Promise<void> {
    this.#dc.endSession()
  }

  /** Restore previous session
   * @returns Wallet accounts that are accessible in this connection or `void` if there is no session
   *         to resume
   */
  async resumeSession(): Promise<WalletAccount[] | void> {
    await this.#dc.setup()
    const storedSessionData = await this.#dc.retrieveSession()

    if (!storedSessionData) return

    return this.#addrsToWalletAccounts(storedSessionData?.session.addrs)
  }

  /** Sign transactions
   * @param txnGroup Transactions or transaction groups to be signed in the form of a `Transaction`
   *                 object or as encoded bytes
   * @param indexesToSign Indexes of the transaction in `txnGroup` to be signed, if all transaction
   *                      in `txnGroup` are not to be signed
   * @returns All the signed transactions, each signed transaction as encoded bytes. If a
   *          transaction at a particular index was not signed, it will be `null`.
   */
  async signTransactions(
    txnGroup: algosdk.Transaction[] | Uint8Array[] | (algosdk.Transaction[] | Uint8Array[])[],
    indexesToSign?: number[]
  ): Promise<(Uint8Array | null)[]> {
    return await Promise.all(
      txnGroup.map(async (txn, i) => {
        if (indexesToSign && indexesToSign.indexOf(i) < 0) return null

        if (txn instanceof Array) {
          throw new Error('Transaction groups are not supported by this wallet.')
        }

        const stxn = await this.#dc.signTransaction(
          txn instanceof Uint8Array ? algosdk.decodeUnsignedTransaction(txn) : txn
        )

        return algosdk.encodeMsgpack(stxn)
      })
    )
  }

  #addrsToWalletAccounts(addrs: string[]): WalletAccount[] {
    return addrs.map((addr, i) => ({ name: `Ducky Account ${i+1}`, address: addr }))
  }

  // // Optional: Sign with ATC-compatible signer
  // async transactionSigner(
  //   txnGroup: algosdk.Transaction[],
  //   indexesToSign: number[]
  // ): Promise<Uint8Array[]> {
  //   // Do something here
  //   return []
  // }
}
