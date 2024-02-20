# Integrate KMD Wallet Management Code

- Status: draft
- Deciders: No-Cash-7970
- Date: 2024-02-19

## Context and Problem Statement

Properly managing keys is difficult, tricky, and requires a significant amount cryptography knowledge. It is very easy to mishandle cryptographic keys and has severe consequences. It would be ideal if reliable and trustworthy code could be used to handle this part of the desktop wallet.

## Decision Drivers

- **Security:** Must be able to store keys in a manner that does not allow an unauthorized third-party to gain access to the keys
- **Supported Algorand features:** Should be able support as many Algorand features as possible
- **Development resources:** Would like to keep the amount of development time and effort to a minimum
- **Cross-Platform:** Should be able to be used by most desktop users, most of whom are using either Linux, Mac or Windows operating systems

## Decision Outcome

Chose to integrate the key management parts of Key Management Daemon (KMD) without the session management parts after reviewing the KMD code in the [go-algorand GitHub repository](https://github.com/algorand/go-algorand). KMD's key management code should need very little modification, while the session management code is insufficient for a desktop wallet for end-users in a nondevelopment environment.

**Confidence**: Medium. Because Go will be use to build the desktop wallet, integrating KMD code should be as easy as importing the KMD's Go packages. However, complications caused by importing KMD in parts are possible.

### Positive Consequences

- Easier to support a large number of Algorand's features including signing transactions, multisignatures, and signing programs
- The most important and dangerous part of the desktop wallet would be handled by well-tested coded developed and maintained by [Algorand Technologies](https://algorandtechnologies.com/)
- Saves time and effort
- Allows for the desktop wallet to continue to be cross-platform because KMD is written in Go

### Negative Consequences

- Would require the project to be licensed under [AGPL v3.0](https://choosealicense.com/licenses/agpl-3.0/). Although the MIT license for this project is preferred, AGPL v3.0 is not a problem since the all of the code will be open source anyway.
- For Ledger devices, access to only the first account is supported. To support access to multiple accounts on the Ledger device, the KMD code may need to be modified to be able to do so.

## Links

- Relates to [Build Algorand Desktop Wallet](20231231-build-algorand-desktop-wallet.md)
- [KMD code](https://github.com/algorand/go-algorand/tree/eceed7c0d3df0f412ede27c1aa2b68e0fa21ccab/daemon/kmd)
- [KMD code for managing keys with SQLite](https://github.com/algorand/go-algorand/blob/master/daemon/kmd/wallet/driver/sqlite.go)
- [KMD code for the mechanisms used to encrypt keys](https://github.com/algorand/go-algorand/blob/master/daemon/kmd/wallet/driver/sqlite_crypto.go)
- [Affero General Public License - Wikipedia](https://en.wikipedia.org/wiki/Affero_General_Public_License)
- [Open Source Software Licenses 101: The AGPL License - FOSSA](https://fossa.com/blog/open-source-software-licenses-101-agpl-license/)
- [Understanding the AGPL: The Most Misunderstood License](https://medium.com/swlh/understanding-the-agpl-the-most-misunderstood-license-86fd1fe91275)
