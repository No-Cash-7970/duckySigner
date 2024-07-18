# Create and Maintain C4 Diagrams

- Status: accepted
- Deciders: No-Cash-7970
- Date: 2024-02-20
- Tags: dev-tools, dev-process, doc

## Context and Problem Statement

The software architecture of the [desktop wallet](20231231-build-algorand-desktop-wallet-from-scratch.md) may be unusual and difficult to understand. Communicating the architecture in a manner that is easier to understand is key to making it easier for anyone to understand the system and provide useful input.

## Decision Drivers

- **Understandability:** The top level (system view) diagram should be able to be understood by people who have little or no software engineering knowledge. The lower-level diagrams showing more details about the software system should be descriptive enough to communicate what is happening while being able to be understood by those with software engineering knowledge but little or no knowledge about this project specifically.
- **Ease of modifying and updating diagrams:** The diagrams should be able to be easily updated as the system changes
- **Ability to sufficiently represent the system at multiple levels:** The diagrams should be able to represent multiple parts of the system at varying levels of abstraction (from system level to class component level)

## Considered Options

- UML (Unified Modeling Language)
- C4 (Context, Containers, Components, Code)

## Decision Outcome

Chose to use C4 diagrams to represent the software architecture. This decision comes after trying C4 diagrams out using [Structurizr](https://structurizr.com/dsl). The diagrams should be updated as the system changes. Their visual graphic form should be accessible to the public.

**Confidence**: Medium. Creating and maintaining the diagrams should be useful, but it is not clear if the extra time and effort required is worth it.

### Positive Consequences

- More transparency through better communication of how the desktop wallet works
- Can be a great help when designing the software architecture
- Can help with threat modeling

### Negative Consequences

- More stuff to create and maintain, which takes away time and effort that can go to building and maintaining the software

## Pros and Cons of the Options

### UML

The Unified Modeling Language (UML) has been around for decades. It is a huge and flexible modeling language has been used frequently by a variety of projects over the years. However, the use of UML has declined in recent years in favor of more lightweight (and perhaps less organized) methods of diagraming systems which has resulted in newer developers not being familiar with UML.

- Pro: Has many options and tools because of being well-established in software engineering
- Con: UML diagrams can be clunky or contain too many details that make them difficult to read and understand

### C4

C4 is a lightweight modeling system for software architecture with 4 levels of abstraction: Context, Containers, Components, and Code. It was created in 2006, which makes it fairly new to the software engineering world.

- Pro: C4 diagrams tend to be simpler to read and understand than UML
- Pro: The diagrams can be specified in a text files and committed to a Git repository
- Pro: Can be combined with UML for more detailed diagrams
- Con: Not as many tools for it because it is not a very well-known modeling system

## Links

- Relates to [Build Algorand Desktop Wallet From Scratch](20231231-build-algorand-desktop-wallet-from-scratch.md)
- [Unified Modeling Language - Wikipedia](https://en.wikipedia.org/wiki/Unified_Modeling_Language)
- [C4 model - Wikipedia](https://en.wikipedia.org/wiki/C4_model)
- [The C4 model for visualising software architecture](https://c4model.com/)
