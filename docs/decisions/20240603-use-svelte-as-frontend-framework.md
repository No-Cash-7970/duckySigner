# Use Svelte as Frontend Framework

- Status: accepted
- Deciders: No-Cash-7970
- Date: 2024-06-06
- Tags: frameworks, frontend

## Context and Problem Statement

A JavaScript/TypeScript framework is typically used to make building a user interface (UI) easier. As such, using the right framework should make development of the UI components as quick and easy as possible.

## Decision Drivers

- **Compatibility with Wails 3.0**: The TypeScript framework to be used for the frontend should work well with Wails
- **Amount of time needed to learn**: Prefer a framework that does not need a large amount of time (more than 1 week) to be proficient enough to build the frontend functionality
- **Support for TypeScript**: The framework should be able to be used with TypeScript

## Considered Options

- ~~Next.js~~
- Svelte

## Decision Outcome

Chose Svelte after trying Next.js and Svelte. Despite the developer's familiarity with Next.js, it did not work well with Wails and is simply not the right tool for the job when it comes to this project. Fortunately, Svelte is fairly easy to learn. It is one of the most recommended JavaScript/TypeScript frameworks in the Wails community.

**Confidence:** (2025-07-16) High. Using Svelte 4 has yielded desirable results. Unit testing the fronted with Svelte has also worked well. There are no plans to upgrade to Svelte 5, which is supposed to have a number of improvements over Svelte 4.

~~**Confidence:** Medium. The developer for this project (No-Cash-7970) has never used Svelte for a sizable project before. However, there are many in the Wails community who have claimed to have used Svelte and have gotten desirable results.~~

## Pros and Cons of the Options

### Next.js

The developer initially attempted to build a basic UI for the desktop wallet app using Next.js because of her familiarity with it. Unfortunately, it did not function well with Wails. The compilation of the code broke most of the time because of Next.js build errors. In the instances where the build did not break, the UI broke from the awkward and inefficient handling of state and the "hydration" issues caused by the Wails JavaScript/TypeScript bindings. In short, **Next.js is no longer an option because of these issues.**

### Svelte

After the failed attempt with Next.js, the developer tried building a basic UI for the desktop wallet app using Svelte. This was the same UI that did not work when built with [Next.js](#nextjs). Fortunately, Svelte worked well with Wails without any problems.

## Links

- Relates to [Use Wails UI Framework](20240101-use-wails-ui-framework.md)
- Relates to [Build Using Go and TypeScript](20240101-build-using-go-and-typescript.md)
