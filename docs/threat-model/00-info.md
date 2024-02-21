# Threat Model Information

**Application Version:** N/A

**Description:** Ducky Signer is a prototype desktop wallet for the Algorand public blockchain. Being that it is a prototype, Ducky Signer should be simple with limited functionality. The primary purpose of a wallet is to make it easier for the use to safely store the private key(s) to their Algorand accounts while enabling those users to use the private key(s) to cryptographically sign various things.

**Document Owner**: No-Cash-7970

**Participants**: No-Cash-7970

## External Dependencies

### EXTERN-00: Algorand Network

All Algorand wallet account data (except for secrets like the private key) are stored on a public blockchain network. The most notable data is the amount of funds within an account in the form of "Algos" or some "Algorand Standard Asset" (ASA). In addition to the account data, the network stores every Algorand transaction that has occurred on that network's chain. There are three main networks for Algorand: MainNet, TestNet, and BetaNet. MainNet is the definitive network for assets linked to valuable resources outside the network (like money). TestNet and BetaNet are for testing.

### EXTERN-01: User's Computer

TODO: Add description about the user's computer being an external dependency

### EXTERN-[ID_NUMBER]: [Add name of external dependency here]

[Insert description here]

## Entry Points

NOTE: A trust level of 1 is the lowest trust level

### ENTRY-00: Wallet GUI

This is the main method in which the user is supposed to interact with the wallet to manage their key(s). The GUI is intended to be used by people who may have very little technical knowledge about Algorand, computers (in general), and blockchain (in general). It is also intended to be used by people who have higher levels of expertise in those subjects.

**Trust Levels**:

1. Anonymous user
2. Authenticated user

### ENTRY-[ID_NUM]: [Add name of entry point here]

[Insert description here]

**Trust Levels**:

1. [Lowest level of trust]
2. [Next lowest level of trust]
3. [Highest level of trust]

## Exit Points

NOTE: A trust level of 1 is the lowest trust level

### EXIT-00: Wallet GUI

One of the main purposes of the GUI is to display information. However, it is possible it can display *too much* information in the form of error messages and account status(es).

**Trust Levels**:

1. Anonymous user
2. Authenticated user

### EXIT-[ID_NUM]: [Add name of exit point here]

[Insert description here]

**Trust Levels**:

1. [Lowest level of trust]
2. [Next lowest level of trust]
3. [Highest level of trust]

## Assets

### ASSET-00: Account private keys

An account's private key is the most valuable asset a wallet can contain. The private key is needed to transfer funds from the account and interact with the network through smart contracts (also known as "applications"). The private key is usually encoded into the form of a 25-word mnemonic to make it easier for the user to record into some non-digital medium. The purpose of wallet software is to safely store private keys in a way that makes it easier for users to use them to authenticate as their accounts to authorize interactions with the network.

**Trust Levels**:

1. Anonymous user
2. Authenticated user

### ASSET-[ID_NUM]: [Add name of asset here]

[Insert description here]

**Trust Levels**:

1. [Lowest level of trust]
2. [Next lowest level of trust]
3. [Highest level of trust]

## Trust Levels

### TRUST-00: Anonymous User

A user who **has not** provided a set of credentials (e.g. username and password) to access the protected parts of the wallet software that provide sensitive Algorand wallet account information and allow for the user to use account keys.

### TRUST-01: Authenticated user

A user who **has** provided a set of credentials (e.g. username and password) to access the protected parts of the wallet software that provide sensitive Algorand wallet account information.
