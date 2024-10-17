# Use MemGuard to Prevent Secret Data in Memory Being Swapped to Disk Unencrypted

- Status: accepted
- Deciders: No-Cash-7970
- Date: 2024-10-17
- Tags: backend, kmd, security

## Context and Problem Statement

Some of the [go-algorand](https://github.com/algorand/go-algorand) maintainers have said multiple times that the Key Management Daemon (KMD) is not "production-ready" and is only meant to be used for developing with Algorand. Although the reason for why KMD is not production-ready has not been fully explained, the code comments and the documentation of KMD point to possible reasons why. One possible reason is mentioned in [KMD's documentation](https://github.com/algorand/go-algorand/tree/8b6c443d6884b4c0d3e3b3faf35b886fb81598a3/daemon/kmd#preventing-memory-from-swapping-to-disk). There is a security issue involving how the wallet password and the secret keys are temporarily stored in memory for convenience. The problem with storing secret data like keys in memory is that they can end up on disk unencrypted due to memory swapping/paging.

## Decision Drivers

- **Security:** The solution must reduce or eliminate possibility unencrypted wallet passwords or secret keys getting exposed to the disk while stored in memory.
- **Ease of use:** Prefer a solution that takes care of the minute details of securing data in memory.
- **Cross-platform:** The solution needs to be able to be compiled on Linux, Mac and Windows. Stay away from a solution that uses CGo because it requires GCC, which is difficult to install on Windows.

## Considered Options

- Using `mlockall` command
- [MemGuard](https://pkg.go.dev/github.com/awnumar/memguard) Go package

## Decision Outcome

Chose MemGuard. It should be cross-platform because it is written in pure Go. Using `mlockall` as an option is killed by the fact that it is a Linux-only command.

**Confidence:** Medium. This is a common problem with the kind of software that is often written Go. MemGuard is the best pure Go solution found so far.

UPDATE (2024-10-17): No problems using Memguard so far with KMD. After looking into the code, it appears Memguard does what it says it does.

## Pros and Cons of the Options

### Using `mlockall` command

A [Linux command](https://linux.die.net/man/2/mlockall) that locks the memory for a process from being put into the swap on disk. This is the solution recommended in the [KMD documentation](https://github.com/algorand/go-algorand/tree/8b6c443d6884b4c0d3e3b3faf35b886fb81598a3/daemon/kmd#preventing-memory-from-swapping-to-disk).

- Con: Linux only
- Pro: Common and well-tested solution for preventing memory swapping/paging
- Pro: Good documentation

### MemGuard

A [Go package](https://pkg.go.dev/github.com/awnumar/memguard) for creating a "software enclave." With a software enclave, memory swapping/paging would be less of an issue because the secret data in memory is encrypted. Therefore, the secret data is less likely to end up on disk unencrypted.

- Pro: Pure Go, no CGo required
- Pro: Good documentation
- Con: Somewhat complex to use

## Links

- Relates to [Integrate KMD Wallet Management Code](20240217-integrate-kmd-wallet-management-code.md)
- [MemGuard on GitHub](https://github.com/awnumar/memguard)
- [KMD Code (with documentation)](https://github.com/algorand/go-algorand/tree/8b6c443d6884b4c0d3e3b3faf35b886fb81598a3/daemon/kmd#preventing-memory-from-swapping-to-disk)
