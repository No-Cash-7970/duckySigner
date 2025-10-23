import { CustomProvider, WalletAccount } from '@txnlab/use-wallet'
import { algosdk } from 'algosdk'

export class DappConnectWalletProvider implements CustomProvider {

  // Required: Connect to wallet and return accounts
  async connect(args?: Record<string, any>): Promise<WalletAccount[]> {
    // TODO
    return [new WalletAccount]
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
    return [null]
  }

  // Optional: Sign with ATC-compatible signer
  async transactionSigner?(
    txnGroup: algosdk.Transaction[],
    indexesToSign: number[]
  ): Promise<Uint8Array[]> {
    // TODO
    return [new Uint8Array]
  }
}
