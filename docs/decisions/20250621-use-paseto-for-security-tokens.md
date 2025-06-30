# Use PASETO for Security Tokens

- Status: accepted
- Deciders: No-Cash-7970
- Date: 2025-06-30
- Tags: dapp-connect, security, backend

## Context and Problem Statement

The [DApp connect protocol](20250619-dapp-connect-protocol.md) requires the server to issue security tokens to dApps. Each dApp eventually presents their token to the server. The tokens need to be in a format that enables the dApp connect protocol to be secure when the token is transferred over an insecure localhost connection.

## Decision Drivers

- **Security:** The token format should not be able to be cracked by any entity besides the issuer of the token (usually the server), thus making it safe to use on [localhost, where there is no SSL/TLS](20240102-use-local-server-to-connect-to-dapps.md)
- **Library in Go:** The security token must be able to be easily created in Go, the language used for the server. The library for the security token format must be maintained and up-to-date.
- **Library in JavaScript/TypeScript and other languages:** Although it is unlikely that dApps will need to create their own security tokens, allow for it to be a possibility. DApps could be built in a variety of languages.

## Considered Options

- PASETO
- JWT

## Decision Outcome

Chose PASETO. It is easier to use because it requires less effort to use it securely and transmit it safely through localhost without SSL/TLS.

**Confidence:** Medium. Both the server and dApps should be able to handle PASETOs. However, PASETO is relatively unknown in the web development world.

## Pros and Cons of the Options

### PASETO

[PASETO (Platform-Agnostic SEcurity TOkens)](https://paseto.io/) (pronunciation: pah-SEH-toh) is a newer security token format that was created in 2018 as a [more secure replacement for JWT](https://paragonie.com/blog/2018/03/paseto-platform-agnostic-security-tokens-is-secure-alternative-jose-standards-jwt-etc).

- Pro: Less error-prone, which makes it easier to use securely
- Pro: Most of the libraries implement PASETO consistently because of how strict it is
- Con: Tooling and literature for PASETO is limited because it is newer and less common

### JWT

[JWT (JSON Web Token)](https://jwt.io/introduction) (pronounced like "jot") has been around since 2010 and is often used for [OAuth 2.0](https://oauth.net/2/).

- Pro: The popularity of JWT means that there are plenty of tools and literature for JWT, which makes it easier to find help regarding JWT
- Con: Transmitting JWT without SSL/TLS is [not recommended](https://snyk.io/blog/top-3-security-best-practices-for-handling-jwts/), which makes it difficult to use securely on localhost
- Con: The overwhelming number of options for JWT make it difficult to use. JWT's level of security can vary greatly depending on the library that is used to create a token.

## Links

- Refines [DApp Connect Protocol](20250619-dapp-connect-protocol.md)
- Relates to [Vocabulary for DApp Connect](20250621-vocab-for-dapp-connect.md)
- [Platform-Agnostic SEcurity TOkens (PASETO)](https://paseto.io/)
- [A Thorough Introduction to PASETO - Okta Developer](https://developer.okta.com/blog/2019/10/17/a-thorough-introduction-to-paseto)
- [Paseto is a Secure Alternative to the JOSE Standards (JWT, etc.) - Paragon Initiative](https://paragonie.com/blog/2018/03/paseto-platform-agnostic-security-tokens-is-secure-alternative-jose-standards-jwt-etc)
- [JSON Web Token (JWT)](https://jwt.io/introduction)
- [No Way, JOSE! Javascript Object Signing and Encryption is a Bad Standard That Everyone Should Avoid - Paragon Initiative](https://paragonie.com/blog/2017/03/jwt-json-web-tokens-is-bad-standard-that-everyone-should-avoid)
- [Top 3 security best practices for handling JWTs - Snyk](https://snyk.io/blog/top-3-security-best-practices-for-handling-jwts/)
- [Attacking and Securing JWT - OWASP (PDF)](https://owasp.org/www-chapter-vancouver/assets/presentations/2020-01_Attacking_and_Securing_JWT.pdf)
