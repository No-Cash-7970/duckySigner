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
    - [ENTRY-01: Ledger device connection](#entry-01-ledger-device-connection)
    - [ENTRY-02: Configuration files](#entry-02-configuration-files)
    - [ENTRY-03: Database files](#entry-03-database-files)
    - [ENTRY-04: Memory](#entry-04-memory)
    - [ENTRY-05: Wallet connection server API](#entry-05-wallet-connection-server-api)
    - [ENTRY-06: Algorand node API](#entry-06-algorand-node-api)
  - [Exit Points](#exit-points)
    - [EXIT-00: Wallet GUI](#exit-00-wallet-gui)
    - [EXIT-01: Ledger device connection](#exit-01-ledger-device-connection)
    - [EXIT-02: Configuration files](#exit-02-configuration-files)
    - [EXIT-03: Database files](#exit-03-database-files)
    - [EXIT-04: Memory](#exit-04-memory)
    - [EXIT-05: Wallet connection server API](#exit-05-wallet-connection-server-api)
    - [EXIT-06: Algorand node API](#exit-06-algorand-node-api)
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

The trust levels for each entry point are listed from the highest level of trust to the lowest level of trust.

### ENTRY-00: Wallet GUI

This is the primary method in which the user is supposed to interact with the wallet to manage their keys. The GUI is intended to be used by people who may have very little technical knowledge about Algorand, computers or blockchain. It is also intended to be flexible enough to be allow for those with higher levels of expertise in those subjects to customize the desktop wallet according to their needs.

**Trust Levels**:

1. Authenticated wallet user
2. Anonymous wallet user

### ENTRY-01: Ledger device connection

A computer or mobile device can connect and communicate with Ledger device over a USB or Bluetooth connection to sign transactions. Data (e.g. signed transaction) is sent from a Ledger device to the user's computer or mobile device.

**Trust Levels**:

1. Authenticated wallet user
2. Ledger device owner
3. DApp
4. Anonymous wallet user

### ENTRY-02: Configuration files

The desktop wallet will most likely store some kind of user configuration in at least one file. In this case, the wallet will read and load the configuration data contained within this file when it initializes. Therefore, a configuration file would determine the behavior of the wallet.

**Trust Levels**:

1. Authenticated wallet user
2. DApp
3. Anonymous wallet user

### ENTRY-03: Database files

The desktop wallet may retrieve data from a file-based database, such as SQLite. The wallet will most likely obtain the user's Algorand account private keys from a database file. Also, it is possible for a [configuration file](#entry-02-configuration-files) to be a database file.

**Trust Levels**:

1. Authenticated wallet user
2. DApp
3. Anonymous wallet user

### ENTRY-04: Memory

Like all software, the desktop wallet requires reading data stored temporarily into memory. Modifications to the data in memory would change the behavior of the wallet. Paging or swapping may occur where some data in memory is temporarily stored onto the disk. This often occurs if the amount of available memory is low.

**Trust Levels**:

1. Authenticated wallet user
2. Anonymous wallet user

### ENTRY-05: Wallet connection server API

The wallet connection server API is for allowing for other software, dApps in particular, to communicate with the desktop wallet. This communication would be through HTTP. It is possible for multiple applications to attempt to communicate with the desktop wallet at the same time.

**Trust Levels**:

1. Authenticated wallet user
2. DApp
3. Anonymous wallet user

### ENTRY-06: Algorand node API

Any interaction with the Algorand blockchain must be done through an Algorand node. This interaction is typically done through an HTTP REST API. Therefore, a node will typically respond with data to a request sent to it. As is typical of a HTTP REST API server, the nature of the data in the node's response depends on the data it received in the request. If a Algorand node service is used, that service most likely imposes limits on the number and size of requests.

**Trust Levels**:

1. Authenticated wallet user
2. Anonymous wallet user

## Exit Points

The trust levels for each exit point are listed from the highest level of trust to the lowest level of trust.

### EXIT-00: Wallet GUI

One of the main purposes of the GUI is to display information. However, it is possible it can display *too much* information in the form of error messages and Algorand account statuses.

**Trust Levels**:

1. Authenticated wallet user
2. Anonymous wallet user

### EXIT-01: Ledger device connection

In order sign a transaction using a key stored on a Ledger device, the unsigned transaction data must be sent to the Ledger device.

**Trust Levels**:

1. Authenticated wallet user
2. Ledger device owner
3. DApp
4. Anonymous wallet user

### EXIT-02: Configuration files

The desktop wallet may write or edit configuration files that store the desktop wallet's configuration. The configuration stored in the file may give too much information about how the user is using the wallet.

**Trust Levels**:

1. Authenticated wallet user
2. Anonymous wallet user

### EXIT-03: Database files

The desktop wallet may write to and update a file-based database, such as SQLite. The wallet will most likely add the user's wallet account private keys to a database file.

**Trust Levels**:

1. Authenticated wallet user
2. DApp
3. Anonymous wallet user

### EXIT-04: Memory

Like all software, the desktop wallet requires writing data into memory. However, some data in memory may be temporarily stored onto the disk if the amount of memory is low.

**Trust Levels**:

1. Authenticated wallet user
2. DApp
3. Anonymous wallet user

### EXIT-05: Wallet connection server API

The wallet connection server API responds with data when requested by some entity. Typically, this entity should be a dApp approved by the user.

**Trust Levels**:

1. Authenticated wallet user
2. DApp
3. Anonymous wallet user

### EXIT-06: Algorand node API

Changing the state of something (e.g. account, smart contract) on the Algorand blockchain requires submitting a transaction. Submitting a transaction requires sending the signed transaction data to Algorand node through its HTTP REST API.

**Trust Levels**:

1. Anonymous wallet user

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
