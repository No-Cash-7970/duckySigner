# DApp Connect Protocol

- Status: accepted
- Deciders: No-Cash-7970
- Date: 2025-06-29
- Tags: dapp-connect, security, backend

## Context and Problem Statement

For the user to be able to interact with a dApp using the desktop wallet, the dApp must be able to connect with the wallet to securely send to and receive data from the wallet. For this to happen, a protocol needs to be established between the dApp and the desktop wallet's [dApp connect server on localhost](20240102-use-local-server-to-connect-to-dapps.md). In this protocol, each dApp (or dApp instance) must be able to identify itself to the server and confirm its unique identity.

## Decision Drivers

- Includes the decision drivers for the [decision to use Hawk](20240821-use-hawk-for-dapp-connection-authentication-and-authorization.md)
- **DApp connect session management:** The dApp connect protocol needs to be able to (1) establish a session between a dApp and the desktop wallet, (2) let dApps end their sessions, (3) let the server end any session, and (4) let the server receive and store information about connected dApps
- **Known related threats:** [THREAT-009](../threat-model/01-threats.md#threat-009-interception-of-http-communication-between-dapp-and-wallet-connection-server), [THREAT-010](../threat-model/01-threats.md#threat-010-wallet-connection-server-overwhelmed-by-too-many-requests-or-requests-that-are-too-large), [THREAT-012](../threat-model/01-threats.md#threat-012-modifying-security-settings-in-configuration-files), [THREAT-013](../threat-model/01-threats.md#threat-013-wallet-password-is-written-down-and-stored-in-insecure-location), [THREAT-022](../threat-model/01-threats.md#threat-022-multiple-dapps-establishing-connect-sessions-with-the-same-dapp-id)
- **User experience (UX):** Asking the user to enter their password too frequently degrades UX. Also, the user needs to be presented with relevant information for making an informed decision when approving a session and when ending a session.
- **Session data stored in encrypted file(s):** DApp connect session data being stored in an encrypted file imposes some constraints that affect UX because the user's wallet password is needed to decrypt and access the session data. This problem is lessened by the fact that the wallet password is typically stored in memory while the wallet is unlocked and open.

## Considered Options

- Iteration 4: Mitigate [THREAT-009](./threat-model/01-threats.md#threat-009-interception-of-http-communication-between-dapp-and-wallet-connection-server) by requiring a "confirmation code", revert use of approvals and session tokens, remove session renewal
- Iteration 3: Authentication using short sessions with long "approval"
- Iteration 2: Mitigate [THREAT-022](../threat-model/01-threats.md#threat-022-multiple-dapps-establishing-connect-sessions-with-the-same-dapp-id), allow session to end if server is offline, add session "ping"
- Iteration 1: Simplest iteration

## Decision Outcome

Chose Iteration 4. It strikes the right balance between security and UX.

**Confidence**: Low. Despite this protocol being inspired by other protocols like OAuth, it is uncertain if this protocol is too complex for dApps. Because this is a new protocol, there may be some unknown security vulnerabilities lurking within this protocol.

### Positive Consequences

- No secret data is transmitted unencrypted
- A dApp that is not web-based can authenticate and interact with the desktop wallet
- The user immediately knows when a dApp is trying to connect to the wallet

### Negative Consequences

- The dApp must be able to generate and store a secret key securely, which can be a problem for web dApps in a browser
- This protocol adds more parts and moving pieces to the desktop wallet, which increases the attack surface of the wallet. There may be some unknown or unforeseen threats lurking in this protocol.

## Pros and Cons of the Options

### Iteration 4

This is the fourth attempt at designing the dApp connect protocol. In this iteration, session renewal and the use of session tokens were dropped. Instead of a session token, there is a single-use session *confirmation* token that a dApp gives to the server to check the "confirmation code" when confirming a session. Parts from the [first](#iteration-1) and [second](#iteration-2) iterations were combined to make a protocol that is simpler than the [third iteration](#iteration-3).

#### Iteration 4 Session Establishment

Like with the [second](#iteration-2-session-establishment) and [third](#iteration-3-session-establishment) iterations, establishing a session consists of initializing the session and then confirming it. However, confirming a session requires the dApp to display a "confirmation code" and the user to enter the code into the wallet's UI. This mitigates the threat of the Diffie-Hellman (DH) Man-in-the-Middle (MitM) attack ([THREAT-009](../threat-model/01-threats.md#threat-009-interception-of-http-communication-between-dapp-and-wallet-connection-server)) that was not mitigated in the previous iterations.

There is no session token, unlike [Iteration 3](#iteration-3-session-establishment). Instead, there is a single-use "confirmation token". The confirmation token is used check the confirmation code entered by the user. This way, web tokens are [used in the way in which they are most suited](http://cryto.net/~joepie91/blog/2016/06/13/stop-using-jwt-for-sessions/).

1. **Initialize session**
   1. DApp: Generate dApp key pair (dApp ID & dApp secret key)
   2. DApp → Server: Send request to `/session/init`, providing dApp ID
   3. Server: Validate received dApp ID
   4. Server: Generate confirmation key pair (confirmation ID + confirmation secret key) and confirmation data (date created, expiration, etc.) and store it into an encrypted file protected by the user's wallet password
   5. Server: Generate session key pair (session ID & session secret key).
      - **Note:** Generating the session key pair and placing it into the confirmation token ensures that the confirmation token is only used once because the server cannot store multiple sessions same session ID when confirming the session. If there is an attempt to use an unexpired confirmation token after it has been used, the server will reject it because it already stores a session with the session ID within the confirmation token.
   6. Server: Generate a confirmation code
   7. Server: Use the confirmation key to create an encrypted "confirmation token" in [PASETO](https://paseto.io/) format that contains the dApp ID, session key pair, confirmation data and confirmation code
   8. Server → DApp: Send authenticated response containing confirmation ID, data, code and token
   9. DApp: Derive the "confirmation shared secret key" using dApp key and confirmation ID to verify authentication in the server's response
2. **Confirm session**
   1. DApp → User: Display confirmation code
   2. DApp → Server: Send authenticated request to `/session/confirm`, providing dApp data (dApp name, icon, etc.), confirmation token and confirmation ID
   3. Server: Retrieve confirmation key using given confirmation ID
   4. Server: Derive confirmation shared key using the confirmation key and dApp ID
   5. Server: Verify authentication and validate dApp data
   6. Server: Extract data from confirmation token by decrypting it using confirmation key
   7. Server → User: *(Through the UI)* Present dApp data and ask for approval of connection to dApp
   8. User → Server: *(Through the UI)* Approve the connection to the dApp by entering password and confirmation code
   9. Server: Generate session data (date created, expiration, etc.)
   10. Server: Derive "session shared secret key" using dApp ID and session key for sending authenticated response
   11. Server: Store session ID, session shared key, session data, dApp ID and dApp data into an encrypted file protected by the user's wallet password
   12. Server → DApp: Send authenticated response containing session ID and session data
   13. DApp: Derive the shared key using dApp key and session ID to verify authentication in the server's response
   14. DApp: Store session ID and session data

Result: The dApp can now use the session shared key to send authenticated requests to the server.

#### Iteration 4 Authenticated Request

Same as [Iteration 1](#iteration-1-authenticated-request).

#### Iteration 4 Session Termination Through DApp

A dApp can optionally end its session by contacting the server before deleting its stored session data, just like ending a session in [Iteration 1](#iteration-1-session-termination-through-dapp). This is the preferred method for a dApp to end its session. However, if the dApp cannot contact the server, the dApp simply deleting its stored session data, like with [Iteration 2](#iteration-2-session-termination-through-dapp), is acceptable.

Result: The dApp cannot successfully send authenticated requests to the server without [establishing a new session](#iteration-4-session-establishment).

#### Iteration 4 Session Termination Through Wallet

Same as [Iteration 2](#iteration-2-session-termination-through-wallet).

Result: The dApp cannot successfully send authenticated requests to the server without [establishing a new session](#iteration-4-session-establishment).

#### Iteration 4 Session Renewal

There is no session renewal. It was removed because it is unnecessary. A session does not need to be short because no secret session data is transmitted unencrypted. A session should be able to safely last about 1-2 weeks, which should be long enough to not hurt UX too much.

This dApp connect protocol is supposed to replace the wallet session protocol used by [Algorand's KMD wallet](20240217-integrate-kmd-wallet-management-code.md). With KMD, renewing sessions was needed because the "wallet handle" a dApp (or client) needed to authenticate with the server was always sent in plain text. The ability to renew a wallet handle was a compromise to make it easier to use KMD. However, that renewal feature undermined the security of the KMD wallet session because a malicious actor could steal a wallet handle while it is still valid and renew it.

#### Pros and Cons of Iteration 4

- Pro: The requirement of a confirmation code when establishing a session mitigates threat of the Diffie-Hellman (DH) Man-in-the-Middle (MitM) attack ([THREAT-009](../threat-model/01-threats.md#threat-009-interception-of-http-communication-between-dapp-and-wallet-connection-server))
- Con: Like with [Iteration 2](#pros-and-cons-of-iteration-2), sessions cannot be very long and should not be very short. Finding the right session length will require some trial-and-error.
- Con: Like with [Iteration 2](#pros-and-cons-of-iteration-2), the extra "ping" requests from dApps checking their session's validity can lead to a higher chance of the server being overwhelmed ([THREAT-010](../threat-model/01-threats.md#threat-010-wallet-connection-server-overwhelmed-by-too-many-requests-or-requests-that-are-too-large)).
- Con: More work for a dApp because it needs to provide a way to display the confirmation code to the user

### Iteration 3

This is the third attempt at designing the dApp connect protocol. It is supposed to improve upon the [second iteration](#iteration-2) by allowing for short sessions without the user needing to enter their password for a long period of time.

#### Iteration 3 Session Establishment

Like with the [second iteration](#iteration-2-session-establishment), establishing a session consists of initializing the session and then confirming it. Unlike the second iteration, this iteration includes the creation of an "approval" during session confirmation. An "approval" is data that represents the user's approval of the wallet's connection to a dApp. The server uses the "approval secret key" to create an encrypted "session token" and gives it to the dApp to use for authentication.

An approval can last for months while a session can last an hour. When the session expires, the dApp can renew the session and get a new token as long as the approval is valid. This mechanism is similar to [OAuth 2.0 grants](https://oauth.net/2/grant-types/).

1. **Initialize session** - Same as [Iteration 2](#iteration-2-session-establishment)
2. **Confirm session**
   1. DApp → Server: Send authenticated request to `/session/confirm`, providing dApp data (dApp name, icon, etc.)
   2. Server: Verify authentication
   3. Server → User: *(Through the UI)* Present dApp data and ask for approval of connection to dApp
   4. User → Server: *(Through the UI)* Approve the connection to the dApp (by entering wallet password) and send the approval
   5. Server: Generate approval secret key and approval data (approval index, expiration, etc.)
   6. Server: Store approval key, approval data, dApp ID and dApp data into an encrypted file protected by the user's wallet password
   7. Server: Generate session key pair (session ID & session secret key) and session data (date created, expiration, etc.)
   8. Server: Use the approval key to create an encrypted "session token" in the [PASETO](https://paseto.io/) format that contains the dApp ID, session ID, session data and approval index
   9. Server: Derive "session shared secret key" using dApp ID and session key or sending authenticated response
   10. Server → DApp: Send authenticated response containing the token, session ID and session data
   11. DApp: Derive the session shared key using dApp secret key and session ID to verify authentication in the server's response
   12. DApp: Store token, session ID, session data and shared key

Result: The dApp can now use the session shared key to send authenticated requests to the server.

#### Iteration 3 Authenticated Request

Certain requests require the dApp to be authenticated.

1. DApp: Use the Hawk authentication scheme to create Hawk header data using session token and shared key
2. DApp → Server: Place the Hawk header data into the request `Authorization` header and send request
3. Server: Use the approval index (a part of the session token) to retrieve the approval key that will be used to decrypt the session token
4. Server: Extract the shared session key from the decrypted token data
5. Server: Verify the authentication of the dApp using the shared key according to the Hawk authentication scheme
6. Server: Attempt to fulfill request
7. Server → DApp: Send authenticated response

Result: The dApp (maybe) gets what it needed from the server.

#### Iteration 3 Session Termination Through DApp

Same as [Iteration 2](#iteration-2-session-termination-through-dapp).

Result: The dApp cannot successfully send authenticated requests to the server without [establishing a new session](#iteration-3-session-establishment).

#### Iteration 3 Session Termination Through Wallet

Same as [Iteration 1](#iteration-1-session-termination-through-wallet). Because sessions are short, the ["ping" requests for checking session validity in Iteration 2](#iteration-2-session-termination-through-wallet) are unnecessary.

Result: The dApp cannot successfully send authenticated requests to the server without [establishing a new session](#iteration-3-session-establishment).

#### Iteration 3 Session Renewal

Before a session is expired, a dApp can "renew" its session to continue being connected to the wallet as long as the session has not expired.

1. DApp → Server: Send authenticated request to `/session/renew`
2. Server: Decrypt session token and extract data
3. Server: Verify authentication
4. Server: Check approval is not expired
5. Server: Create new session key pair and session data
6. Server: Derive new session shared key using dApp ID and new session key
7. Server: Create a new token containing the dApp ID, approval index, new session ID, new session data and new shared key
8. Server → DApp: Send authenticated response containing the new token, session ID and session data
9. DApp: Derive the new session shared key using dApp secret key and new session ID to verify authentication in the server's response
10. DApp: Store new token, session ID, session data and shared key

Result: The expiration of the session has been extended and the dApp can continue sending authenticated requests.

#### Pros and Cons of Iteration 3

- Pro: Allows for short sessions while not bothering the user asking them to approve a dApp connection frequently
- Con: Shorter sessions can lead to a higher chance of the server being overwhelmed ([THREAT-010](../threat-model/01-threats.md#threat-010-wallet-connection-server-overwhelmed-by-too-many-requests-or-requests-that-are-too-large)) when too many dApps are renewing their sessions too frequently. Finding the right session length is crucial.
- Con: When a session expires, the session cannot be renewed and the dApp must establish a new session, starting all over again. This makes the session renewal mechanism essentially useless and defeats the purpose of having this short-session/long-approval scheme. Attempting the [myriad of ways](http://cryto.net/~joepie91/blog/2016/06/19/stop-using-jwt-for-sessions-part-2-why-your-solution-doesnt-work/) of working around this issue leads to more issues or back to one of the same issues as before, which is why it is said that tokens are [not suited to be used for sessions](http://cryto.net/~joepie91/blog/2016/06/13/stop-using-jwt-for-sessions/).
- Con: Vulnerable to a Man-in-the-Middle (MitM) attack ([THREAT-009](../threat-model/01-threats.md#threat-009-interception-of-http-communication-between-dapp-and-wallet-connection-server)) that is [common to Diffie-Hellman (DH) protocols](https://asecuritysite.com/dh/diffie_crack) where a malicious actor generates two key pairs to intercept dApp (client) requests and server responses by doing DH key exchanges with both of them.

### Iteration 2

This is the second attempt at designing the dApp connect protocol. It is supposed to improve upon the [first iteration](#iteration-1) by addressing and fixing its core issues.

#### Iteration 2 Session Establishment

To be able to authenticate itself to the server, a dApp must first obtain a "session shared secret key" by establishing a session with the server. To establish a session, the dApp must "initialize" the session to get a confirmation ID and then "confirm" the session by sending an authenticated request to the server. In confirming the session, the dApp proves it owns the dApp ID it initially provided to the server.

1. **Initialize session**
   1. DApp: Generate dApp key pair (dApp ID & dApp secret key)
   2. DApp → Server: Send request to `/session/init`, providing dApp ID
   3. Server: Validate received dApp ID
   4. Server: Generate confirmation key pair (confirmation ID + confirmation secret key) and confirmation data (date created, expiration, etc.)
   5. Server: Derive the "confirmation shared secret key" using dApp ID and confirmation key
   6. Server: Store confirmation ID, confirmation shared key and confirmation data into an encrypted file protected by the user's wallet password
   7. Server → DApp: Send authenticated response containing confirmation ID and confirmation data
   8. DApp: Derive confirmation shared key using dApp key and confirmation ID to verify authentication in the server's response
2. **Confirm session**
   1. DApp → Server: Send authenticated request to `/session/confirm`, providing dApp data (dApp name, icon, etc.)
   2. Server: Verify authentication and validate dApp data
   3. Server → User: *(Through the UI)* Present dApp data and ask for approval of connection to dApp
   4. User → Server: *(Through the UI)* Approve the connection to the dApp (by entering wallet password) and send the approval
   5. Server: Generate session key pair (session ID & session secret key) and session data (date created, expiration, etc.)
   6. Server: Derive the "session shared secret key" using dApp ID and session key
   7. Server: Store session ID, shared key, session data (date created, expiration, etc.), dApp ID and dApp data into an encrypted file protected by the user's wallet password
   8. Server → DApp: Send authenticated response containing session ID and session data
   9. DApp: Derive the shared key using dApp key and session ID to verify authentication in the server's response
   10. DApp: Store session ID and session data

Result: The dApp can now use the session shared key to send authenticated requests to the server.

#### Iteration 2 Authenticated Request

Same as [Iteration 1](#iteration-1-authenticated-request).

#### Iteration 2 Session Termination Through DApp

The dApp can end the session before the session expires. Unlike the [first iteration](#iteration-1-session-termination-through-dapp), the dApp does not contact the server.

1. DApp: Delete its stored session data (session ID, shared key, etc.)
2. DApp: *No contact with server*
3. Server *(Some time later)*: Delete its stored session data because it expired or the session was manually ended from the wallet.

Result: The dApp cannot successfully send authenticated requests to the server without [establishing a new session](#iteration-2-session-establishment).

#### Iteration 2 Session Termination Through Wallet

The user, through the UI, can command the server to end the session before the session expires. However, the dApp does not immediately know that the session has ended. For the dApp to know when its session has as soon as possible, it should periodically "ping" the server to check if the session is still valid.

1. \[Optional\] Session validity check (Valid session)
   1. DApp → Server: Send authenticated request to `/session/:session_id`
   2. Server: Verify authentication
   3. Server: Check if session still exists
   4. Server → DApp *(session exists)*: Send authenticated response with `200 OK` HTTP code
   5. DApp: Use session shared key to verify authentication in the server's response
2. End session
   1. Wallet UI → Server: End session
   2. Server: Delete session data (date created, expiration, etc.) from session database
   3. Server → Wallet UI: Respond with success
   4. Server: *No contact with DApp*
3. \[Optional\] Session validity check (Invalid session)
   1. DApp → Server: Send authenticated request to `/session/:session_id`
   2. Server: Verify authentication
   3. Server: Check if session still exists
   4. Server → DApp *(session does not exist)*: Send authenticated response with `404 Not Found` HTTP code
   5. DApp: Use session shared key to verify authentication in the server's response
   6. DApp: Delete stored session data (session ID, shared key, etc.)

Result: The dApp cannot successfully send authenticated requests to the server without [establishing a new session](#iteration-2-session-establishment).

#### Iteration 2 Session Renewal

Same as [Iteration 1](#iteration-1-session-renewal).

#### Pros and Cons of Iteration 2

- Pro: The server does not have to be online for a dApp to be able to end the session
- Pro: Mitigates [THREAT-022](../threat-model/01-threats.md#threat-022-multiple-dapps-establishing-connect-sessions-with-the-same-dapp-id). A malicious actor cannot impersonate a dApp simply by using its dApp ID.
- Con: Sessions cannot be very long (more than 2 weeks) because the shared key needs be changed frequently to reduce the damage in the case of the shared key being stolen. With short sessions, an attacker with a stolen key would be limited in how much damage they can cause. However, short sessions can ruin UX because the user has to enter their wallet password more frequently, which increases the likelihood of the user storing their password insecurely ([THREAT-013](../threat-model/01-threats.md#threat-013-wallet-password-is-written-down-and-stored-in-insecure-location)).
- Con: The extra "ping" requests for a dApp to periodically check the session's validity are an extra burden on the server, which increases the chance of the server being overwhelmed ([THREAT-010](../threat-model/01-threats.md#threat-010-wallet-connection-server-overwhelmed-by-too-many-requests-or-requests-that-are-too-large)).
- Con: Vulnerable to a Man-in-the-Middle (MitM) attack ([THREAT-009](../threat-model/01-threats.md#threat-009-interception-of-http-communication-between-dapp-and-wallet-connection-server)) that is [common to Diffie-Hellman (DH) protocols](https://asecuritysite.com/dh/diffie_crack) where a malicious actor generates two key pairs to intercept dApp requests and server responses by doing DH key exchanges with both the dApp and the server.

### Iteration 1

This is the first attempt at designing the dApp connect protocol, so it is the simplest one.

#### Iteration 1 Session Establishment

To be able to authenticate itself to the server, the dApp must obtain a "session shared secret key" by establishing a session with the server.

1. DApp: Generate dApp key pair (dApp ID & dApp secret key)
2. DApp → Server: Send request to `/session/init`, providing dApp ID and dApp data (dApp name, icon, etc.)
3. Server: Validate received dApp ID and information
4. Server → User: *(Through the UI)* Present dApp data and ask for approval of connection to dApp
5. User → Server: *(Through the UI)* Approve the connection to the dApp (by entering wallet password) and send the approval
6. Server: Generate session key pair (session ID & session secret key) and session data (date created, expiration, etc.)
7. Server: Derive the shared key using dApp ID and session key using [ECDH](20250611-use-ecdh-for-establishing-dapp-connect-shared-key.md)
8. Server: Store session ID, shared key, session data, dApp ID and dApp data into an encrypted file protected by the user's wallet password
9. Server → DApp: Respond with session ID and session data
10. DApp: Derive the shared key using dApp key and session ID using ECDH
11. DApp: Store session ID, session data and shared key

Result: The dApp can now use the session shared key to send authenticated requests to the server.

#### Iteration 1 Authenticated Request

Certain requests require the dApp to be authenticated.

1. DApp: Use the Hawk authentication scheme to create Hawk header data using session ID and shared key
2. DApp → Server: Place the Hawk header data into the request `Authorization` header and send request
3. Server: Verify the authentication of the dApp using the shared key according to the Hawk authentication scheme
4. Server: Attempt to fulfill request
5. Server → DApp: Send response

Result: The dApp (maybe) gets what it needed from the server.

#### Iteration 1 Session Termination Through DApp

The dApp can end the session before the session expires.

1. DApp → Server: Send authenticated request to `/session/end`
2. Server: Verify authentication
3. Server: Delete session data from session database
4. Server → DApp: Respond with HTTP success code (maybe `200 OK` or `204 No Content`)
5. DApp: Delete stored session data (session ID, shared key, etc.)

Result: The dApp cannot successfully send authenticated requests to the server without [establishing a new session](#iteration-1-session-establishment).

#### Iteration 1 Session Termination Through Wallet

The user, through the wallet UI, can command the server to end the session before it expires.

1. Wallet UI → Server: Command to end session
2. Server: Delete session data from session database
3. Server → Wallet UI: Respond with success

Result: The dApp cannot successfully send authenticated requests to the server without [establishing a new session](#iteration-1-session-establishment).

#### Iteration 1 Session Renewal

Before the session is expired, a dApp can "renew" its session to continue being connected to the wallet as long as the session has not expired.

1. DApp → Server: Send authenticated request to `/session/renew`
2. Server: Verify authentication
3. Server: Check session is not expired
4. Server: Discard old session data, and create new session key pair and session data
5. Server: Derive the new shared key using the dApp ID and the new session key using [ECDH](20250611-use-ecdh-for-establishing-dapp-connect-shared-key.md)
6. Server: Store new session ID, shared key and session data
7. Server → DApp: Respond with new session ID and session data
8. DApp: Derive the new shared key using dApp key (unchanged) and new session ID using ECDH
9. DApp: Store session ID, session data and shared key

Result: The expiration of the session has been extended and the dApp can continue sending authenticated requests.

#### Pros and Cons of Iteration 1

- Pro: Less complex, which makes it easier to understand
- Con: Vulnerable to [THREAT-022](../threat-model/01-threats.md#threat-022-multiple-dapps-establishing-connect-sessions-with-the-same-dapp-id), where two or more dApps using the same dApp ID breaks the security of the protocol
- Con: DApp cannot properly end session if server is offline
- Con: DApp would still show user as connected to wallet (until session expiration passes) if session was terminated from wallet
- Con: Vulnerable to a Man-in-the-Middle (MitM) attack ([THREAT-009](../threat-model/01-threats.md#threat-009-interception-of-http-communication-between-dapp-and-wallet-connection-server)) that is [common to Diffie-Hellman (DH) protocols](https://asecuritysite.com/dh/diffie_crack) where a malicious actor generates two key pairs to intercept dApp requests and server responses by doing DH key exchanges with both of them.

## Links

- Refined by [Use PASETO for Security Tokens](20250621-use-paseto-for-security-tokens.md)
- Relates to [Vocabulary for DApp Connect](20250621-vocab-for-dapp-connect.md)
- Relates to [Use \"DApp Connect\" Term for Wallet-DApp Connection](20250608-use-dapp-connect-term-for-wallet-dapp-connection.md)
- Relates to [Use ECDH for Establishing session shared Secret Key](20250611-use-ecdh-for-establishing-dapp-connect-shared-key.md)
- Relates to [Use Hawk for DApp Connection Authentication and Authorization](20240821-use-hawk-for-dapp-connection-authentication-and-authorization.md)
- Relates to [Use Local Server to Connect to DApps](20240102-use-local-server-to-connect-to-dapps.md)
- [List of threats for this project](../threat-model/01-threats.md)
- [Hawk API](https://github.com/mozilla/hawk/blob/main/API.md)
- [Diffie-Hellman (Man-in-the-middle) - Asecuritysite.com](https://asecuritysite.com/dh/diffie_crack)
- [Platform-Agnostic SEcurity TOkens (PASETO)](https://paseto.io/)
- [Stop using JWT for sessions - joepie91's Ramblings](http://cryto.net/~joepie91/blog/2016/06/13/stop-using-jwt-for-sessions/)
- [Stop using JWT for sessions, part 2: Why your solution doesn't work - joepie91's Ramblings](http://cryto.net/~joepie91/blog/2016/06/19/stop-using-jwt-for-sessions-part-2-why-your-solution-doesnt-work/)
