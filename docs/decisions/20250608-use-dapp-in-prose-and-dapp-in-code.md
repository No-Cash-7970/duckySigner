# Use "DApp" in Prose and "Dapp" in Code

- Status: accepted
- Deciders: No-Cash-7970
- Date: 2025-06-09
- Tags: dev-process, docs

## Context and Problem Statement

The term "decentralized application" is commonly abbreviated using into "dApp" (with uppercase A) or "dapp" (with lowercase A). The inconsistent capitalization of "dApp" within this project's codebase is irritating.

## Decision Drivers

- **Existing conventions:** How is the term capitalized elsewhere? Being consistent with how the term is commonly written helps make both documentation and code easier to read.
- **Readability in code:** Whichever form is used, it needs to be easily read when stuck into variable names

## Considered Options

- dApp/DApp (with uppercase A)
- dapp/Dapp (with lowercase A)

## Decision Outcome

Chose to use "dApp" (with uppercase A) in documentation or other types of prose, and "dapp" (with lowercase A) in code or filenames. The first letter in with both forms will be capitalized according to English capitalization rules or code style like any other English word.

**Confidence:** Very high

## Pros and Cons of the Options

### DApp (With Uppercase A)

Wikipedia always uses "DApp" (with uppercase D and uppercase A) and never "dApp" (with lowercase D and uppercase A). On the other hand, most publications, like Investopedia, use the form with lowercase D and uppercase A unless the word is at the beginning of a sentence.

- Pro: More obvious that it is an abbreviation for something
- Pro: More common in formal writing (e.g. research papers, news articles)
- Con: Not consistent with standard English rules of capitalization
- Con: Can cause some awkward-looking variable names in code

### Dapp (With Lowercase A)

Publications from the Ethereum Foundation use the all-lowercase form. This form capitalized like a typical English word.

- Pro: Consistent with the way English words are typically capitalized
- Pro: Used by the organization that coined the term (Ethereum Foundation)
- Pro: May become the most common form in the future
- Con: Not typically used in formal writing (e.g. research papers, news articles)
- Con: The way it looks obscures that it is an abbreviation

## Links

- [Wikipedia page for "DApp"](https://en.wikipedia.org/wiki/Decentralized_application)
- [Ethereum page about "dapps"](https://ethereum.org/en/dapps/)
- [Investopedia page about "dApps"](https://www.investopedia.com/terms/d/decentralized-applications-dapps.asp)
