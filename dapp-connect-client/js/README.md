# Ducky Signer Connect JavaScript/TypeScript Library

This is a JavaScript/TypeScript library for connecting to Ducky Signer. Includes a `CustomProvider` to use with [`@txnlab/use-wallet`](https://github.com/TxnLab/use-wallet). Refer to the [`use-wallet` documentation](https://txnlab.gitbook.io/use-wallet/getting-started/configuration) for how to use the custom provider.

## Using Ducky Connect

```js
import { DuckyConnect } from 'ducky-connect'

// Must set up before doing anything else. This also resumes an existing connect session that is
// stored locally.
const duckconn = await (new dc.DuckyConnect({ dapp: { name: 'Test DApp'} })).setup()

// Check if there is an existing connect session by trying to retrieve the session data. If there is
// no session, `retrieveSession()` returns `null`.
let session = await duckconn.retrieveSession()

// If there there is no connect session, connect to the wallet by establishing a session where the
// user must approve the connection within Ducky Signer
let session = (await duckconn.establishSession()).session

// Sign a single transaction. The wallet user is prompted to sign the transaction in Ducky Signer.
const signedTxn = await duckconn.signTransaction(algosdk.makePaymentTxnWithSuggestedParamsFromObject({
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
}))

// Disconnect from Ducky Signer by ending the session
await duckconn.endSession() 
```

## Development

To build the library:

```sh
yarn build
```

To run all tests once:

```sh
yarn test
```

To run the linter:

```sh
yarn lint
```
