# Document Threat Model

- Status: draft
- Deciders: No-Cash-7970
- Date: 2024-02-19

## Context and Problem Statement

A more formal process for considering and mitigating security concerns is crucial for designing, building and maintaining a desktop wallet. The process for designing for security should be effective and transparent.

## Decision Drivers

- **Security:** Should be a process that enables an easy and maintainable way of communicating and mitigating security concerns.
- **Transparency:** Security concerns should be documented and available to the public, along with how those security concerns are addressed.

## Decision Outcome

Chose to document the threat model as a Markdown document. This document should be maintained as the system changes and more information is known. Threat modeling should be done regularly, perhaps for every release with a version bump.

**Confidence**: High. Creating and maintaining documentation of the threat model will likely be proven to be essential for this use case (an Algorand desktop wallet).

### Positive Consequences

- Easier for the other developers, security professionals or end-users to evaluate the security risks of the software and what is being done about those risks.

### Negative Consequences

- Requires more time and effort, which is less time and effort for building the software

## Links

- Relates to [Build Algorand Desktop Wallet](20231231-build-algorand-desktop-wallet.md)
- [Threat Modeling Cheat Sheet - OWASP](https://cheatsheetseries.owasp.org/cheatsheets/Threat_Modeling_Cheat_Sheet.html)
- [Threat Modeling Process](https://owasp.org/www-community/Threat_Modeling_Process)
- [Threat Modeling in Practice](https://owasp.org/www-project-developer-guide/draft/design/threat_modeling/practical_threat_modeling/)
- [Threat modeling using C4 diagrams](https://medium.com/flat-pack-tech/threat-modeling-as-code-f3555f5d9024)
