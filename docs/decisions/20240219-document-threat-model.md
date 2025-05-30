# Document the Threat Model

- Status: accepted
- Deciders: No-Cash-7970
- Date: 2024-02-21
- Tags: dev-process, docs, security

## Context and Problem Statement

A more formal process for addressing and mitigating security concerns is crucial for designing, building and maintaining a desktop wallet application. The process should be effective and transparent.

## Decision Drivers

- **Security:** There should be a process that enables an easy and maintainable way of addressing, communicating and mitigating security concerns
- **Transparency:** Addressed and mitigated security concerns should be documented and available to the public with information about the mitigations

## Decision Outcome

Chose to document the threat model as a collection of Markdown documents. The documents should be maintained as the system changes and more information is known. Threat modeling and updates to the threat model documents should be done regularly, perhaps for every release with a version bump.

**Confidence**: High. Creating and maintaining documentation of the threat model will likely be proven to be essential for this use case (an Algorand desktop wallet application).

### Positive Consequences

- Easier for other developers, security professionals or end-users to evaluate the security risks of the software and learn what is being done about those risks.

### Negative Consequences

- Requires more time and effort, which is less time and effort for building and maintaining the software
- It is possible to compromise security by putting too much information into the threat model documents. Information about an unmitigated threat should NEVER be published for the public to see.

## Links

- [Threat Modeling Cheat Sheet - OWASP](https://cheatsheetseries.owasp.org/cheatsheets/Threat_Modeling_Cheat_Sheet.html)
- [Threat Modeling Process - OWASP](https://owasp.org/www-community/Threat_Modeling_Process)
- [Threat Modeling in Practice - OWASP](https://owasp.org/www-project-developer-guide/draft/design/threat_modeling/practical_threat_modeling/)
- [Threat modeling using C4 diagrams](https://medium.com/flat-pack-tech/threat-modeling-as-code-f3555f5d9024)
