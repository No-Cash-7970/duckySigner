<!-- omit in toc -->
# Threat Model Information

**Application Version:** N/A

**Description:** Ducky Signer is a prototype desktop wallet for the Algorand public blockchain. Being that it is a prototype, Ducky Signer should be simple with limited functionality. The goal of the prototype is to explore the technical feasibility of a secure and user-friendly desktop wallet for Algorand. The primary purpose of a wallet is to make it easier for the use to safely store the private key(s) to their Algorand accounts while enabling those users to use the private key(s) to cryptographically sign various things using their desktop computer.

**Document Owner**: No-Cash-7970

**Participants**: No-Cash-7970

<!-- omit in toc -->
## Table of Contents

- [External Dependencies](#external-dependencies)
  - [EXTERN-01: Algorand network](#extern-01-algorand-network)
  - [EXTERN-02: User's machine](#extern-02-users-machine)
  - [EXTERN-03: Algorand node](#extern-03-algorand-node)
  - [EXTERN-04: Hardware wallet device](#extern-04-hardware-wallet-device)
- [Trust Levels](#trust-levels)
  - [TRUST-01: Anonymous wallet user](#trust-01-anonymous-wallet-user)
  - [TRUST-02: Authenticated wallet user](#trust-02-authenticated-wallet-user)
  - [TRUST-03: DApp](#trust-03-dapp)
  - [TRUST-04: Ledger device owner](#trust-04-ledger-device-owner)
- [Entry Points](#entry-points)
  - [ENTRY-01: Wallet GUI](#entry-01-wallet-gui)
  - [ENTRY-02: Ledger device connection](#entry-02-ledger-device-connection)
  - [ENTRY-03: Configuration files](#entry-03-configuration-files)
  - [ENTRY-04: Database files](#entry-04-database-files)
  - [ENTRY-05: Memory](#entry-05-memory)
  - [ENTRY-06: Wallet connection server API](#entry-06-wallet-connection-server-api)
  - [ENTRY-07: Algorand node API](#entry-07-algorand-node-api)
- [Exit Points](#exit-points)
  - [EXIT-01: Wallet GUI](#exit-01-wallet-gui)
  - [EXIT-02: Ledger device connection](#exit-02-ledger-device-connection)
  - [EXIT-03: Configuration files](#exit-03-configuration-files)
  - [EXIT-04: Database files](#exit-04-database-files)
  - [EXIT-05: Memory](#exit-05-memory)
  - [EXIT-06: Wallet connection server API](#exit-06-wallet-connection-server-api)
  - [EXIT-07: Algorand node API](#exit-07-algorand-node-api)
- [Assets](#assets)
  - [ASSET-01: Account private keys](#asset-01-account-private-keys)
  - [ASSET-02: User preferences](#asset-02-user-preferences)
  - [ASSET-03: Algorand account information](#asset-03-algorand-account-information)

## External Dependencies

> External dependencies are items external to the code of the application that may pose a threat to the application. These items are typically still within the control of the organization, but possibly not within the control of the development team.
>
> — [Threat Modeling Process - OWASP](https://owasp.org/www-community/Threat_Modeling_Process#external-dependencies)

### EXTERN-01: Algorand network

All data regarding an Algorand wallet account (except for secrets like the private key) is stored on a public blockchain network. The most notable datum is the amount of funds within an account in the form of Algos or some Algorand Standard Asset (ASA). In addition to the account data, the network stores every successful Algorand transaction that has occurred on that network's chain. There are three main networks for Algorand: MainNet, TestNet and BetaNet. MainNet is the definitive network for assets linked to valuable resources in the real world (like money). TestNet and BetaNet are for testing.

### EXTERN-02: User's machine

The user's machine is most likely to be a laptop or desktop computer, but it could also be a type of tablet computer. The "user's machine" not only refers to the physical machine owned by the user, but it also refers to the machine's "soft" components such as memory, operating system and file system.

### EXTERN-03: Algorand node

An Algorand node is a special type of server that is required to connect to and communicate with the [Algorand Network](#extern-01-algorand-network). Oftentimes, a node runs 24/7 like many servers on the internet. However, a node can be a small machine, such as a Raspberry Pi, in a local network. Alternatively, a cloud server can be used to create and host a node remotely. Node services, such as [Nodely (AlgoNode)](https://nodely.io/), are commonly used in many projects in the Algorand ecosystem. It is also possible for a user to setup a node on a machine not dedicated to running a node, such as a home or work computer. The node in this instance would not be running 24/7 and would have stale data after being turned off.

### EXTERN-04: Hardware wallet device

A hardware wallet device allows a user to use their wallet account keys to sign transactions without exposing the keys to any system outside the device. Currently, the only hardware wallet that Algorand supports is Ledger. All Ledger devices can connect to a computer or mobile device using USB, but some models can connect using Bluetooth. It is possible for multiple hardware wallets to be connected to the desktop wallet at the same time.

## Trust Levels

> Trust levels represent the access rights that the application will grant to external entities. The trust levels are cross-referenced with the entry points and assets. This allows us to define the access rights or privileges required at each entry point, and those required to interact with each asset.
>
> — [Threat Modeling Process - OWASP](https://owasp.org/www-community/Threat_Modeling_Process#trust-levels)

The trust levels listed in this section are not in any particular order.

### TRUST-01: Anonymous wallet user

An anonymous wallet user is a user, typically a human, who **has not** authenticated themselves by provided a set of credentials (e.g. username and password) to access the protected parts of the desktop wallet that provide sensitive Algorand wallet account information and allow for the user to use account keys. It is assumed that an anonymous user has access to the machine and the desktop wallet installed on it. It is possible for this type of user to not be the owner of the private keys stored within the desktop wallet or be the owner of the machine itself. It may be possible for there to be more than one anonymous user at the same time.

### TRUST-02: Authenticated wallet user

An authenticated wallet user is a user, typically a human, who **has** somehow authenticated themselves by provided a set of credentials (e.g. username and password) to access the protected parts of the desktop wallet that provide sensitive Algorand wallet account information. Typically, there should be at most one authenticated wallet user accessing the protected parts of the desktop wallet at any given time.

### TRUST-03: DApp

For this document, a dApp ("decentralized" application) is simply software that uses the Algorand blockchain in some manner. The degree of centralization of the dApp software's structure and its functionality is irrelevant. Consequently, a "dApp" in the context of this document does not need to be decentralized, contrary to other commonly accepted definitions of "dApp." DApps are typically web applications run by a web browser. However, with a desktop wallet that does not depend on a web browser, it is possible for a dApp to be software that is installed on the user's machine that also does not depend on a web browser. This makes it more likely that more than one dApp will try to connect to and communicate with the desktop wallet at the same time.

### TRUST-04: Ledger device owner

The ledger device owner is the human who owns and controls a Ledger device that is connected to the desktop wallet. The ledger device owner may not be the same person as the [Anonymous wallet user](#trust-01-anonymous-wallet-user) or the [Authenticated wallet user](#trust-02-authenticated-wallet-user). Additionally, it is possible for there to be multiple Ledger device owners if there are multiple Ledger devices connected to the desktop wallet.

## Entry Points

> Entry points define the interfaces through which potential attackers can interact with the application or supply it with data. In order for a potential attacker to attack an application, entry points must exist.
>
> ...
>
> \[They\] show where data enters the system (i.e. input fields, methods) and exit points are where it leaves the system (i.e. dynamic output, methods), respectively.
>
> — [Threat Modeling Process - OWASP](https://owasp.org/www-community/Threat_Modeling_Process#entry-points)

The trust levels for each entry point are listed from the highest level of trust to the lowest level of trust.

### ENTRY-01: Wallet GUI

This is the primary method in which the user is supposed to interact with the wallet to manage their keys. The GUI is intended to be used by people who may have very little technical knowledge about Algorand, computers or blockchain. It is also intended to be flexible enough to be allow for those with higher levels of expertise in those subjects to customize the desktop wallet according to their needs.

**Trust Levels**:

1. Authenticated wallet user
2. Anonymous wallet user

### ENTRY-02: Ledger device connection

A computer or mobile device can connect and communicate with a Ledger device over a USB or Bluetooth connection to sign transactions. Data (e.g. a signed transaction) is sent from a Ledger device to the user's computer or mobile device.

**Trust Levels**:

1. Authenticated wallet user
2. Ledger device owner
3. DApp
4. Anonymous wallet user

### ENTRY-03: Configuration files

The desktop wallet will most likely store some kind of user configuration in at least one file. In this case, the wallet will read and load the configuration data contained within this file when it initializes. Therefore, a configuration file would determine the behavior of the wallet.

**Trust Levels**:

1. Authenticated wallet user
2. DApp
3. Anonymous wallet user

### ENTRY-04: Database files

The desktop wallet may retrieve data from a file-based database, such as SQLite. The wallet will most likely obtain the user's Algorand account private keys from a database file. Also, it is possible for a [configuration file](#entry-03-configuration-files) to be a database file.

**Trust Levels**:

1. Authenticated wallet user
2. DApp
3. Anonymous wallet user

### ENTRY-05: Memory

Like all software, the desktop wallet requires reading data stored temporarily into memory. Modifications to the data in memory would change the behavior of the wallet. Paging or swapping may occur where some data in memory is temporarily stored onto the disk. This often occurs if the amount of available memory is low.

**Trust Levels**:

1. Authenticated wallet user
2. Anonymous wallet user

### ENTRY-06: Wallet connection server API

The wallet connection server API is for allowing for other software, dApps in particular, to communicate with the desktop wallet. This communication would be through HTTP. It is possible for multiple applications to attempt to communicate with the desktop wallet at the same time.

**Trust Levels**:

1. Authenticated wallet user
2. DApp
3. Anonymous wallet user

### ENTRY-07: Algorand node API

Any interaction with the Algorand blockchain must be done through an Algorand node. This interaction is typically done through an HTTP REST API. Therefore, a node will typically respond with data to a request sent to it. As is typical of a HTTP REST API server, the nature of the data in the node's response depends on the data it received in the request. If a Algorand node service is used, that service most likely imposes limits on the number and size of requests.

**Trust Levels**:

1. Authenticated wallet user
2. Anonymous wallet user

## Exit Points

> Exit points might prove useful when attacking the client: for example, cross-site-scripting vulnerabilities and information disclosure vulnerabilities both require an exit point for the attack to complete.
>
> In the case of exit points from components handling confidential data (e.g. data access components), exit points lacking security controls to protect confidentiality and integrity can lead to disclosure of such confidential information to an unauthorized user.
>
> In many cases threats enabled by exit points are related to the threats of the corresponding entry point.
>
> — [Threat Modeling Process - OWASP](https://owasp.org/www-community/Threat_Modeling_Process#exit-points)

The trust levels for each exit point are listed from the highest level of trust to the lowest level of trust.

### EXIT-01: Wallet GUI

One of the main purposes of the GUI is to display information. However, it is possible it can display *too much* information in the form of error messages and Algorand account statuses.

**Trust Levels**:

1. Authenticated wallet user
2. Anonymous wallet user

### EXIT-02: Ledger device connection

In order sign a transaction using a key stored on a Ledger device, the unsigned transaction data must be sent to the Ledger device.

**Trust Levels**:

1. Authenticated wallet user
2. Ledger device owner
3. DApp
4. Anonymous wallet user

### EXIT-03: Configuration files

The desktop wallet may write or edit configuration files that store the desktop wallet's configuration. The configuration stored in the file may give too much information about how the user is using the wallet.

**Trust Levels**:

1. Authenticated wallet user
2. Anonymous wallet user

### EXIT-04: Database files

The desktop wallet may write to and update a file-based database, such as SQLite. The wallet will most likely add the user's wallet account private keys to a database file.

**Trust Levels**:

1. Authenticated wallet user
2. DApp
3. Anonymous wallet user

### EXIT-05: Memory

Like all software, the desktop wallet requires writing data into memory. However, some data in memory may be temporarily stored onto the disk if the amount of memory is low.

**Trust Levels**:

1. Authenticated wallet user
2. DApp
3. Anonymous wallet user

### EXIT-06: Wallet connection server API

The wallet connection server API responds with data when requested by some entity. Typically, this entity should be a dApp approved by the user.

**Trust Levels**:

1. Authenticated wallet user
2. DApp
3. Anonymous wallet user

### EXIT-07: Algorand node API

Changing the state of something (e.g. account, smart contract) on the Algorand blockchain requires submitting a transaction. Submitting a transaction requires sending the signed transaction data to Algorand node through its HTTP REST API.

**Trust Levels**:

1. Anonymous wallet user

## Assets

> Assets are essentially targets for attackers, i.e. they are the reason threats will exist. Assets can be both physical assets and abstract assets. For example, an asset of an application might be a list of clients and their personal information; this is a physical asset. An abstract asset might be the reputation of an organization.
>
> — [Threat Modeling Process - OWASP](https://owasp.org/www-community/Threat_Modeling_Process#assets)

The trust levels for each asset are listed from the highest level of trust to the lowest level of trust.

### ASSET-01: Account private keys

An account's private key is the most valuable asset any wallet software can contain. The private key is needed to transfer funds from the account and interact with the network through smart contracts (also known as "applications"). The private key is usually encoded into the form of a 25-word mnemonic to make it easier for the user to record it into some secure, typically non-digital, medium. The purpose of wallet software is to safely store private keys in a way that makes it easier or possible for users to use them to authorize their accounts' interactions with the network.

**Trust Levels**:

1. Authenticated wallet user
2. DApp
3. Anonymous wallet user

### ASSET-02: User preferences

The user's preferences for the desktop wallet may prove useful for an attacker, especially for phishing.

**Trust Levels**:

1. Authenticated wallet user
2. DApp
3. Anonymous wallet user

### ASSET-03: Algorand account information

Although the information about an Algorand account is publicly available for free for anyone willing to look for it, the user may not want to reveal details about their account by having it displayed on a screen that anyone can read from a distance. This is because revealing such information could lead to harm to the user in the real world, such as a ["$5 wrench attack"](https://xkcd.com/538/).

**Trust Levels**:

1. Authenticated wallet user
2. DApp
3. Anonymous wallet user
