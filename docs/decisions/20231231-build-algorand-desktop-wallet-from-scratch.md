# Build Algorand Desktop Wallet From Scratch

- Status: superseded by [20240217-integrate-kmd-wallet-management-code](20240217-integrate-kmd-wallet-management-code.md)
- Deciders: No-Cash-7970
- Date: 2024-01-01
- Tags: roadmap

## Context and Problem Statement

There is no easy-to-use desktop wallet, that is not a website or a browser extension, for signing transactions from decentralized apps (DApps) in Algorand that fully support most of Algorand's features.

## Decision Drivers

- **Security**: Website wallets and browser extension wallets are not the most secure for holding a user's private keys, but they are the only wallet options for desktop that can interact with DApps in Algorand. A browser extension wallet is typically more secure than a website wallet. The MyAlgo (a website wallet) data breach of user private keys proved how insecure website wallets can be and how devastating such a breach can be.
- **Longevity:** A website wallet can be shut down and be inaccessible to the user. While browser extension wallets typically do not have this same problem, a user could lose the ability to download and and install the extension if the extension is closed source and removed from the marketplace (like the Chrome Store).
- **Cross-Platform:** A wallet should be able to be used by most desktop users, most of whom are using either Linux, Mac or Windows operating systems.
- **Developer Ambition:** The developer (No-Cash-7970) has been a full-stack web developer for years and wants to try building something that is not a website.

## Considered Options

1. Build a desktop wallet that is not a website or browser extension from scratch
2. Create a graphical user interface (GUI) wrapper for Algorand's Key Management Daemon (KMD)
3. Do nothing

## Decision Outcome

Chose to build a desktop wallet from scratch, because if built correctly, the desktop wallet from scratch would be the most secure and possibly longest-lasting solution while giving the developer a chance to build something that is not a website. However, some of KMD's designs can be used as a guide for building a desktop wallet from scratch.

**Confidence:** Medium. Building a desktop wallet should be feasible with a reasonable amount of time and effort. However, despite there being some architecture design ideas to achieve the goals for the desktop app, it is unclear if those design ideas will pan out.

## Pros and Cons of the Options

### Build Desktop Wallet from Scratch

Build a wallet for desktop that is not a website or a browser extension. The user would download and install the desktop wallet app. The user's keys would stored only on their computer.

- Pro: If built correctly, a desktop wallet that does not depend on the user's browser would be the most secure wallet option for desktop.
- Pro: Desktop wallet app that is downloaded and installed allows the user more control over the wallet app and more privacy.
- Pro: This would be a challenge that provides the developer an opportunity to build something potentially useful that is not a website or trivial tutorial.
- Con: Risky. If the wallet is built insecurely, there could be a massive mess like the MyAlgo breach.
- Con: Among the options, this requires the most time, effort and research to do well. The ways for users to utilize Algorand's more esoteric and advanced features would have to built from scratch using an Algorand SDK.

### Create a GUI wrapper for KMD

Create a GUI that gets input from the user and calls KMD using that input. The GUI would merely serve as a more friendly interface for KMD.

- Pro: Would be able to take advantage of KMD's security that was developed by the Algorand team with less effort.
- Pro: Utilizing KMD would make it easier to support almost all of Algorand's features.
- Pro: It is a challenge that provides the developer an opportunity to build something potentially useful that is not a website or trivial tutorial.
- Con: KMD does not work very well with Windows without using Docker and the Windows Subsystem for Linux (WSL) because the Algorand node software was built for and tested on \*nix operating systems (Mac and Linux). It may not be possible to get a working solution for Windows that does not require the user to download and set up Docker and WSL using this option.
- Con: It is difficult to separate the `kmd` daemon from the [go-algorand](https://github.com/algorand/go-algorand) code base into a standalone tool that does not include unnecessary components such as `algod` and `goal`.

### Do Nothing

Do nothing and let the problem remain as it is. Some might say this is the smartest option.

- Pro: The safest, easiest and most comfortable option. Maybe someone else more capable would develop a solution (and a possibly a better one) to the problem.
- Con: Does not solve the problem. For some reason, no one seems to have taken up solving the problem yet despite there being a demand for a solution. It is possible that no one will ever try to solve this problem because it is deemed too difficult, risky and unprofitable.
- Con: No challenge for the developer to use as an opportunity.

## Links

- Superseded by [Integrate KMD Wallet Management Code](20240217-integrate-kmd-wallet-management-code.md)
- [List of Algorand tools and infrastructure, which includes wallets](https://algorand.co/ecosystem/infrastructure-tools)
- [Incident Report about the 2023 MyAlgo data breach](https://github.com/HalbornSecurity/PublicReports/blob/master/Incident%20Reports/RandLabs_MyAlgo_Wallet_Executive_Summary_Halborn%20.pdf)
- [List of Algorand wallets (outdated)](https://developer.algorand.org/ecosystem-projects/?tags=wallets)
