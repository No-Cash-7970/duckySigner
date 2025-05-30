# Use Echo Web Framework

- Status: accepted
- Deciders: No-Cash-7970
- Date: 2024-01-03
- Tags: frameworks, backend, wallet-connection

## Context and Problem Statement

Which Go web framework should be used to create the local wallet connection server? The server should not receive a large number of requests and would need to be able to handle a connection that lasts for several minutes (while waiting for user to respond).

## Decision Drivers

- **Ease of use:** Prefer a framework that has a short learning curve and is flexible enough to accommodate fluctuating requirements
- **Feature set:** Prefer a framework with a wide array features, but looking for features pertaining to persistent connections and security
- **Documentation and help:** Prefer a framework with clear documentation and an active community that can help with questions
- **Free and open source:** Must be a framework where the code can be examined and the license allows for commercial use
- **Lightweight:** Prefer a framework that is lightweight so it consumes a little of the user's computer resources as possible

## Considered Options

- Echo
- Fiber
- Chi
- None (Only using Go's `net/http`)

## Decision Outcome

Chose Echo because of its included HTTP/2 and web sockets support with a large number of useful examples in the documentation.

**Confidence**: Low. Because what is needed to create the local server for the wallet app is uncertain, it is unclear if Echo is the best option.

## Pros and Cons of the Options

### Echo

Echo is a "high performance, extensible, minimalist" web framework that is one of the most popular options in Go.

Website: <https://echo.labstack.com/>

- Pro: Flexible
- Pro: Well-written documentation with plenty of examples
- Pro: HTTP/2 support, which would allow for a the server to wait minutes for a response
- Pro: Easy to get started
- Pro: WebSocket support

### Fiber

Fiber is one of the fastest web frameworks for Go because it is built on top of Fasthttp. The syntax for defining routes is like Express.js.

Website: <https://gofiber.io/>

- Pro: Fast
- Pro: Well-written documentation with plenty of examples
- Pro: The Express-style syntax is more familiar to the developer (No-Cash-7970)
- Con: No included HTTP/2 support

### Chi

Chi is one of the older web frameworks for Go. It has many features for supporting a small and large systems.

Website: <https://go-chi.io/>

- Pro: Flexible
- Pro: Well-written documentation with plenty of examples
- Pro: HTTP/2 support
- Pro: Easy to get started

### None (Only Using Go's `net/http`)

Go's built-in `net/http` package is capable of a lot and serves as the foundation of most Go web frameworks.

Website: <https://pkg.go.dev/net/http>

- Pro: Very flexible
- Pro: HTTP/2 support
- Pro: Large amount of documentation
- Pro: Nothing extra to import
- Con: Requires more time and effort to use because various features would have to be built by hand

## Links

- Relates to [Build Algorand Desktop Wallet From Scratch](20231231-build-algorand-desktop-wallet-from-scratch.md)
- Relates to [Build Using Go and TypeScript](20240101-build-using-go-and-typescript.md)
- Relates to [Use Local Server to Connect to DApps](20240102-use-local-server-to-connect-to-dapps)
