# Use Local Server to Connect to DApps

- Status: accepted
- Deciders: No-Cash-7970
- Date: 2024-01-02
- Tag: backend, wallet-connection

## Context and Problem Statement

The Algorand desktop wallet needs a way to interact with DApps (**D**ecentralized **Apps**) to be useful. In Algorand, the one of the main reasons for using wallet software is to connect to DApps.

## Decision Drivers

- **Protection of user keys:** Connected DApps must not be able to access the user's wallet keys. An unauthorized third-party must not be able to access the user's keys.
- **Spoofing and fraud prevention:** An unauthorized third-party should not be able to intercept communications and present themselves as something or someone they are not
- **Stability of connection between DApp and wallet app:** The connection should be reliable for the DApp and the wallet app to communicate
- **Difficulty for DApps to integrate:** Prefer a method that makes it easy for DApps to connect to the wallet app
- **User privacy:** User's information should not be sent to the DApp or some third-party. All communications should be contained within the user's machine.
- **Reasonable usage of computer resources:** Must not consume an "unreasonable" amount of the user's computing resources (RAM, storage space, CPU, etc). The exact definition of "unreasonable" has yet to be determined.

## Considered Options

- Local server REST API
- WalletConnect
- A "DApp store"
- Just don't

## Decision Outcome

Chose to use a local server REST API because it should not consume much of the user's computing resources while providing a stable and private connection to DApps. Additional well-known protection and security measures, such as SSL/TLS, _may_ be used.

**Confidence**: Low. A local server may not be able to be secure enough. May also be an issue for users with machines that have a strict firewall.

## Pros and Cons of the Options

### Local Server REST API

An HTTP(S) server integrated into the wallet app that serves only on `localhost` or `127.0.0.1`, which should not be exposed to anywhere outside the user's machine. DApps in the browser or installed on the user's machine can connect and interact with this local server through REST API requests. This is a method that Algorand's Key Management Daemon (KMD) uses to securely connect and interact with other services (which may be DApps).

- Pro: Entirely on the user's machine
- Pro: Almost every kind of software that can connect to the internet is capable of doing REST API calls
- Pro: HTTP(S) server connections tend to be reliable and widely supported
- Con: Firewall on user's machine may cause problems
- Con: May not be secure enough, especially in the scenario where malicious software is installed on the user's machine
- Con: Requires the development and maintenance of a library for DApps to easily connect and interact with the local server.

### WalletConnect

WalletConnect is the most widespread method for wallets to connect to DApps.

- Pro: The most well-established method for connecting to wallets
- Pro: If there is a way to integrate into a desktop app, it should require only a little time and effort to get working
- Con: Seems to be no support for connecting to a wallet that is not a mobile wallet
- Con: Connection is managed a third-party closed-source system
- Con: There have been complaints about WalletConnect connections not being reliable

### DApp store

A type of "app store" for DApps, which would be integrated into the desktop wallet app. This would be like MyEtherWallet's "DApps Center" where MyEtherWallet only interacts with a list of registered DApps.

- Pro: Could provide a nice experience for the user
- Con: Would take a large amount of time and effort to build and maintain
- Con: DApp developers would most likely be unwilling spend the time and effort to integrate their DApps into the desktop wallet app

### Just Don't

Do not bother trying to get the desktop wallet to interact with DApps. The desktop wallet would only be for storing keys and looking up the statuses of wallet accounts. Many desktop wallets go with this option.

- Pro: Easiest and most secure option
- Con: If the wallet app cannot connect to DApps, it is not very useful

## Links

- Relates to [Build Algorand Desktop Wallet](20231231-build-algorand-desktop-wallet.md)
- [KMD REST API](https://developer.algorand.org/docs/rest-apis/kmd/)
- [WalletConnect](https://walletconnect.com/)
- [How DApps work with MyEtherWallet](https://www.myetherwallet.com/how-it-works#dapps)
