<h1 align="center"> The Spectra 💫 Gnoland Indexer</h1>
<div align="center">
<img src="./images/spectra-gnoland-indexer.svg" alt="Spectra Gnoland Indexer" width="250" style="padding: 25px;">
</div>
<p align="center" style="font-style: italic;">This project is still in development. Future releases might have some breaking changes.</p>

The Spectra Gnoland Indexer(SGI) is a tool that records the data from the Gnoland blockchain and stores the data
in the timeseries database, Timescale DB. This indexer will be a part of the Spectra explorer and will be used to
store the data for the explorer. This program can be used for any kind of data/analytics program or app.

The biggest problem when dealing with the blockchain data is how to store the data in a way that is easy to query
and analyze. Some projects rely on the direct access to the data from the blockchain nodes and the problem in this
case is that the nodes are meant for multiple purposes.

They usually provide at least one, sometimes even more, endpoints for querying the data, then this node need to be
able to work with other nodes and to acquire the block data from other nodes via P2P network. Most nodes can be
used as a miners/validators that need to sign the blocks. Even if these nodes are only to be used as a RPC nodes
they can still provide the data but depending on the underlying SDK, tech stack and other factors it might be
not the best choice. And finally the nodes are not the best when it comes to storing the data and scalability.

## Table of content

- [Table of content](#table-of-content)
- [Quick Start](#quick-start)
- [The solution](#the-solution)
- [Why TimescaleDB? And can it work on other SQL databases?](#why-timescaledb-and-can-it-work-on-other-sql-databases)
- [How does the indexer work?](#how-does-the-indexer-work)
- [What data is stored in the database?](#what-data-is-stored-in-the-database)
  - [Blocks](#blocks)
  - [Validator signings](#validator-signings)
  - [Transactions](#transactions)
  - [Counters](#counters)
- [Pros and cons of the SGI](#pros-and-cons-of-the-sgi)
  - [🦾 Pros](#-pros)
  - [🐞 Cons](#-cons)
- [In depth documentation](#in-depth-documentation)

## Quick Start

The fastest way to get the indexer running is with Docker Compose for the database and a local binary for setup and ingestion.

**Prerequisites:** Go 1.26+ and Docker with Docker Compose.

### 1. Clone and build

```bash
git clone https://github.com/Cogwheel-Validator/spectra-gnoland-indexer.git
cd spectra-gnoland-indexer
make build-indexer
```

### 2. Configure

```bash
cp config.yml.example config.yml
```

Open `config.yml` and set the `rpc` field to your Gnoland node RPC endpoint (e.g. `https://gnoland-testnet-rpc.cogwheel.zone`). Make sure `chain_name` matches the name you want to use for the database.

### 3. Start the database

```bash
docker-compose up -d timescaledb
```

This starts a TimescaleDB instance on port 5432 with the default password `12345678`. Change the password in `docker-compose.yml` for any non-local deployment.

### 4. Initialize the database

```bash
./build/indexer setup create-db \
  --db-host localhost --db-port 5432 \
  --db-user postgres --db-name postgres \
  --ssl-mode disable --new-db-name gnoland --chain-name gnoland
```

When prompted, enter the postgres password. This creates the schema and all necessary tables.

Optionally create a dedicated writer user:

```bash
./build/indexer setup create-user writer \
  --db-host localhost --db-port 5432 \
  --db-user postgres --db-name postgres \
  --ssl-mode disable --privilege writer
```

### 5. Run the indexer

**Historic mode** — index a specific block range (useful for initial sync):

```bash
./build/indexer run historic --config config.yml --from-height 1 --to-height 50000
```

**Live mode** — follow the chain tip in real time after historic sync is done:

```bash
./build/indexer run live --config config.yml
```

If you want to start from the current chain tip without syncing history, add `--skip-db-check`:

```bash
./build/indexer run live --config config.yml --skip-db-check
```

For production deployment, systemd service examples, Docker usage, and configuration details see [docs/setup.md](docs/setup.md).

## The solution

The data in the blockchain is mostly tied to blocks, however when any kind of analytics is needed we need some
parameter that falls familiar to us from the daily life. So it comes more natural to compare the data with some
sort of time parameter then to a block height. Then the data needs to be stored in a way that is easy to query and
analyze. And since we are dealing with the time parameter, we need to have the ability to aggregate the data over
some time period so we can timeseries and easily plot the data.

A lot of indexers use NoSQL databases for this use case. However they usually have their own problems and
limitations. They are not as flexible as the SQL databases when it comes to the data types and the query language.
that is the reason why the TimescaleDB is a perfect fit for this use case.

Anyone with some experience with the SQL, especially with the PostgreSQL, can easily understand the TimescaleDB.
It is a extension of the PostgreSQL that adds the ability to store the time series data in a way that is easy to
query and analyze.

## Why TimescaleDB? And can it work on other SQL databases?

TimescaleDB (Tiger Data is their commercial version) is a extension of the PostgreSQL that adds the ability to
store the time series data in a way that is easy to query and analyze. It also has a lot of features that can
extend the capabilities of the PostgreSQL and make it more powerful.

The TimescaleDB sits between the OLTP and the OLAP databases. It is a hybrid database that can be used for both.
It is a good fit for the time series data and it is a good fit for the analytics and the exploration of the data.

The TimescaleDB has also feature of it's own that is not present in the Postgres. The user could extend the indexer
by adding the data aggregation features, automatic jobs scheduling, hyperfunctions and more.

This database extension also has a data compression feature that can be used to reduce the storage space and
segment the data into smaller chunks by using time based intervals making the queries faster.

Technically speaking the indexer can work on Postgres database, however you would need to create the tables and
types manually. This might be added later in the future if there is a demand for it but for now the focus is on the
TimescaleDB.

There are some other SQL databases that could work in theory. For them to work they would need to have:

- The postgres wire protocol
- Equivalent of Numeric type ( some databases might have it as DECIMAL, but not all of them )
- The ability to create custom types

If all of the above are met then the indexer can work on them. It might work on CockroachDB for example but this is
out of the scope of this project. Maybe in the future it might be an interesting idea to support it.

## How does the indexer work?

The indexer has 2 main modes of operation:

- Live mode
- Historic mode

Live mode is the mode that is used to index the data in the real time. It will sync up the database to the latest
block height and will continue to index the data in the real time. It works by using near-real-time batch processing method. It will collect all of the data up to the latest block height and then it will process it in
batches. So the process is not instant but it is very fast and it is able to keep up with the latest block height.

Historic mode is the mode that allows the user to index a certain range of blocks. For example you might not need
the whole chain history or just need a part of it. This mode is useful for the testing, partial indexing of the
chain or gradual indexing of the chain.

During both modes the indexer will use fan out method to send the requests to the RPC node. Then all of the data
is collected and processed and decoded in parallel. All of the addresses are collected and stored in the database
and the indexer stores them in it's own cache. Then any address that is found in the transaction is referenced by
the integer value in the transaction tables. At the end of the processing the data is inserted in batches into the
database.

## What data is stored in the database?

The indexer stores the essential data that is related to the blocks, transactions, messages, and accounts.
It also stores the validator block signings and the validator addresses.
The data schema is denormalized to decrease the space needed and to provide the easy and fast access for the
explorer and any other visualizations and analytics tools.

The data is stored in the following tables:

- blocks
- transactions general data
- transaction messages (each message type has its own table)
- regular and validator addresses (each address type has its own table)
- validator block signings
- validator addresses
- ties between the addresses and the transactions (AddressTx table)
- counters for the blocks, transactions, validator signings, daily active accounts and fee volume

Some of the data is not indexed and it is not planned to be indexed in the future. Such as:

### Blocks

Stored data:

- Basic block hash
- Block height
- Block timestamp
- Block chain ID
- Block proposer address
- Block transactions hashes

Not stored:

- Last commit hash
- App hash
- Data hash
- Validators hash
- Next validators hash
- Consensus hash

### Validator signings

Stored data:

- Validator block signing height
- Validator block signing timestamp
- Validator block signing signed validators
- Proposer address

Not stored:

- Missed validators
- Precommits and all of the hashes and other data, so only a confirmation that the validator signed the block

### Transactions

Almost all of the data regarding to the transaction and the messages are stored with the exception for the
VM message Add Package and Call where in theory one could extract even the body of the smart contract. So this is unnecessary to store for explorers and other analytics tools. This might be added in the future if there is a demand for it.

### Counters

- Blocks counter: Counts the number of blocks for each chain
- Transactions counter: Counts the number of transactions for each chain
- Validator signings counter: Counts the number of validator signings for each validator for each chain
- Daily active accounts counter: Counts the number of daily active accounts for each chain
- Fee volume counter: Counts the volume of fees for each chain and each denom

This are all aggregations that TimescaleDB insert by processing already inserted data. It provides a full
history recorded by each of the counters. This is useful for the analytics and the visualization of the data.

## Pros and cons of the SGI

### 🦾 Pros

- The indexer process the data using goroutines and channels, which can provide a faster processing of the data.
- Fast data processing. [see benchmarks](./docs/benchmarks.md)
- The program has 2 modes that can be used for the indexing of the data. This can be useful for the testing, partial indexing of the chain or gradual indexing of the chain.
- The data is stored ready to be used for any kind of analytics and visualization with any programming language.
- No need to deal with Amino encoding and decoding for the messages as the indexer decodes the messages and stores them in the database.
- It comes with a REST API to get you started quickly.
- It relies on a SQL database, which can provide a easier experience for any user that is familiar with the SQL.
- It uses TimescaleDB, a PostgresSQL extension that can be extended with any other extensions, plus the TimescaleDB has a lot of features that are not present in the Postgres.

### 🐞 Cons

- The indexer has a address cache of all of the addresses that were ever used in the transactions. This gives the indexer ability to swap the addresses with the their integer index in the database. However this introduces complexity. Anyone who plans to use the indexer and plans to make some custom solution on working with the data will need to fully understand the data structure and how to use it. The REST API provides a easy way to interact with the data and to get the data in a readable format.
- The indexer relies on the RPC node for the data. If the RPC node is not available the indexer will not be able to index the data. ( although in the future the indexer might be able to use multiple RPC nodes )
- Technically the indexer has a limit of 2 billion addresses. If at any point the Gnoland grows to that size the indexer would need to be updated to support it. It is not a problem for now but it is something to keep in mind.

## In depth documentation

For more detailed documentation, please refer to the [docs](./docs/README.md) directory.
