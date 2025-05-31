# No Major Upgrades for Frontend Dependencies

- Status: draft
- Deciders: No-Cash-7970
- Date: 2025-05-31
- Tags: roadmap, frontend, dev-process

## Context and Problem Statement

Since beginning this project, major version of some of the frontend dependencies were released. Upgrading to these major version would take a lot of time and effort because much of the code would have to be refactored.

## Decision Drivers

- **Developer time and effort**: The amount of time and effort available for development is very limited and needs to go to the most important parts of the project

## Decision Outcome

Chose to do no upgrades to major versions of the frontend dependencies. Frontend dependencies will only be upgraded to the next minor or patch version. The only possible exception is the Wails runtime JavaScript library.

### Positive Consequences

- Almost no time and effort is needed to upgrade dependencies to their next minor or patch version
- Allows for a more stable development workflow that focuses on the development of more important features

### Negative Consequences

- Miss out features that may make development easier
- If the code becomes old enough, it may be incompatible with more recent environments (operating systems, browsers, etc.)

## Links

- Relates to: [Deprioritize UI Design and UX](20240602-deprioritize-ui-design-and-ux.md)
