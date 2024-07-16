# Use SQLite Database to Store Keys

- Status: accepted
- Deciders: No-Cash-7970
- Date: 2024-01-04
- Tags: backend

## Context and Problem Statement

Which database should be used to store the user's wallet keys?

## Decision Drivers

- **Security:** Must be able to store keys in a manner that does not allow an unauthorized third-party to gain access to the keys
- **Portability:** Prefer a database that allows the user to easily move the database to wherever whenever, which would allow for easier backups.
- **Ease of use:** Prefer a database that does not require a lot of development time and effort to integrate and use

## Decision Outcome

Chose SQLite because is a portable option that is easy to use and can be secure, as demonstrated by Algorand's Key Management Daemon (KMD). Much of KMD's code for managing keys should be able to be ported because it is written in Go. Using KMD's code would make it easier to provide a high enough level of security to keep the user's keys safe.

**Confidence**: High. SQLite is a well-established file-based database with a wide range of support.

## Links

- Relates to [Build Algorand Desktop Wallet](20231231-build-algorand-desktop-wallet.md)
- [KMD code](https://github.com/algorand/go-algorand/tree/eceed7c0d3df0f412ede27c1aa2b68e0fa21ccab/daemon/kmd)
- [KMD code for managing keys with SQLite](https://github.com/algorand/go-algorand/blob/master/daemon/kmd/wallet/driver/sqlite.go)
- [KMD code for the mechanisms used to encrypt keys](https://github.com/algorand/go-algorand/blob/master/daemon/kmd/wallet/driver/sqlite_crypto.go)
