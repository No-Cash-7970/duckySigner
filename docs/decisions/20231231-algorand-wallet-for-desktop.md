# Algorand Wallet for Desktop

- Status: draft
- Deciders: No-Cash-7970
- Date: 2023-12-31

## Context and Problem Statement

There is no easy-to-use desktop wallet, that is not a website or a browser extension, for signing transactions from decentralized apps (DApps) for Algorand that fully support most of Algorand's features.

## Decision Drivers

- **Security:** Website wallets and browser extension wallets are not very secure for holding a user's private keys, but are the only type of wallet options for Algorand. The MyAlgo data breach of private keys proved how insecure website wallets can be and how devastating such a breach can be.
- **Longevity:** A website wallet can be shut down and be inaccessible to the user. Browser wallets typically do not have this same problem.
- **Cross-Platform:** The wallet would be able to be used by most desktop users, whom most are using either Linux, Mac, or Windows operating systems.
- **Developer Ambition:** The developer (No-Cash-7970) has been a full-stack web developer for years and wants to try building something that is not a website.

## Considered Options

- Build a desktop wallet that is not a website or browser extension from scratch
- Create a graphical user interface (GUI) for Algorand's Key Management Daemon (KMD)
- Do nothing

## Decision Outcome

Chose to build a desktop wallet from scratch", because if built correctly, the desktop wallet from scratch would be the most secure and possibly longest-lasting solution while giving the developer a chance to build something that is not a website. However, some of KMD's designs can be used as a guide for building a desktop wallet from scratch.

## Pros and Cons of the Options

### Build Desktop Wallet from Scratch

Build a wallet for desktop that is not a website or a browser extension. The user would download and install the desktop wallet software with the keys stored on the user's computer and without using a third party, such as Wallet Connect, to connect to DApps.

- Pro: If built correctly, a desktop wallet that does not depend on the user's browser would be the most secure wallet option for desktop.
- Pro: Desktop wallet software that is downloaded and installed allows the user more control over the wallet software and more privacy.
- Pro: It is a challenge that provides the developer an opportunity to build something potentially useful that is not a website or a trivial tutorial.
- Con: Risky. If the wallet is built insecurely, there could be a massive mess like the MyAlgo breach.
- Con: Among the options, this requires the most time, effort, and research to do it well. Ways for users to utilize Algorand's more esoteric and advanced features would have to built from scratch.

### Create a GUI for KMD

Create a GUI that get input from the user and calls KMD using the input. In this way, the GUI would server merely as a more friendly interface for KMD.

- Pro: Would be able to take advantage of KMD's security that was developed by the Algorand team with less effort.
- Pro: Utilizing KMD would make it easier to support almost all of Algorand's features.
- Pro: It is a challenge that provides the developer an opportunity to build something potentially useful that is not a website or a trivial tutorial.
- Con: KMD does not work very well with Windows because the Algorand node software was built for *nix operating systems (Mac and Linux). It may not be possible to get a working solution for Windows using this option.
- Con: It is difficult to separate the `kmd` daemon from the [go-algorand](https://github.com/algorand/go-algorand) code base into a standalone tool that does not include unnecessary components such as `algod` and `goal`.

### Do Nothing

Do nothing and let the problem remain as it is.

- Pro: The safest, easiest and most comfortable option. Maybe someone else more capable would develop a solution (possibly a better one) to the problem.
- Con: Does not solve the problem. For some reason, no one seems to have taken up solving the problem despite there being a demand for it. It is possible that no one will ever try to solve this problem because it is deemed too difficult, risky and unprofitable.
- Con: No challenge for the developer to use as an opportunity.

## Links

- [List of Algorand Wallets](https://developer.algorand.org/ecosystem-projects/?tags=wallets)
- [Incident Report about the 2023 MyAlgo data breach](https://github.com/HalbornSecurity/PublicReports/blob/master/Incident%20Reports/RandLabs_MyAlgo_Wallet_Executive_Summary_Halborn%20.pdf)
