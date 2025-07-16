# Use DuckDB & Parquet Files for Data Storage

- Status: draft
- Deciders: No-Cash-7970
- Date: 2025-07-16
- Tags: backend, security

## Context and Problem Statement

The desktop wallet requires the user's data (e.g. account keys, settings, dApp connect sessions) to be stored somewhere on their device.

## Decision Drivers

- **Encryption**: Almost all of the data files need to be encrypted to ensure privacy ([THREAT-023](../threat-model/01-threats.md#threat-023-exposure-of-sensitive-or-secret-data-within-desktop-wallet-files)) and prevent tampering ([THREAT-012](../threat-model/01-threats.md#threat-012-modifying-security-settings-in-configuration-files)). The algorithm used to encrypt the files must be strong enough to make it very hard to crack offline ([THREAT-011](../threat-model/01-threats.md#threat-011-cracking-encryption-on-wallet-files-offline)).
- **Portability**: The data storage solution must be able to be embedded into the desktop wallet software and store data into files that can be easily moved and backed up
- **Go library/SDK**: The data storage solution must be able to be used in Go, preferably without requiring CGo to be enabled

## Considered Options

- DuckDB with Parquet files
- DuckDB with DuckDB database files
- SQLite
- DuckLake

## Decision Outcome

Chose to use DuckDB with Parquet files. DuckDB is a lightweight database that can easily be embedded using Go. Using DuckDB with Parquet files should provide easy support for encrypting whole data files. The robust support for encrypting a whole Parquet file is the primary reason for using DuckDB with Parquet files instead of the other options.

**Confidence**: Low. DuckDB is not designed for use cases similar to this project's use case. If using DuckDB fails, then SQLite will have to be used.

### Positive Consequences

- More privacy through the encryption of the entire data files
- Data files will be tamper resistant because they are encrypted

### Negative Consequences

- A "driver" for KMD must be created for it to use DuckDB with Parquet files instead of SQLite
- Accessing the data within the files may be significantly slower because they are encrypted
- Must enable CGo. Fortunately, the [Go SQL Driver For DuckDB](https://github.com/marcboeker/go-duckdb) provides instructions on how to install GCC in Windows.

## Pros and Cons of the Options

### DuckDB with Parquet Files

[DuckDB](https://duckdb.org/) is an small in-process database designed for analytics. It can read and write [Parquet](https://duckdb.org/docs/stable/data/parquet/overview) files directly.

- Pro: The name is duck themed 🦆
- Pro: Small and in-process like SQLite, making it very portable
- Pro: An entire Parquet file does not need to be loaded into memory to read it, even when the file is encrypted
- Pro: The Parquet file format is independent from DuckDB, so any Parquet file reader that supports encryption can read the files. The Parquet file encryption mechanism is part of the [specification](https://github.com/apache/parquet-format/blob/master/Encryption.md) and is [supported by DuckDB](https://duckdb.org/docs/stable/data/parquet/encryption) out of the box.
- Pro: [Client in Go (Go SQL Driver For DuckDB)](https://github.com/marcboeker/go-duckdb) is officially supported. Unfortunately, it requires CGo.
- Con: Not as common, so there is less support and documentation available. However, the documentation on the [official website](https://duckdb.org/docs/stable/) is good.
- Con: DuckDB is a columnar OLAP (OnLine Analytical Processing) database that is designed for analytics, not for simple storing and retrieval of single rows in a table. However, its capability to handle data analysis may come in handy for certain features in the future.
- Con: Cannot easily add to, delete from or update an item within a Parquet file. Entire file needs to be read and then written to a new file. This is fine as long as the data files remain small, but it will become a performance bottleneck if one or more data files becomes large.
- Con: A new KMD "driver" needs to be created for KMD to use DuckDB instead of SQLite
- Con: More files to manage. Unlike SQLite where all the wallet data is stored one database file, using the Parquet file format would require more files for the same data. This is because a Parquet file can store only one table.

### DuckDB with DuckDB Database Files

[DuckDB](https://duckdb.org/) is an small in-process database designed for analytics. It has its own native database format.

- Pro: The name is duck themed 🦆
- Pro: Fast for analytics, but analytics is not this project's use case
- Con: DuckDB database files do not support encryption
- Con: Offers no significant advantage over SQLite for this project's use case

### SQLite

[SQLite](https://sqlite.org/index.html) is an in-process database. It is the most popular database and is embedded in a variety of devices and environments.

- Pro: Widely supported with plenty of documentation and tutorials
- Pro: Small and in-process, which makes it very portable
- Pro: Already used in KMD, which makes it easier to expand for other uses (e.g. storing dApp connect session data)
- Pro: More efficient at adding, deleting and updating data
- Pro: Multiple tables can be placed into one `.db` file, which is more convenient for moving or backing up data
- Pro: There is a [Go SQLite Driver](https://pkg.go.dev/modernc.org/sqlite) that does not require CGo
- Con: Encrypting whole file is not easily supported. It requires installing an [extension](https://github.com/utelle/SQLite3MultipleCiphers) that would have to be managed separately.
- Con: Data compression is not easily supported, which can become a problem as more data is stored

### DuckLake

[DuckLake](https://ducklake.select/) is an integrated data lake and catalog format, where a database is used to catalog and track the Parquet files within a data lake (a bunch of data files stored together somewhere).

- Pro: The name is duck themed 🦆
- Pro: Deletes, inserts and updates within Parquet files are handled by DuckLake, unlike with plain DuckDB with Parquet files where changes to a Parquet file require the whole file to be replaced
- Pro: Utilizes DuckDB and Parquet files
- Pro: Supports encryption
- Con: Cannot set key used for encryption (a random key is generated for each Parquet file)
- Con: Requires using "snapshots", which would create more bloat for this project's use case
- Con: Designed to be used for massive amounts of data, which does not fit this project's use case

## Links

- Supersedes [Use SQLite Database to Store Keys](20240103-use-sqlite-database-to-store-keys.md)
- Related to [Use Local Server to Connect to DApps](20240102-use-local-server-to-connect-to-dapps.md)
- [Types of DBMS and Databases: Advantages, Limitations, and Best Applications - altexsoft](https://www.altexsoft.com/blog/databases-database-management-systems/)
- [Deep Dive into DuckDB with CTO Mark Raasveldt](https://www.youtube.com/watch?v=f9QlkXW4H9A)
