# Use Encrypted DuckDB Files for Data Storage

- Status: draft
- Deciders: No-Cash-7970
- Date: 2026-02-10
- Tags: backend, security

## Context and Problem Statement

The desktop wallet requires the user's data to be securely stored on the user's device as a file. The file format used to store the user's data needs to support a variety of data types and formats that may change drastically over time.

## Decision Drivers

- **Encryption**: The file format must support encryption of an entire file
- **Ease of use (in development)**: Ideally, the data in the file should be able to be easily changed (with the right encryption key) while reducing the number of files needed to store all the user's data
- **Flexibility**: The file format should allow for the structure of the data to change

## Considered Options

- Parquet
- DuckDB

## Decision Outcome

Chose to use encrypted DuckDB files to store the user's data.

### Positive Consequences

- Features, such as Ledger support, can be easily integrated
- Easier to use more complex data structures to store user data

### Negative Consequences

- More work. A new [KMD](20240217-integrate-kmd-wallet-management-code.md) driver needs to be created and the dApp connect session storage logic needs to be refactored to replace Parquet files with DuckDB files.

## Pros and Cons of the Options

### Parquet

Encrypted Parquet files were used to store wallet keys and dApp connect session data. Every data table was its own file. DuckDB was used to manage Parquet files.

- Pro: Supports encryption of an entire file
- Pro: A more common format that is supported by a number for platforms
- Pro: The wallet currently uses Parquet files. It has been used and well-tested.
- Con: Massive restructuring of the data is needed for new features, like support for [Ledger devices](https://www.ledger.com/)
- Con: More files to manage. Each table is a file and they cannot be combined into one file.

### DuckDB

DuckDB has its own file format. Its support for encryption was [recently added in DuckDB v1.4](https://duckdb.org/2025/11/19/encryption-in-duckdb).

- Pro: Supports encryption of an entire file
- Pro: Like a SQLite file, a single DuckDB file can contain multiple tables, which means fewer files to be managed
- Con: The DuckDB file format is not a common format and almost exclusive to DuckDB

## Links

- Supersedes [20250716-use-duckdb-and-parquet-files-for-data-storage](20250716-use-duckdb-and-parquet-files-for-data-storage.md)
- [Data-at-Rest Encryption in DuckDB - DuckDB Blog](https://duckdb.org/2025/11/19/encryption-in-duckdb)
- [Reading and Writing Parquet Files - DuckDB Docs](https://duckdb.org/docs/stable/data/parquet/overview)
