# Use Hawk for DApp Connection Authentication and Authorization

- Status: accepted
- Deciders: No-Cash-7970
- Date: 2024-09-30
- Tags: wallet-connection, security

## Context and Problem Statement

For the [dApp connection server](20240102-use-local-server-to-connect-to-dapps.md), which authentication or authorization scheme should be used?

## Decision Drivers

- **Security:** The authentication/authorization solution should mitigate security threats [THREAT-001: Impersonation of a trustworthy dApp or platform](../threat-model/01-threats.md#threat-001-impersonation-of-a-trustworthy-dapp-or-platform) and [THREAT-009: Interception of HTTP communication between dApp and wallet connection server](../threat-model/01-threats.md#threat-009-interception-of-http-communication-between-dapp-and-wallet-connection-server).
- **Ability to be used for localhost:** Localhost does not behave exactly like a normal web server connection. One of the most notable differences is the efficacy of using SSL/TLS. Using SSL/TLS on localhost is largely useless.
- **DApp developer experience:** DApps come in a variety of shapes and sizes, and not all of them may be web-based in a web browser. The goal is to maintain a satisfying developer experience for dApps of all kinds when using this desktop wallet. This means making the integration of the desktop wallet as simple and easy as possible for as many platforms as possible.
- **Flexibility:** Allow for software other than a browser to connect to and use the wallet. Also allow for the dApp connection server to separated from local user's computer and placed into a global web server.

## Considered Options

- Hawk
- API key
- OAuth 2.0
- Cookie
- Nothing

## Decision Outcome

Chose Hawk. Hawk is a solution that in somewhere between OAuth 2.0 and API key authentication. It has most of the powerful features that OAuth 2.0 has with the straightforwardness of API key authentication.

**Confidence:** Very low. Using Hawk may be overkill and not very effective for an HTTP connection that is on localhost. It is also possible that Hawk is not a long-lasting solution and would not be able to adjust to the changing security landscape because of the lack of development.

## Pros and Cons of the Options

### Hawk

Hawk was designed to be an alternative to OAuth 2.0 that is easier for developers to use.

- Pro: Provides straightforward and flexible schemes for authentication/authorization
- Pro: Support using "scopes" that would restrict what a connected dApp is allowed to as for in the wallet
- Con: Not very well known, so there are not that many up-to-date tools and libraries that support Hawk
- Con: There will no further maintenance at this time because the protocol is considered "complete"

### API Key Authentication

API key authentication is a simple method for authentication where the dApp is given a key that it must provide to make an authenticated request.

- Pro: Method used by the original KMD code, so it would be easy to implement (copy & paste)
- Pro: Simple to understand, which makes it fairly easy to maintain
- Con: Requires user to enter the API key into dApp. This would greatly degrade user experience, possibly to the point where it compromises security.

### OAuth 2.0

OAuth 2.0 is one of the most common tools for authentication/authorization. It is the mechanism used to create the "Log in with XXX" methods that make it easier and safer for users to log in without entering a password.

- Pro: A commonly used method for authentication and authorization, so there are plenty of tools and libraries that support OAuth 2.0
- Pro: Provides dApps with more flexibility in storing wallet connection session data
- Pro: Support using "scopes" that would restrict what a connected dApp is allowed to ask for in the wallet
- Con: Heavily relies on SSL/TLS for security, which is not ideal for a localhost connection
- Con: Does not provide a straightforward scheme for authentication/authorization, so using a third-party authentication/authorization service is typically encouraged

### Cookie

Using a cookie to store wallet connection session data. The data is stored on the server by the wallet and within the browser (in the form of a cookie) by the dApp.

- Pro: A tried-and-true method that is sometimes considered more secure than "stateless" methods like OAuth 2.0 and API key.
- Con: Browser only

### Nothing

No authentication or authorization scheme is used. The connection between the dApp and the dApp connection server is completely unsecured.

- Pro: Extremely easy to implement and maintain
- Con: Provides no reliable way of restricting dApps in any way (e.g. permissions, length of time allowed to connect to wallet)

## Links

- Relates to [Use Local Server to Connect to DApps](20240102-use-local-server-to-connect-to-dapps.md)
- Relates to: [Use "DApp" in Prose and "Dapp" in Code](20250608-use-dapp-in-prose-and-dapp-in-code.md)
