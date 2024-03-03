# Threat Model Information

**Application Version:** N/A

**Description:** Ducky Signer is a prototype desktop wallet for the Algorand public blockchain. Being that it is a prototype, Ducky Signer should be simple with limited functionality. The goal of the prototype is to explore the technical feasibility of a secure and user-friendly desktop wallet for Algorand. The primary purpose of a wallet is to make it easier for the use to safely store the private key(s) to their Algorand accounts while enabling those users to use the private key(s) to cryptographically sign various things using their desktop computer.

**Document Owner**: No-Cash-7970

**Participants**: No-Cash-7970

<!-- omit in toc -->
## Table of Contents

- [Threat Model Information](#threat-model-information)
  - [External Dependencies](#external-dependencies)
    - [EXTERN-00: Algorand network](#extern-00-algorand-network)
    - [EXTERN-01: User's machine](#extern-01-users-machine)
    - [EXTERN-02: Algorand node](#extern-02-algorand-node)
    - [EXTERN-03: Hardware wallet device](#extern-03-hardware-wallet-device)
  - [Trust Levels](#trust-levels)
    - [TRUST-00: Anonymous wallet user](#trust-00-anonymous-wallet-user)
    - [TRUST-01: Authenticated wallet user](#trust-01-authenticated-wallet-user)
    - [TRUST-02: DApp](#trust-02-dapp)
    - [TRUST-03: Ledger device owner](#trust-03-ledger-device-owner)
  - [Entry Points](#entry-points)
    - [ENTRY-00: Wallet GUI](#entry-00-wallet-gui)
    - [ENTRY-{{ID\_NUM}}: {{Add name of entry point here}}](#entry-id_num-add-name-of-entry-point-here)
  - [Exit Points](#exit-points)
    - [EXIT-00: Wallet GUI](#exit-00-wallet-gui)
    - [EXIT-{{ID\_NUM}}: {{Add name of exit point here}}](#exit-id_num-add-name-of-exit-point-here)
  - [Assets](#assets)
    - [ASSET-00: Account private keys](#asset-00-account-private-keys)
    - [ASSET-{{ID\_NUM}}: {{Add name of asset here}}](#asset-id_num-add-name-of-asset-here)

## External Dependencies

### EXTERN-00: Algorand network

All data regarding an Algorand wallet account (except for secrets like the private key) is stored on a public blockchain network. The most notable datum is the amount of funds within an account in the form of Algos or some Algorand Standard Asset (ASA). In addition to the account data, the network stores every successful Algorand transaction that has occurred on that network's chain. There are three main networks for Algorand: MainNet, TestNet and BetaNet. MainNet is the definitive network for assets linked to valuable resources in the real world (like money). TestNet and BetaNet are for testing.

### EXTERN-01: User's machine

The user's machine is most likely to be a laptop or desktop computer, but it could also be a type of tablet computer. The "user's machine" not only refers to the physical machine owned by the user, but it also refers to the machine's "soft" components such as memory, operating system and file system.

### EXTERN-02: Algorand node

An Algorand node is a special type of server that is required to connect to and communicate with the [Algorand Network](#extern-00-algorand-network). Oftentimes, a node runs 24/7 like many servers on the internet. However, a node can be a small machine, such as a Raspberry Pi, in a local network. Alternatively, a cloud server can be used to create and host a node remotely. Node services, such as [Nodely (AlgoNode)](https://nodely.io/), are commonly used in many projects in the Algorand ecosystem. It is also possible for a user to setup a node on a machine not dedicated for running a node, such as a home or work computer. The node in this instance would not be running 24/7 and would have stale data after being turned off.

### EXTERN-03: Hardware wallet device

A hardware wallet device allows a user to use their wallet account keys to sign transactions without exposing the keys to any system outside the device. Currently, the only hardware wallet that Algorand supports is Ledger. All Ledger devices can connect to a computer or mobile device using USB, but some models can connect using Bluetooth. It is possible for multiple hardware wallets to be connected to the desktop wallet at the same time.

## Trust Levels

The trust levels listed in this section are not in any particular order.

### TRUST-00: Anonymous wallet user

An anonymous wallet user is a user, typically a human, who **has not** authenticated themselves by provided a set of credentials (e.g. username and password) to access the protected parts of the desktop wallet that provide sensitive Algorand wallet account information and allow for the user to use account keys. It is assumed that an anonymous user has access to the machine and the desktop wallet installed on it. It is possible for this type of user to not be the owner of the private keys stored within the desktop wallet or be the owner of the machine itself. It may be possible for there to be more than one anonymous user at the same time.

### TRUST-01: Authenticated wallet user

An authenticated wallet user is a user, typically a human, who **has** somehow authenticated themselves by provided a set of credentials (e.g. username and password) to access the protected parts of the desktop wallet that provide sensitive Algorand wallet account information. Typically, there should be at most one authenticated wallet user accessing the protected parts of the desktop wallet at any given time.

### TRUST-02: DApp

For this document, a dApp ("decentralized" application) is simply software that uses the Algorand blockchain in some manner. The degree of centralization of the dApp software and its functionality is irrelevant. Consequently, a "dApp" in the context of this document does not need to be decentralized, contrary to other commonly accepted definitions of "dApp." DApps are typically web applications run using a web browser. However, with a desktop wallet that does not depend on a web browser, it is possible for a dApp to be software that is installed on the user's machine that also does not depend on a web browser. This makes it more likely that more than one dApp will try to connect to and communicate with the desktop wallet at the same time.

### TRUST-03: Ledger device owner

The ledger device owner is the human who owns and controls a Ledger device that is connected to the desktop wallet. The ledger device owner may not be the same person as the [Anonymous wallet user](#trust-00-anonymous-wallet-user) or the [Authenticated wallet user](#trust-01-authenticated-wallet-user). Additionally, it is possible for there to be multiple Ledger device owners if there are multiple Ledger devices connected to the desktop wallet.

## Entry Points

NOTE: A trust level of 1 is the lowest trust level

### ENTRY-00: Wallet GUI

This is the main method in which the user is supposed to interact with the wallet to manage their key(s). The GUI is intended to be used by people who may have very little technical knowledge about Algorand, computers (in general), and blockchain (in general). It is also intended to be used by people who have higher levels of expertise in those subjects.

**Trust Levels**:

1. Anonymous user
2. Authenticated user

### ENTRY-{{ID_NUM}}: {{Add name of entry point here}}

{{Insert description here}}

**Trust Levels**:

1. {{Lowest level of trust}}
2. {{Next lowest level of trust}}
3. {{Highest level of trust}}

## Exit Points

NOTE: A trust level of 1 is the lowest trust level

### EXIT-00: Wallet GUI

One of the main purposes of the GUI is to display information. However, it is possible it can display *too much* information in the form of error messages and account status(es).

**Trust Levels**:

1. Anonymous user
2. Authenticated user

### EXIT-{{ID_NUM}}: {{Add name of exit point here}}

{{Insert description here}}

**Trust Levels**:

1. {{Lowest level of trust}}
2. {{Next lowest level of trust}}
3. {{Highest level of trust}}

## Assets

### ASSET-00: Account private keys

An account's private key is the most valuable asset a wallet can contain. The private key is needed to transfer funds from the account and interact with the network through smart contracts (also known as "applications"). The private key is usually encoded into the form of a 25-word mnemonic to make it easier for the user to record into some non-digital medium. The purpose of wallet software is to safely store private keys in a way that makes it easier for users to use them to authenticate as their accounts to authorize interactions with the network.

**Trust Levels**:

1. Anonymous user
2. Authenticated user

### ASSET-{{ID_NUM}}: {{Add name of asset here}}

{{Insert description here}}

**Trust Levels**:

1. {{Lowest level of trust}}
2. {{Next lowest level of trust}}
3. {{Highest level of trust}}
