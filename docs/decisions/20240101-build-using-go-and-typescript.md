# Build Using Go and TypeScript

- Status: accepted
- Deciders: No-Cash-7970
- Date: 2024-01-02

## Context and Problem Statement

Which programming language(s) should be used to implement the Algorand wallet for desktop? May need to use multiple languages.

## Decision Drivers

- **Available help and documentation:** Prefer a language where there is plenty of documentation and a large community to go to for help
- **Free and open source tools and libraries:** Prefer a language where there are libraries and tools for building a production-ready desktop graphical user interface (GUI) while providing as much flexibility as possible
- **Easy to build for multiple platforms:** Prefer a language that can be used to build for as many computer platforms and operating systems as possible, especially Linux, Mac and Windows
- **Amount of time needed to learn:** Prefer a language that does not need a large amount of time (2 weeks or more) to be proficient enough to build a functioning and modern-looking desktop app
- **Developer's familiarity with other languages:** Prefer to use languages that the developer (No-Cash-7970) is familiar with (besides Go) like TypeScript, Python, and PHP
- **Developer's interests:** The developer wants to learn Go and make something with it because Go is becoming one of the most popular languages

## Considered Options

- Go
- TypeScript/JavaScript
- Rust
- Python

## Decision Outcome

Chose to use Go with TypeScript/JavaScript because Go is an easy-to-learn language that can be easily compiled for the multiple operating systems. With [Wails](https://wails.io/), the flexibility of JavaScript and TypeScript can be utilized to build a user interface (UI) for a backend built in Go that can better utilize computing resources in a manner that looks native to the user's operating system.

**Confidence**: Medium. It is not clear if the Go and TypeScript/JavaScript combination will allow for the desired results.

## Pros and Cons of the Options

### Go

- Pro: Developer wants to learn it
- Pro: Easy to get started
- Pro: Cross-platform support by default
- Pro: Popular language with many libraries, tools and tutorials
- Pro: May be able to borrow code from Algorand's Key Management Daemon (KMD)
- Con: Will take some time to learn
- Con: Not commonly known for being used to develop desktop apps

### TypeScript/JavaScript

- Pro: Developer is proficient in using JavaScript and TypeScript
- Pro: Cross-platform, with a browser or Node.js
- Con: Pure TypeScript/JavaScript framework options for building cross-platform desktop apps, like [Electron](https://www.electronjs.org/), tend to be resource hogs because they embed a full browser

### Rust

- Pro: It is a performant language
- Pro: Cross-platform by default
- Con: Requires a lot of time to learn proficiently (at least a month)
- Con: Number of tools and libraries are limited, although the community and the tooling is growing
- Con: Developer doesn't care to learn it

### Python

- Pro: Developer is familiar with Python
- Con: Difficult to cross-compile

## Links

- Relates to [Use Wails UI Framework](20240101-use-wails-ui-framework.md)
