<!-- omit in toc -->
# List of Threats

Not an exhaustive list of threats.

<!-- omit in toc -->
## Table of contents

- [THREAT-001: Impersonation of a trustworthy dApp or platform](#threat-001-impersonation-of-a-trustworthy-dapp-or-platform)
- [THREAT-002 Impersonation of this desktop wallet project](#threat-002-impersonation-of-this-desktop-wallet-project)
- [THREAT-003: Compromised GitHub repository](#threat-003-compromised-github-repository)
- [THREAT-004: Exposure of account private key mnemonic displayed on screen to third-party](#threat-004-exposure-of-account-private-key-mnemonic-displayed-on-screen-to-third-party)
- [THREAT-005: Exploitation of vulnerabilities exposed in public documentation](#threat-005-exploitation-of-vulnerabilities-exposed-in-public-documentation)
- [THREAT-006: Exploitation of vulnerabilities exposed in publicly published source code](#threat-006-exploitation-of-vulnerabilities-exposed-in-publicly-published-source-code)
- [THREAT-007: Guessing the wallet password](#threat-007-guessing-the-wallet-password)
- [THREAT-008: Modification of data sent to hardware wallet](#threat-008-modification-of-data-sent-to-hardware-wallet)
- [THREAT-009: Interception of HTTP communication between dApp and wallet connection server](#threat-009-interception-of-http-communication-between-dapp-and-wallet-connection-server)
- [THREAT-010: Wallet connection server overwhelmed by too many requests or requests that are too large](#threat-010-wallet-connection-server-overwhelmed-by-too-many-requests-or-requests-that-are-too-large)
- [THREAT-011: Cracking encryption on wallet files offline](#threat-011-cracking-encryption-on-wallet-files-offline)
- [THREAT-012: Modifying security settings in configuration files](#threat-012-modifying-security-settings-in-configuration-files)
- [THREAT-013: Wallet password is written down and stored in insecure location](#threat-013-wallet-password-is-written-down-and-stored-in-insecure-location)
- [THREAT-014: Spam transaction with scam link in note field](#threat-014-spam-transaction-with-scam-link-in-note-field)
- [THREAT-015: Algorand node malfunctioning](#threat-015-algorand-node-malfunctioning)
- [THREAT-016: Algorand chain halts](#threat-016-algorand-chain-halts)
- [THREAT-017: Transaction spam harassment](#threat-017-transaction-spam-harassment)
- [THREAT-018: Compromised software dependency](#threat-018-compromised-software-dependency)
- [THREAT-019: Impersonation of software dependency](#threat-019-impersonation-of-software-dependency)
- [THREAT-020: User overwhelmed by too many request prompts](#threat-020-user-overwhelmed-by-too-many-request-prompts)
- [THREAT-021: Change to less secure settings](#threat-021-change-to-less-secure-settings)
- [THREAT-022: Multiple dApps establishing connect sessions with the same dApp ID](#threat-022-multiple-dapps-establishing-connect-sessions-with-the-same-dapp-id)

## THREAT-001: Impersonation of a trustworthy dApp or platform

- **Actor:** Scammer/Spammer, Cybercriminal
- **Purpose:** To manipulate the user into handing over their private keys, signing a malicious transaction, or signing a malicious program (logic signature) to steal funds and assets from their account
- **Target:** Funds and assets in the user's account that can be sold for fiat money (e.g. USD, Euro)
- **Action:** Actor creates a fake website that looks like the website of the trustworthy dApp, perhaps one the user has used before, with a URL that looks official. Then they post about the website on social media, possibly with a compromised account of a typically trustworthy person or entity, to promote this fake website in a way that is enticing to users (e.g. "sign up for airdrop").
- **Result of the action:** The private key of the account the user used to interact with website is compromised, and then funds and assets within account are stolen
- **Occurrence likelihood**: High
- **Impact:** Low to high, depending on what funds and assets are in the user's account, for what purposes the user uses their account, how quickly users are warned, and how quickly the fake website is taken down
- **Threat type:** Spoofing
- **Potential mitigations:**
  1. Only allow certain "trusted" dApps to connect to desktop wallet (i.e. whitelist dApps)
  2. Allow user to have a list of "favorites" with links to web dApps
  3. Do not have the feature of dApps being able to connect to wallet
  4. Disallow signing of certain dangerous and advanced transactions (e.g. transaction with rekey) by default
  5. Disallow signing of logic signatures by default

[Back to top ↑](#table-of-contents)

## THREAT-002 Impersonation of this desktop wallet project

- **Actor:** Scammer/Spammer, Cybercriminal
- **Purpose:** To manipulate the user into handing over their private keys, signing a malicious transaction, or signing a malicious program (logic signature) to steal funds and assets from their account
- **Target:** Funds and assets in the user's account that can be sold for fiat money (e.g. USD, Euro)
- **Action:** Actor creates a fake website that looks like the website of the desktop app with a URL that looks official. Then they post about the website on social media, possibly with a compromised account of a typically trustworthy person (e.g. a maintainer of the desktop wallet project), to promote this fake website in a way that is enticing to users (e.g. "sign up for airdrop").
- **Result of the action:** The private keys put into the fake desktop wallet are compromised, and then funds and assets within accounts are stolen
- **Occurrence likelihood**: High
- **Impact:** Low to high, depending on what funds and assets were in the accounts, for what purposes the user uses their accounts, how quickly users are warned, and how quickly the fake website is taken down
- **Threat type:** Spoofing
- **Potential mitigations:**
  1. Explicitly state on official websites and social media accounts that no one associated with project will ask for private key, send an unsolicited DM, create an airdrop, etc.
  2. Sign each released binary, similar to how [go-algorand releases](https://github.com/algorand/go-algorand/releases) are done. Also encourage the user to verify the signature of a downloaded binary by providing directions for how user can verify the signature.

[Back to top ↑](#table-of-contents)

## THREAT-003: Compromised GitHub repository

- **Actor:** Cybercriminal, rogue maintainer
- **Purpose:** To modify the code so that user private keys put into wallet are compromised
- **Target:** Typically the parts of the code that manage the private keys
- **Action:** Actor takes advantage of weak security of a maintainer's GitHub account and gains access to that account, or creates a pull request with malicious code and hopes maintainer(s) merge changes without noticing the malicious code
- **Result of the action:** The actor can use compromised private keys to drain the accounts of users
- **Occurrence likelihood**: Medium
- **Impact:** Very high
- **Threat type:** Tampering
- **Potential mitigations:**
  1. Require all maintainers protect their accounts with 2-factor authentication
  2. Establish a strict policy of reviewing every pull requests before merging changes and rejecting any change that cannot be understood or justified

[Back to top ↑](#table-of-contents)

## THREAT-004: Exposure of account private key mnemonic displayed on screen to third-party

- **Actor:** Anyone who can view user's screen
- **Purpose:** To obtain user's private key
- **Target:** User's screen, which is displaying a private key mnemonic (also known as a "seed phrase")
- **Action:** Actor records the mnemonic by taking a photo of the screen or writing down the mnemonic. This can occur when the user is entering the mnemonic into the desktop wallet or viewing the mnemonic for a private key stored in the wallet.
- **Result of the action:** The actor obtains the private key of the account that can then be used to drain the account immediately or at a later time (months or years since recording the mnemonic)
- **Occurrence likelihood**: Low to high, depending on user's environment
- **Impact:** High
- **Threat type:** Information Disclosure
- **Potential mitigations:**
  1. Mask the input of each word in the mnemonic when the user is entering the mnemonic, similar to how user entry of passwords are typically handled in most software
  2. Whenever the user wants to view the mnemonic, display a dialog box that asks the user if they are in an environment safe from prying eyes

[Back to top ↑](#table-of-contents)

## THREAT-005: Exploitation of vulnerabilities exposed in public documentation

- **Actor:** Cybercriminal
- **Purpose:** To find ways to attack and compromise the desktop wallet
- **Target:** Codebase
- **Action:** Actor examines the documentation, such as the architecture diagrams and the threat model documentation, to look for design flaws in the desktop wallet that can be used to create exploits
- **Result of the action:** The desktop wallet is compromised and users' private keys are potentially exposed, which can lead to users' accounts getting drained of all assets and funds
- **Occurrence likelihood**: Medium
- **Impact:** Medium
- **Threat type:** Information Disclosure
- **Potential mitigations:**
  1. Do not publish all documentation to the public. For example, do not publish documentation of an unaddressed dangerous threat in this *List of Threats*.

[Back to top ↑](#table-of-contents)

## THREAT-006: Exploitation of vulnerabilities exposed in publicly published source code

- **Actor:** Cybercriminal
- **Purpose:** To find ways to attack and compromise the desktop wallet
- **Target:** Codebase
- **Action:** Actor examines the codebase to look for bugs and design flaws in the desktop wallet that can be used to create exploits
- **Result of the action:** The desktop wallet is compromised and users' private keys are potentially exposed, which can lead to users' accounts getting drained of all assets and funds
- **Occurrence likelihood**: Medium
- **Impact:** Medium
- **Threat type:** Information Disclosure
- **Potential mitigations:**
  1. Do not publish source code publicly (closed source)
  2. Do not publish the most recent version of the code
  3. Periodically invite others (white hats) to review code for potential vulnerabilities

[Back to top ↑](#table-of-contents)

## THREAT-007: Guessing the wallet password

- **Actor:** Anyone with access to wallet on user's machine
- **Purpose:** To access the private keys, to view information about user's accounts
- **Target:** Private keys, account information (e.g. account address, account balance, transaction history)
- **Action:** Actor uses a rainbow table or information known about user to guess the wallet password
- **Result of the action:** The actor can drain user's accounts, use the user's account to attack some other account or application (smart contract)
- **Occurrence likelihood**: Low to high, depending on user's environment
- **Impact:** High
- **Threat type:** Elevation of privilege
- **Potential mitigations:**
  1. Require user to use a strong password (e.g. at least one capital letter, at least one symbol, etc.)
  2. Allow user to somehow use 2-factor authentication (e.g. Time-based One-Time Password (TOTP), FIDO U2F device)

[Back to top ↑](#table-of-contents)

## THREAT-008: Modification of data sent to hardware wallet

- **Actor:** Cybercriminal, malware, a bug
- **Purpose:** To make user sign transactions or bytes they did not intend to sign. There may be no purpose when the modification of the transaction data is the result of a bug
- **Target:** Hardware wallet
- **Action:** Actor somehow modifies the transaction or byte data sent from the desktop wallet to the hardware wallet for user to approve the signing
- **Result of the action:** The user uses their hardware wallet to sign a transaction they never intended to sign, such as a transaction that drains their account or approving Algorand application (smart contract) that hands over control of the account to the actor (e.g. rekeying). In the case of a dApp or system that uses signing bytes as part of an authentication scheme (i.e. logging in with a wallet account), the user unintentionally signs bytes that gives the actor restricted access.
- **Occurrence likelihood**: Very low
- **Impact:** High
- **Threat type:** Tampering
- **Potential mitigations:**
  1. Display notice to user telling them to check transaction data shown on desktop wallet is same as data shown on hardware wallet
  2. Display transaction data in a way that is similar to hardware wallet so it is easier for the user to check if the data matches

[Back to top ↑](#table-of-contents)

## THREAT-009: Interception of HTTP communication between dApp and wallet connection server

- **Actor:** Malware
- **Purpose:** To disable communication between dApp and wallet connection server, to extract sensitive or secret data (e.g. authentication keys), to modify the communication to change behavior of dApp or desktop wallet
- **Target:** Data within the HTTP communications
- **Action:** Actor listens to HTTP communications and either (1) intercepts and halts these communications or (2) extracts useful data from the communications
- **Result of the action:** Actor can then use the extracted data (e.g. authentication key) to impersonate trusted dApp and get user to sign a dangerous transaction
- **Occurrence likelihood**: Medium
- **Impact:** High
- **Threat type:** Information disclosure, tampering, denial of service
- **Potential mitigations:**
  1. Use Transport Layer Security (TLS), if reasonably possible
  2. Do not send secret or sensitive data over HTTP or WebSockets

[Back to top ↑](#table-of-contents)

## THREAT-010: Wallet connection server overwhelmed by too many requests or requests that are too large

- **Actor:** Malware, malicious or malfunctioning dApp
- **Purpose:** To overwhelm the wallet connection server and cause it to go into an invalid state or disable it, to cause the desktop wallet to consume too much memory that paging/swapping is needed which can cause unencrypted secret data (e.g. private keys, authentication keys) to be written onto the disk, malfunction (no purpose)
- **Target:** Functionality of the desktop wallet, secret data (e.g. private keys, authentication keys)
- **Action:** The actor creates and sends a large number of HTTP requests or a few large request (e.g. 1 GB of data) to the wallet connection server
- **Result of the action:** The desktop wallet is unable to communicate with legitimate dApps, and exposed secret data may be used to impersonate a trusted dApp or drain the user's accounts
- **Occurrence likelihood**: Medium
- **Impact:** High
- **Threat type:** Information disclosure, denial of service
- **Potential mitigations:**
  1. Throttle the number of requests
  2. Set and enforce a maximum size for request header and request body
  3. Allow wallet connection server to be disabled or switched off
  4. Protect secret data temporarily stored in memory with a software enclave by using something the like [MemGuard](https://pkg.go.dev/github.com/awnumar/memguard) in case a memory paging/swapping occurs
  5. Allow the user to switch the wallet connection server off, even when the server is receiving requests
  6. Allow the user to switch the wallet connection server on for a limited period of time. This would limit the window of opportunity the actor has in doing this type of attack.

[Back to top ↑](#table-of-contents)

## THREAT-011: Cracking encryption on wallet files offline

- **Actor:** Cybercriminal, malware
- **Purpose:** To access the private keys
- **Target:** Encrypted files where private keys are stored
- **Action:** The actor copies the wallet files and then brute force guesses the password(s) used to encrypt those files in some other location, possibly with a large amount of computing power
- **Result of the action:** The actor gains access to the private keys stored in files
- **Occurrence likelihood**: Medium
- **Impact:** High
- **Threat type:** Elevation of privilege
- **Potential mitigations:**
  1. Use a strong cryptographic scheme to encrypt files

[Back to top ↑](#table-of-contents)

## THREAT-012: Modifying security settings in configuration files

- **Actor:** Cybercriminal, malware
- **Purpose:** To decrease the security protections of the desktop wallet
- **Target:** Security settings within configuration files
- **Action:** The actor modifies the security settings in the configuration files in a way that reduces the security of the desktop wallet
- **Result of the action:** The desktop wallet with reduced security is more vulnerable to some other attack
- **Occurrence likelihood**: Medium
- **Impact:** High
- **Threat type:** Tampering
- **Potential mitigations:**
  1. Encrypt entire configuration file
  2. Put security settings into a separate encrypted configuration file, while other settings are placed into an unencrypted file
  3. Put security settings into same encrypted file as the private keys

[Back to top ↑](#table-of-contents)

## THREAT-013: Wallet password is written down and stored in insecure location

- **Actor:** User
- **Purpose:** Convenience, as protective measure against forgetting wallet password
- **Target:** Wallet password
- **Action:** The user writes down password and stores in an insecure location (e.g. underneath keyboard, in a drawer nearby)
- **Result of the action:** Access to the wallet does not solely rely on the user remembering the password and it is easily accessible to the user...and anyone else curious enough to look around
- **Occurrence likelihood**: Medium, however, likelihood of password being found depends on the user's environment
- **Impact:** High
- **Threat type:** Information disclosure
- **Potential mitigations:**
  1. Support a secure mechanism for backing up keys and recovering them so user does not need to worry about losing access to keys due to losing the password.
  2. Suggest user to use a password manager
  3. Warn user about storing password in an insecure location

[Back to top ↑](#table-of-contents)

## THREAT-014: Spam transaction with scam link in note field

- **Actor:** Scammer/Spammer, Cybercriminal
- **Purpose:** To deceive user into signing a malicious transaction or handing over private keys, to install malware onto the user's computer
- **Target:** Whatever funds and assets are in the user's account that can be sold for fiat money (e.g. USD, Euro)
- **Action:** Send a transaction to the user with a scam link in the note field accompanied with a message that is intended to entice the user into clicking the link and interacting with the scam website
- **Result of the action:** If user ignores the transaction, then nothing happens. If the user clicks on link and interacts with the scam website, the user potentially loses funds or assets.
- **Occurrence likelihood**: High
- **Impact:** Low to High, depends on how much the user interacts with the scam
- **Threat type:** Spoofing, denial of service
- **Potential mitigations:**
  1. Do not allow links in transaction notes to be clicked on
  2. Upon detecting if there is a link in the note, warn the user about not going to website(s) in transaction notes because it could be a scam
  3. Allow user to block being shown or notified of transactions from certain addresses (blacklist)
  4. Allow user to whitelist which addresses can trigger a notification
  5. Allow user to choose the minimum amount of Algos or a certain Algorand Standard Asset (ASA) sent in a transaction that triggers a notification or gets the transaction shown to the user
  6. Disallow signing of certain dangerous and advanced transactions (e.g. transaction with rekey) by default
  7. Disallow signing of logic signatures by default

[Back to top ↑](#table-of-contents)

## THREAT-015: Algorand node malfunctioning

- **Actor:** Unknown
- **Purpose:** None
- **Target:** Algorand node
- **Action:** The node malfunctions and is not able to retrieve the current blockchain state. This may be caused by a number of things, such as an internet service outage, a shortage of computing resources, or hardware failure.
- **Result of the action:** The data from the node is outdated and the node cannot submit transactions to the network.
- **Occurrence likelihood**: High
- **Impact:** Low
- **Threat type:** Denial of service
- **Potential mitigations:**
  1. Design components to be able to handle situations where the node is not available or has outdated data
  2. Display notice to user, in a way that does not cause panic, whenever node is malfunctioning. Mention that blockchain is highly likely to be functioning just fine. Recommend changing node settings (within desktop wallet) or fixing issues with node, if it is their own node.

[Back to top ↑](#table-of-contents)

## THREAT-016: Algorand chain halts

- **Actor:** Algorand network
- **Purpose:** None, most likely due to malfunction in the Algorand network
- **Target:** Algorand blockchain
- **Action:** For some reason, the network fails to add blocks to the blockchain and process transactions
- **Result of the action:** The Algorand blockchain is in a state where it does not function
- **Occurrence likelihood**: Low
- **Impact:** Medium
- **Threat type:** Denial of service
- **Potential mitigations:**
  1. Design components to be able to handle situations where the chain halts
  2. Display notice to user, in a way that does not cause panic, whenever chain is malfunctioning. Assure the funds and assets are likely to be fine.

[Back to top ↑](#table-of-contents)

## THREAT-017: Transaction spam harassment

- **Actor:** Anyone
- **Purpose:** To harass or annoy the user by having the wallet constantly notify the user of a transaction
- **Target:** User
- **Action:** The actor spams the user's account with a transaction often enough to be an annoyance or inconvenience to the user
- **Result of the action:** User's emotional or psychological well-being deceases, user cannot use the desktop wallet as they normally would
- **Occurrence likelihood**: Medium
- **Impact:** Medium
- **Threat type:** Denial of service
- **Potential mitigations:**
  1. Allow user to block being shown or notified of transactions from certain addresses (blacklist)
  2. Allow user to whitelist which addresses can trigger a notification
  3. Allow user to choose the minimum amount of Algos or certain Algorand Standard Asset (ASA) sent in a transaction that triggers a notification or gets the transaction shown to the user

[Back to top ↑](#table-of-contents)

## THREAT-018: Compromised software dependency

- **Actor:** Compromised software dependency developer/maintainer, malicious software dependency contributor, cybercriminal (usually remote)
- **Purpose:** To compromise software that use the dependency
- **Target:** Software that use the dependency, user's sensitive or secret data managed by software that use the dependency
- **Action:** The actor compromises the software dependency by changing its code and then releases an update with the malicious code. Other codebases using the dependency install the update. This causes the malicious code to be included in whatever software is released to users. The malicious code sends sensitive or secret data to the actor.
- **Result of the action:** Modification of the behavior of software that use the dependency, user's sensitive or secret data is leaked to the actor
- **Occurrence likelihood**: Medium
- **Impact:** High
- **Threat type:** Tampering, information disclosure
- **Potential mitigations:**
  1. Review the changes of every dependency when upgrading
  2. Refrain from using software dependencies as much as reasonably possible
  3. Use a vulnerability scanner that scans for compromised software dependencies

[Back to top ↑](#table-of-contents)

## THREAT-019: Impersonation of software dependency

- **Actor:** Cybercriminal (usually remote)
- **Purpose:** To compromise software that use the dependency
- **Target:** Software that use the dependency, user's sensitive or secret data managed by software that use the dependency
- **Action:** The actor forks or duplicates the code of a software dependency (often a popular one), modifies it, and then releases or deploys it. Developers of other software install or integrate the malicious fake dependency in their codebases, thinking it is the real dependency they were looking to use. This causes the malicious code to be included in the software released to users. The malicious code sends sensitive or secret data to the actor.
- **Result of the action:** Modification of the behavior of software that use the dependency, user's sensitive or secret data is leaked to the actor
- **Occurrence likelihood**: Medium
- **Impact:** High
- **Threat type:** Spoofing, information disclosure
- **Potential mitigations:**
  1. Review any new dependencies introduced into the codebase, especially dependencies introduced within contributed code
  2. Refrain from using software dependencies as much as reasonably possible
  3. Use a vulnerability scanner that scans for malicious software dependencies

[Back to top ↑](#table-of-contents)

## THREAT-020: User overwhelmed by too many request prompts

Similar to *[THREAT-010: Wallet connection server overwhelmed by too many requests or requests that are too large](#threat-010-wallet-connection-server-overwhelmed-by-too-many-requests-or-requests-that-are-too-large)*, but with focus on the user instead of the wallet connection server.

- **Actor:** Malware, malicious or malfunctioning dApp
- **Purpose:** To harass the user, to make the wallet unusable for the user, malfunction (no purpose), dApp intentionally wants user to sign a large number of transactions
- **Target:** User, functionality of the desktop wallet
- **Action:** The actor creates and sends a large number of HTTP requests to the wallet connection server that are all supposed to trigger a prompt to the user (e.g. wallet session initialization prompt, sign transaction prompt). The wallet connection server forwards all of these requests to the UI. The UI displays a prompt for each of the received requests, which is more than what a human user can reasonably handle.
- **Result of the action:** The user is unable to use most of the desktop wallet functions because there are too many prompts in the way. This would typically result in the user being annoyed by the overwhelming number of prompts.
- **Occurrence likelihood**: Medium
- **Impact:** Medium
- **Threat type:** Denial of service
- **Potential mitigations:**
  1. Set and enforce a maximum number of outstanding prompts for the user
  2. Allow the user to turn off the wallet connection server and not accept requests from dApps, even when there are outstanding prompts
  3. Design the UI in a way that makes handling many prompts at the same time easier for the user
  4. Allow the user to switch the wallet connection server off, even when they are receiving prompts
  5. Allow the user to switch the wallet connection server on for a limited period of time. This would limit the window of opportunity the actor has in doing this type of attack.

[Back to top ↑](#table-of-contents)

## THREAT-021: Change to less secure settings

- **Actor:** User
- **Purpose:** To be able to do something that could not be done with more secure settings, to make the desktop wallet easier and more convenient to use
- **Target:** Desktop wallet settings (stored in a configuration file), desktop wallet functionality
- **Action:** Assuming the default settings are secure, the user changes the settings in a way that makes the desktop wallet less secure
- **Result of the action:** The user is exposed to more threats, the desktop wallet is possibly easier and more convenient for the user, an advanced or more dangerous feature is enabled (not a problem for a more knowledgeable user)
- **Occurrence likelihood**: High
- **Impact:** Medium
- **Threat type:** Tampering, elevation of privilege
- **Potential mitigations:**
  1. Allow the user to change certain settings where the change lasts for a limited time. After the time for the change lapses, the settings reverts back to the more secure default value. This is for the instance where the user may only need an advanced feature once or twice within a given period of time.
  2. Warn user when they change a setting to a less secure or dangerous value
  3. Warn user when a setting is currently set to a value that is less secure or dangerous
  4. Educate the user on what each setting does and how it can affect the functionality of the wallet
  5. Design the UI in a way that makes secure settings as easy and convenient as possible

[Back to top ↑](#table-of-contents)

## THREAT-022: Multiple dApps establishing connect sessions with the same dApp ID

Part of this is similar to [THREAT-001](#threat-001-impersonation-of-a-trustworthy-dapp-or-platform), but the focus is on the dApp connect feature.

When establishing a dApp connect session, the ID the dApp creates and sends to the server is *not* secret and *should* to be unique.

> [!TIP]
> Refer to the [decision about the vocabulary terms for DApp Connect](20250609-terms-for-parts-of-dapp-connect.md#decision-outcome) for short explanations of parts of DApp Connect.

- **Actor:** Malware, malicious or malfunctioning dApps
- **Purpose:** To impersonate another dApp (similar to [THREAT-001](#threat-001-impersonation-of-a-trustworthy-dapp-or-platform)), to cause the server to go into error state whenever a targeted dApp tries to connect to the wallet, to overwrite dApp information the server has stored for the dApp with the given dApp ID, malfunction (no purpose)
- **Target:** The dApp being impersonated, dApp information stored by the server, functionality of the desktop wallet
- **Action:** The actor somehow (maybe by intercepting messages) obtains the dApp ID of another dApp that is already connected to the wallet or has been connected to the wallet before. The actor then initializes a connect session by sending a request to the server using the obtained dApp ID.
- **Result of the action:** Depends on how the server is configured. The server may block the target dApp from establishing a dApp connect session because a session with the ID already exists. If configured in another way, the server overwrites the information about the target dApp stored in the server.
- **Occurrence likelihood**: Medium
- **Impact:** Medium
- **Threat type:** Spoofing, tampering, denial of service
- **Potential mitigations:**
  1. The dApp connect server must not trust that the dApp ID provided by a dApp is unique and use the dApp ID as a unique identifier for a dApp connect session. Instead, the server must use the session ID it generates as the unique identifier for a session.
  2. To establish a new dApp connect session, the dApp must *confirm* its connect session by sending an authenticated request after *initializing* the session. This way, the dApp proves to the server that it owns its dApp ID because the dApp key (along with the session ID) is required to derive the shared secret key that is needed to send authenticated requests to the server.

[Back to top ↑](#table-of-contents)

<!--
## THREAT-{{3-DIGIT_ID}}: {{Threat name}}

- **Actor:** {{Who or what instigates the attack?}}
- **Purpose:** {{What is the actor’s goal or intent?}}
- **Target:** {{What asset is the target?}}
- **Action:** {{What action does the actor perform or attempt to perform? Here you should consider both the resources and the skills of the actor. You will also be describing HOW the actor might attack your system and its expansion into misuse cases}}
- **Result of the action:** {{What happens as a result of the action? What assets are compromised? What goal has the actor achieved?}}
- **Occurrence likelihood**: {{How likely will the attacker attempt to exploit threat (high, medium, or low)}}
- **Impact:** {{What is the severity of the result (high, medium, or low)}}
- **Threat type:** {{STRIDE (e.g., denial of service, spoofing)}}
- **Potential mitigations:**
  1. {{Mitigation Technique #1}}
  2. {{Mitigation Technique #2}}
  3. {{Mitigation Technique #3}}

[Back to top ↑](#table-of-contents)
-->
