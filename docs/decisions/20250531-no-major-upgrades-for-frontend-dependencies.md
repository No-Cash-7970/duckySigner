# No Major Upgrades for Frontend Dependencies

- Status: accepted
- Deciders: No-Cash-7970
- Date: 2025-06-05
- Tags: roadmap, frontend, dev-process

## Context and Problem Statement

Since starting this project, major version upgrades of some of the frontend dependencies were released. Doing these upgrades would take a lot of time and effort because much of the code would have to be refactored.

## Decision Drivers

- **Developer time and effort**: The amount of time and effort available for development is very limited and needs to go to the most important parts of the project
- **Project purpose**: This project is just a prototype that will be thrown away eventually. It is not meant to be a long-term project that is maintained indefinitely.

## Decision Outcome

Chose to make *not* upgrading to later major versions of the frontend dependencies the policy for this project. Frontend dependencies will only be upgraded to later minor or patch versions. The only possible exception is the Wails runtime JavaScript library.

### Positive Consequences

- Almost no time and effort is needed to upgrade dependencies to a later minor or patch version
- Allows for a more stable development workflow that enables more focus on the development of more important features

### Negative Consequences

- May miss out on features that make development easier
- If the code becomes old enough, it may become incompatible with newer environments (operating systems, browsers, etc.)

## Links

- Relates to: [Deprioritize UI Design and UX](20240602-deprioritize-ui-design-and-ux.md)
