# Deprioritize UI Design and UX

- Status: accepted
- Deciders: No-Cash-7970
- Date: 2024-06-05
- Tags: roadmap, frontend

## Context and Problem Statement

The desktop wallet is a prototype that will eventually be cast aside or thrown away after getting enough useful information from it. Unfortunately, only one part-time developer is available to build this desktop wallet prototype. This is unlike most good wallets, which have so many features and components that they require a team of developers to build and manage them. The amount of developer time and effort available for building this project very limited. That precious developer time and effort needs to go to the features and components of the desktop wallet that will yield the most results in the shortest amount of time possible.

## Decision Drivers

- **Developer time and effort**: The amount of time and effort available for development is very limited
- **Only the basic features are necessary**: There are so many features commonly found in good wallet software. However, most of these features are not needed to build a prototype that yields insightful results.

## Decision Outcome

Chose to make user interface (UI) design and user experience (UX) low priorities with the desktop wallet prototype. Getting the UI and UX right takes a lot of time and effort. This time and effort is better spent on building the basic features. Although the quality of the UI and UX will likely suffer significantly, it should not get in the way of assessing the security and technical feasibility of desktop wallet prototype.

**Confidence:** High

### Positive Consequences

- More development time and effort can go to the experimentation of critical security and technical components

### Negative Consequences

- The UI would probably not look good
- The UI may be difficult for many users
- It is possible for the UI to be so terrible that it undermines security and usefulness of features

## Links

- Relates to: [No Major Upgrades for Frontend Dependencies](20250531-no-major-upgrades-for-frontend-dependencies.md)
