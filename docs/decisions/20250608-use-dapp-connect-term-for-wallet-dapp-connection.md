# Use "DApp Connect" Term for Wallet-DApp Connection

- Status: accepted
- Deciders: No-Cash-7970
- Date: 2025-06-10
- Tags: dev-process, docs

## Context and Problem Statement

Multiple vocabulary terms, such as "wallet connection" and "dApp connect", have been used interchangeably and inconsistently to refer to the mechanism for connecting a dApp to the desktop wallet. The inconsistency makes the documentation more confusing and the code harder to understand. A single term needs to be chosen and used consistently throughout the documentation and the code.

A decision on this matter is important because the wallet-dApp connection mechanism is a key standout feature of the desktop wallet. This feature needs a suitable name that can be easily identified and understood by users and readers.

## Decision Drivers

- **Descriptiveness:** The term needs to be succinct and sufficiently describe the wallet-dApp connection mechanism
- **Similarity to other terms:** To prevent confusion, the term should not be too similar to terms that are already used
- **Prior use in project:** It is easier to use the term that has been used most frequently in this project's documentation and code

## Considered Options

- DApp connect
- DApp connection
- Wallet connection
- Wallet connect

## Decision Outcome

Chose to use the term "dApp connect" to refer to the wallet-dApp connection mechanism. The term is short and memorable while being descriptive enough. It is also not too similar to any term already used in blockchain or cryptocurrency. In contrast to the term "wallet connect", "dApp connect" puts the focus on the dApps rather than the wallet. If a shortened term is needed in certain contexts, then the term "connect" can be used. For example, "dApp connect server" can be shortened to "connect server".

The term "dApp connect" may be used as the proper brand name for the wallet-dApp connection mechanism, so the "DApp Connect" or "DAppConnect" proper name variation may eventually become more common.

**Confidence:** Very high

### Positive Consequences

- Consistent use of a clear vocabulary term improves documentation and code
- Less work in the long run

### Negative Consequences

- Older documentation and code that used other terms need to be revisited and edited to use newly decided term
- More work is necessary immediately

## Pros and Cons of the Options

### DApp Connect

- Pro: Short and memorable
- Pro: Distinctive, no similar term is widely used.
- Con: Ambiguous, the term does not describe *what* the dApp is connecting to

### DApp Connection

- Pro: More descriptive and distinct
- Con: A little long
- Con: Ambiguous, the term does not describe *what* the dApp is connecting to

### Wallet Connection

- Pro: More descriptive
- Con: Too close to "wallet connect", which is already widely used by [WalletConnect](https://walletconnect.network/)

### Wallet Connect

- Pro: Short and memorable
- Con: Already used by [WalletConnect](https://walletconnect.network/), which is widely used throughout blockchain and cryptocurrency

## Links

- Relates to [Use \"DApp\" in Prose and \"Dapp\" in Code](20250608-use-dapp-in-prose-and-dapp-in-code.md)
- Relates to [Terms for Parts of DApp Connect](20250609-terms-for-parts-of-dapp-connect.md)
- [WalletConnect](https://walletconnect.network/)
