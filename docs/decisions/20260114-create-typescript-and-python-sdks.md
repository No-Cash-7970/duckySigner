# Create Typescript and Python SDKs

- Status: accepted
- Deciders: No-Cash-7970
- Date: 2026-01-15
- Tags: dapp-connect

## Context and Problem Statement

A dApp needs to connect to the desktop wallet and manage its connect sessions with the wallet. A software development kit (SDK) would make it easier for dApp developers to use the desktop wallet in their projects. Ideally, a developer would use the SDK in the programming language they are using to develop their dApp. However, it is not possible to create an SDK for every language. A decision needs to be made on how many SDKs to create and which programming languages to create an SDK for.

## Decision Drivers

- **Typical use of the language**: Most languages have a wide range of uses. However, in practice, a language's use is typically far narrower than what it is capable of. For example, JavaScript and TypeScript are most frequently used for font-end web development, Python is frequently used for data analysis, and Go is frequently used for back-end web development.
- **Popularity of the language**: How many developers use the language? It is more worthwhile to create an SDK for a language that many developers use.
- **Developer's familiarity with the language**: The developer of this project (No-Cash-7970) will not create an SDK for a programming language she is not familiar with
- **Developer time and resources**: The developer has limited time and resources for creating and maintaining SDKs. At most 3 SDKs can be created.

## Considered Options

- TypeScript
- Python
- Go

## Decision Outcome

Chose to create only 2 SDKs: TypeScript and Python. Typescript and Python are among the most popular languages and the developer (No-Cash-7970) has some familiarity with both languages. These two languages are enough to test and demonstrate how a both web-based dApps and non-web-based dApps can connect to and use the desktop wallet.

**Confidence**: High. The Typescript and Python SDKs have been implemented. However, no example dApps have been created for either SDK.

## Pros and Cons of the Options

### TypeScript

Typescript is a very popular language for front-end web development. An SDK for TypeScript can also be used for plain JavaScript as well.

- Pro: Most dApps are web based and written in TypeScript or JavaScript
- Pro: Typescript can be compiled into Javascript that can be run in any modern web browser

### Python

Python is a very popular multipurpose language that is considered to be friendly for those learning to program.

- Pro: Often used for creating scripts and applications that run in the terminal
- Pro: The developer wants a project for practicing her Python
- Con: Not often used for building applications that have a GUI (graphical user interface)

### Go

Go is a common programming language used for a variety of high-performance applications. The back-end for the desktop wallet is written in Go.

- Pro: Go applications are often performant because of the built-in support for concurrency
- Con: While Go applications can be compiled for a wide range of machines and environments, it is still limited and a few issues can arise on certain machines
