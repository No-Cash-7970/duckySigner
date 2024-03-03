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
  - [Entry Points](#entry-points)
    - [ENTRY-00: Wallet GUI](#entry-00-wallet-gui)
    - [ENTRY-{{ID\_NUM}}: {{Add name of entry point here}}](#entry-id_num-add-name-of-entry-point-here)
  - [Exit Points](#exit-points)
    - [EXIT-00: Wallet GUI](#exit-00-wallet-gui)
    - [EXIT-{{ID\_NUM}}: {{Add name of exit point here}}](#exit-id_num-add-name-of-exit-point-here)
  - [Assets](#assets)
    - [ASSET-00: Account private keys](#asset-00-account-private-keys)
    - [ASSET-{{ID\_NUM}}: {{Add name of asset here}}](#asset-id_num-add-name-of-asset-here)
  - [Trust Levels](#trust-levels)
    - [TRUST-00: Anonymous User](#trust-00-anonymous-user)
    - [TRUST-01: Authenticated user](#trust-01-authenticated-user)

## External Dependencies

### EXTERN-00: Algorand network

All data regarding an Algorand wallet account (except for secrets like the private key) is stored on a public blockchain network. The most notable datum is the amount of funds within an account in the form of Algos or some Algorand Standard Asset (ASA). In addition to the account data, the network stores every successful Algorand transaction that has occurred on that network's chain. There are three main networks for Algorand: MainNet, TestNet and BetaNet. MainNet is the definitive network for assets linked to valuable resources in the real world (like money). TestNet and BetaNet are for testing.

### EXTERN-01: User's machine

The user's machine is most likely to be a laptop or desktop computer, but it could also be a type of tablet computer. The "user's machine" not only refers to the physical machine owned by the user, but it also refers to the machine's "soft" components such as memory, operating system and file system.

### EXTERN-02: Algorand node

An Algorand node is a special type of server that is required to connect to and communicate with the [Algorand Network](#extern-00-algorand-network). Oftentimes, a node runs 24/7 like many servers on the internet. However, a node can be a small machine, such as a Raspberry Pi, in a local network. Alternatively, a cloud server can be used to create and host a node remotely. Node services, such as [Nodely (AlgoNode)](https://nodely.io/), are commonly used in many projects in the Algorand ecosystem. It is also possible for a user to setup a node on a machine not dedicated for running a node, such as a home or work computer. The node in this instance would not be running 24/7 and would have stale data after being turned off.

### EXTERN-03: Hardware wallet device

A hardware wallet device allows a user to use their wallet account keys to sign transactions without exposing the keys to any system outside the device. Currently, the only hardware wallet that Algorand supports is Ledger. All Ledger devices can connect to a computer or mobile device using USB, but some models can connect using Bluetooth. It is possible for multiple hardware wallets to be connected to the desktop wallet at the same time.

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

## Trust Levels

### TRUST-00: Anonymous User

A user who **has not** provided a set of credentials (e.g. username and password) to access the protected parts of the wallet software that provide sensitive Algorand wallet account information and allow for the user to use account keys.

### TRUST-01: Authenticated user

A user who **has** provided a set of credentials (e.g. username and password) to access the protected parts of the wallet software that provide sensitive Algorand wallet account information.
