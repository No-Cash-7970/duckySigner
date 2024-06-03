# Use Svelte as Frontend Framework

- Status: draft
- Deciders: No-Cash-7970
- Date: 2024-06-03

## Context and Problem Statement

A JavaScript/TypeScript framework is typically used to build a frontend user interface (UI). For this desktop wallet project, using the right framework should make development of the UI components as quick and easy as possible.

## Decision Drivers

- **Compatibility with Wails v3-alpha**: The TypeScript framework to be used for the frontend should work well with Wails.
- **Amount of time needed to learn**: Prefer a framework that does not need a large amount of time (more than 1 week) to be proficient enough to build the frontend functionality
- **Support for TypeScript**: The framework should be able to be used with TypeScript

## Considered Options

- ~~Next.js~~
- Svelte

## Decision Outcome

Chose Svelte after trying Next.js and Svelte. Despite the developer's familiarity with Next.js, it did not work well with Wails and is simply not the right tool for the job when it comes to this project. Fortunately, Svelte is fairly easy to learn. It is one of the most recommended JavaScript/TypeScript frameworks in the Wails community.

**Confidence:** Medium. The developer for this project (No-Cash-7970) has never used Svelte for a sizable project before. However, there are many in the Wails community who have claimed to have use Svelte and have gotten desirable results.

## Pros and Cons of the Options

### Next.js

Tried building a basic UI using Next.js with Wails because of the developer's familiarity with it. Next.js did not function well with Wails. The compilation of the code broke most of the time because of Next.js build errors. In the instances where the build does not break, the UI breaks from the awkward and inefficient handling of state and "hydration" issues caused by the Wails TypeScript bindings. **Next.js is no longer an option because of these issues.**

### Svelte

Tried building a basic UI using Svelte. This is the same UI that did not work when built with [Next.js](#nextjs). Svelte worked well with Wails without any problems.

## Links

- Relates to [Build Algorand Desktop Wallet](20231231-build-algorand-desktop-wallet.md)
- Relates to [Use Wails UI Framework](20240101-use-wails-ui-framework.md)
- Relates to [Build Using Go and TypeScript](20240101-build-using-go-and-typescript.md)
