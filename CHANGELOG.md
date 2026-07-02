# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.7.3] - 2026-07-02

### Added

- Feat(api): add cmd and route to make a health check [0ca33b9](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/0ca33b96bc4a9701a914fd7e1f79a2f7c26bc334)

### Fixed

- Fix(pkgs/database): fix miscalculation of validator signing percentage [2b9c98f](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/2b9c98f29e9dc506cc223aaf2f73a306d58411e1)
- Fix(indexer): some VM messages contain unsupported chars [f076d34](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/f076d34d7cbbc1ddfcc3f89250d3ff19bdcb289f)
- Fix(indexer): decoder returns duplicate message types [c088191](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/c088191e4ef3aa0ffd7da33f7198ce8d191b6452)

## [0.7.2] - 2026-06-30

This version does include some minor fixes, some caused by the overall change of the database schema.
There were multiple refactors which should improve the maintainability and readability of the codebase.

### Added

- Feat(indexer): add option to accept log level as a flag in the cli [7c295e2](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/7c295e276a6d10f7f0ca6af5a35041721089debe)
- Feat: add initial support for the ssl to postgres [395d536](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/395d5361d5c52bde384fa31ce0afd05574970690)

### Changes

- Chore: small text corrections [dae2116](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/dae211697cc5f588644a02829d4c4c71060f3073)
- Refac(indexer/orchestrator): historic run awaits the context now [72fad55](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/72fad55adc9c16173f5745fa5c5417e49e2e717e)
- Deps: update chi and gnark-crypto [6088cec](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/6088ceccde2f489828c47fcf6fbb4f3f4474b93f)
- Refac(indexer/schema):adjust aggregate tables similarly to db tables [5c48476](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/5c484764f2138c4c125fb4e3c823f84dcb12c7eb)
- Chore(data_processor): remove dead code [f55e7ce](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/f55e7ce535e2dbb54e4f76291f883cb85d20926f)
- Refac(schema&cli): move setup logic closer to schema [35324dd](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/35324dddf59968cc0e766b35ee0cc6997fce37a9)
- Refac(pkgs/schema): add data type assertion and reflaction testing [21b01eb](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/21b01eb9a72756721fd010cc464ed79c37f90f42)
- Refac(indexer/decoder): add registry based system [d998fd1](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/d998fd12990cc1472e1dd88b8598d6f4bff768ff)
- Refac: use message interface for data grouping [ecafb9e](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/ecafb9e61927e09064d83cc8cd622e05326f051f)
- Refac(pkgs/schema): add generics for inserting data [eca95f9](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/eca95f9d419977a79121c35b7523d4b080b9473e)
- Refac(pkgs/schema): move the copy the row logic closer to the schemas [cd43e3b](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/cd43e3bbe2ca566b2a5a40b4116c35758af16918)
- Refac(pkgs): rename sql_data_types to schema [8874feb](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/8874feb9ee8face4188a19deafbf8bada2feef36)
- Docs: update grammer and add quick start section [46f9d0b](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/46f9d0b79b42dd64da847d008e64ce00e9165144)
- Change(api): add function that will provide a mean block prod value [5531175](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/553117588e78686056b23633a7719862a042c8dc)

### Fixed

- Fix(indexer/dp): return errors if processing message fails [2426e89](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/2426e89195563f95c71240ad25c9fe68b09abd84)
- Fix(indexer): a bug where the database pool closes before the ingestion [076f3f5](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/076f3f5f2417085cfd5266de54e1da5924f72372)
- Fix(indexer/cli): add missing tables for adding privilages to the user [74be438](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/74be438702ec7197b3d7be7c0dabe6e45ac60b76)
- Fix(ci): revert the go version for govulncheck [13ba5fc](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/13ba5fc2e0403c767c3a41e1eeab864e2d111099)
- Fix(api): fix empty volume response [3011c77](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/3011c7747e814545701f5134442a1ff7d3522f70)

### Tests and Code Check

- Audit(coderabbit): add additional error handle for query block prod [814b575](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/814b575a33ee5e9db689e311ccef7b2a140e2546)
- Audit(coderabbit): add buildDSN for database connection [dc810b4](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/dc810b41d6061cd74cada3771c4c787979d5ba0b)
- Audit(sonarqube): adjust code to use variable instead of typed in params [747da47](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/747da477ff6481ebd020d7eafe0af9cefa60eebd)
- Audit(sonarqube): drop complexity for some api handler function [f29123c](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/f29123c1ceddc03462a6ae59817bc08233725458)
- Audit(sonarqube): reduce complexity of copyraw test [62070b5](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/62070b5c06fd98882e2888d93a10ce1ed29921f7)
- Audit(sonarqube): drop complexity of aggregation table test [b7b03c0](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/b7b03c05854ab34a5d2a8afeceafa8b7e47e5f63)
- Test(pkgs/schema): add validation for aggregate tables [bf55820](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/bf55820063fa9ae7b38d05359855859bb7879495)

## [0.7.1] - 2026-06-24

This version should add support to process new Gnoland testnet 13. Added support for Auth message
types, and the bank multi send should be supported.

### Added

- Feat(api): add new message types to the tx message route [e38c9f3](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/e38c9f38fe4b0a7285593196d89c140b29f71125)
- Feat(cli): add to setup db to init auth tables and multi send [1b56f5f](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/1b56f5f3832682d04457ca9b8300a6a0fdc06c5a)
- Feat(decoder): add convert to auth methods [72e2334](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/72e2334b44b2625e70ec3b863faf104c5f6ad3d2)
- Feat(sql_data_types): add schemas for the new auth msg types [0ecf0c7](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/0ecf0c7fed377b75f8532dbae7d39029a4521eb9)
- Feat(decoder): add auth msg types [32fb8df](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/32fb8df4c820a6d9a6c7782720194137aeb7bf87)
- Feat(timescaledb): add auth insert methods [33f4d34](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/33f4d340dc27fa31467625438fcdd266e07348b1)

### Changes

- Refac(timescaledb): move the GetAllBlockSigners to query_block.go [a4f9bf3](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/a4f9bf317c5da9bca5ff1abc4fb9bdebc144ed16)
- Refac(data_processor): add new auth types and partially move decoding [81ff897](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/81ff897728d48b4916c86aee79143ad2e15659b0)
- Refactor(decoder): refactor the decoder to use smaller fn per msg type [c71f39c](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/c71f39cbf4211f40e29193f39d2439506ba2628a)
- Deps: update gno to chain/test13 [f08460d](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/f08460d6a44dcf3cf437f8456f7c1ed6a85b0b65)

### Fixed

- Fix: add missing add address [f681153](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/f6811537e78ba1bcb9d46199a9938968b2cd7556)
- Fix(sql_data_types): fix chain_name to use enum type in database [ca36797](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/ca36797cfd6e92d3b9745ef0abe2f0ebd87551fc)
- Fix: add missing data to the create session [ac2130c](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/ac2130c71e40ad8bd435d49dc58624d06c27778e)
- Fix: dockerfile indexer.go path updated and force the toolcahin to auto [a04c48c](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/a04c48ca497a82565673900f90a6e0343232a1ba)
- Fix(ci): fix the release.yml to use correct path to indexer.go [61b96b3](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/61b96b3df53d342567fa8f46c69b63da31505312)

## [0.7.0] - 2026-06-14

This release brings a versioned REST API, several new routes, real-time statistics, partial support 
for the bank multi send message type, and a large number of query fixes and performance refactors. The
database layer received significant work, including a schema migration table for future releases and
a refactor of the transaction lookups to use `tx_hash_id`. The Go toolchain and dependencies were
updated and the TimescaleDB code was separated out of the database package.

This version will make a freeze on any new feature unless it is related to performance, stability, or
security until the v1 release. All of the development will be focused on bug fixes and improvements of 
the existing functionality.

### Added

- REST API routes are now versioned. All routes are served under a `v1` group using Huma route groups. [70b7330](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/70b733076826d664810a73f25ed6c06e4eb0423b), [5dafaf3](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/5dafaf3136056ab4f6125c7e52f35a6841bfc53a)
- Real time statistics served from an in-memory handler. [1b6d563](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/1b6d56326127f2869c8ddc57f0906741037adcab)
- New route to get all validator signings in the last 24h. [df369ad](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/df369ad221ef3a5f392927a2b82da3d915d9a72f)
- Several new REST API routes added and existing ones fixed. [34360b7](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/34360b70b5fbbc0af5aefc85fc97fac35bcf4a6c)
- Partial support for the bank multi send message type in the decoder, SQL data types and the DataProcessor. Full support will be added once Gno Amino encoder has it implemented. [d8c9e2f](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/d8c9e2fcbe9f7922ea7795cf4374fd958a7f32cd)
- Schema migration table to ease database migrations in future releases. [588500b](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/588500beecda954a251be512233dc3e92ceed0be)
- Option to use a base64 hash directly in the `transactions/{hash}` query. [641e5f7](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/641e5f79485fb0ad3ebc68a57806cf3436afc8b3)
- Sort by date/time on queries that can support it.
- Error log and success/execution status to transaction and address transaction queries.
- User-agent support for the RPC client.
- `SECURITY.md` file outlining the security policy and reporting guidelines. [74cdd48](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/74cdd4841dc0c3e4bc013d144c3894a6a1a6d590)

### Changes

- The TimescaleDB code was separated out of the `database` package into its own package. [afdc20d](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/afdc20dd15543a4115e6ab8d55afeb6bfa07bdde)
- Refactored database queries to use `tx_hash_id` for transaction lookups, and refactored the hypertables that contained the `tx_hash` column. [a4f4f3d](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/a4f4f3de27c7f47fa5c0255f415efebb6b2235b5), [f13d980](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/f13d980cde9808ddcbecb2b921916b03dd278af6)
- Optimized the indexer orchestrator and adjusted the codebase to the recent changes. [8d53baa](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/8d53baa00674fafb9b85bbf995b411242e4af676)
- Implemented rate limiting for specific API routes. [0e39720](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/0e39720d7afa21186cbe47b21f1c2a540051535f)
- Refactored hypertable creation to use a structured `HypertableParams` struct for better configuration management, including `OrderBy` and `SegmentBy` options. [79c4fd8](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/79c4fd8b5b5fde01e11de0a5b9bd93e68c72bab5)
- Moved the CLI commands to the `cli` directory and `indexer.go` to the `cmd` file. [59999fa](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/59999fad13ab83080a80255edfcb7ad51e544811)
- Force the pgx driver to use UTC time and refactored date handling across the API and database queries. [0a1cfcb](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/0a1cfcb6f7abed1d7408719418fd52e59807b8ea)
- Daily query views now include data up to the end of the day, and every daily query returns `0` when there is no data instead of returning nothing.
- Changed `NextCursor` and `PrevCursor` fields to be non-omittable in the JSON output.
- Refactored error handling in the address and block handlers, and removed the sort order parameter from address transaction handling.
- Adjusted the RPC HTTP client to use a better HTTP configuration.
- Refactored the Valkey client implementation and used pure Go implementation.
- Updated the Go toolchain to go1.26.4 and updated other dependencies.

### Fixed

- The indexer querying the new database for the last block height returned an error.
- The validator 24h signing query incorrectly counted all blocks instead of only the blocks with validator signings. [6c8a920](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/6c8a920d1a4b920826093ced073dabb67320f463)
- Fixed duplicate entries being made for the `address_tx` table. [04b1a01](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/04b1a011b310c20568b58f552ecfc9fe09cd0514)
- The rate limiter added a TTL of a couple of thousand years; fixed alongside the rate limiting work. [0e39720](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/0e39720d7afa21186cbe47b21f1c2a540051535f)
- Fixed missing query columns in `GetBlock` on the TimescaleDB query.
- Fixed unique detection for the SQL database initialization and added unique keys to some tables.
- `extractCoins` now inserts a `0` value and the coins are inserted at the index.
- Various fixes to the view and transaction queries, including the daily block production query.
- Reverted `performRequest` to first read the body and then parse the JSON.

## [0.6.0] - 2026-03-14

This release has some new features added and minor improvements.

### Added

- Continuous aggregates for the data. Now the database will automatically aggregate the data by the time bucket and the data is aggregated by the chain name. This feature adds metrics for the blocks, transactions, validator signings, daily active accounts and fee volume.
- Added API keys authentication to the API. It is optional and can be disabled in the configuration.
- Add ratelimit by IP address and by API key.
- Docker service Valkey for rate limiting.
- Add multiple new routes to the API to get the data from the continuous aggregates.

### Changes

- Renamed routes of some API routes to be more descriptive.
- Changed the blocks table, now it doesn't store tx hashes.
- Some slight performance improvements were made in data processing.
- Adjusted the SQL commands to use newer TimescaleDB API commands and functions.

## [0.5.0] - 2026-03-05

This release has some new features added and minor improvements.

The zstd compression has been added. This it is not production ready and still in development, however
initial testing has been done and it seems to work, about 30% less storage is required.

The API has been improved with some new routes and minor improvements. The cursor based pagination has been added
to the API on certain routes. The API now also has POST utilities to convert between Base64 and Base64URL.

The docker image for the API has been added. This allows for easy deployment of the API via docker.
The docker compose file has been adjusted for full deployment via docker(for production and development).

### Added

- Indexer can now compress the events using zstandard compression with the use of `-e or --compress-events` flag. Still in development.
- CLI tool to train the zstandard dictionary from the database.
- POST utilities to convert between Base64 and Base64URL
- Docker image for the API
- Cursor based pagination to the API on certain routes.

### Changes

- Renamed routes of some API routes to be more descriptive.
- Docker compose adjusted for full deployment via docker(for production and development)
- Some minor performance improvements during data processing.
- Updated the go version to 1.25.7 and all of the dependencies to the latest version.

## [0.4.0] - 2025-11-26

Mostly it has some bug fixes.

### Fixed

- The indexer would go into the blocks data and store the signers from the last commit, which is actually all of the block signers from the previous block. So it would insert it like it was meant for that block height. From now on the indexer will fetch data from the /commit method and insert it properly. [de40740](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/de407407e67ae588141b361307fae18215f53a18)
- Gnoland can indeed execute multiple message types in the same transaction. The indexer wasn't able to hold this data properly and would cause an error because in the postgres primary key was attached making each message type unique. Now each message type has a message_counter which is a smallint(int16) which is used as a index for that transaction. [d647901](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/d64790181c626e2ac0137dd9cce08d10ec2b6a7c)

### Changes

- The REST API now returns a map of int16 and message data for the `/transaction/{tx_hash}/message` route. [74a35b30](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/74a35b301070bb1dab5cbadec6fe64b16ad7eb3b)
- The indexer should stop using sync.Map and use sync.Mutex with regular map to store the addresses before handing the operation to the AddressSolver function. [dde82a4](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/dde82a4ecc9bce50963fbff3f4e13a3a20b47f9d)
- The Orchestrator needs to make a fetch to the commit RPC method. This operation is done side by side with fetching the blocks method. A side effect of this is that indexer at that moment will fetch 2 times more request to the RPC, and if the Gnoland node has a limit on the amount of RPC clients that can use it, it might cause the indexer to slow down and throw errors on this requests. In future releases this should be improved. [d647901](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/d64790181c626e2ac0137dd9cce08d10ec2b6a7c)
- Updated the go version to 1.25.4

## [0.3.0] - 2025-11-10

In this release there are some fixes and improvements. The live process should work properly now and the REST API has some new routes. CLI commands are now combined with the ones from the setup cli. Some processes have been improved to use less memory. 

### Added

- The REST API has some new routes. The API can now return last block height, last number(x) of blocks, last x transactions. [789c24b](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/789c24b18881cd33758677f8b878aff7bb42f9dc)

### Changed

- The CLI commands are now combined all together so the cmd/setup.go is removed and the users can now only download the main cli and initiate everything they needs for the indexer to work. [805513b](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/805513b20fdd4f452e8d6f5ad6d56d318e78d5d9)
- Changed the data and query operators to use mutex and store any data they process/collect directly into the type they need to return. There shouldn't be any major perfromance difference but it should allocate less memory. [89e5b6d](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/89e5b6d103710970cbe69c02865bb4b0727649b3)

### Fixed

- When Indexer started to run in live mode without any previous data it should start to process the data from first block height. But instead it tried to query block height 0. [3517e5a](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/3517e5aec32a21d70f43b0595d842841098a4c47)

## [0.2.1] - 2025-10-06

Not really much of a change just added dockerignore file, small changes to the release.yml so it pushes the api also.

## [0.2.0] - 2025-10-04

Added the REST API with some basic routes that will come in handy. Small bug fixes and changes.

### Added

- Rest API with 5 basic routes that will come in handy. [90bef0e](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/90bef0ec5a0bff468d4a1d6771b82706029f4ea9),[f2a39d6](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/f2a39d65c5622e0a5953d1f6113ab5eea1996cad),[d875be2](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/d875be2c42eb5fd71a02b4b29d1496c4b7c3de1e)
- Some basic documentation for the API and modified the existing docs. [b2c20d2](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/b2c20d222e4f52b74b333974a7930df3cddad29e)
- Database queries for the API ( althoguh they can be used for any other app or service if needed ) [d875be2](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/d875be2c42eb5fd71a02b4b29d1496c4b7c3de1e), [f2a39d6](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/f2a39d65c5622e0a5953d1f6113ab5eea1996cad)

### Fixed

- Some table fixes, the msg_types would insert the empty string because when the make slice function was called by the accident it added the empty string instead of just allocating the size of the slice. [d875be2](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/d875be2c42eb5fd71a02b4b29d1496c4b7c3de1e)
- The validator signing has the column name addresses for the signed validators. While this is not a bug it wasn't intended. [90bef0e](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/90bef0ec5a0bff468d4a1d6771b82706029f4ea9)

## [0.1.1] - 2025-10-02

Mostly some fixes. The live should work now there was a bug with the RPC client when making a request to the last block height recorded. The retry worker could sometimes send the data to the closed channel. Now the query operator calls the wg.Done functions directly. It should work now but there might be some other bugs.

### Fixed

- There was a bug with the RPC client when making a request to the last block height recorded. [f44b3f3](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/f44b3f34a23a4244cb6b273332af4139b3c1ed05)
- The live process was not working because the RPC client was not sending the height parameter when making a request to the last block height recorded. [f44b3f3](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/f44b3f34a23a4244cb6b273332af4139b3c1ed05)

### Changed

- Moved the database to the pkgs directory [f44b3f3](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/f44b3f34a23a4244cb6b273332af4139b3c1ed05)

## [0.1.0] - 2025-10-01

This is officially first working version of the indexer. The historic version was successfuly tested on the real 
data. It took about 2m30s on 2vCPU for 10K blocks and about 1m30s on 4vCPU for 10K blocks. 
The live version was not tested on the real data yet mostly because there is active pullic Gnoland testnet.
The live will be tested properly on the testnet 9 when it is released. So the index will probably work but expect
some bugs. Some features are still missing and this is still a work in progress.

### Added

- Added a GitHub Actions workflow to build the indexer and release it as a binary and docker image [5f20967](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/5f20967c429dbc95d959cbb09b3b050afe79477b)
- Added docker file and docker compose [18129be](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/18129beb6a02da6f4b4def55d94c6df9b0ef0b28), [5f20967](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/5f20967c429dbc95d959cbb09b3b050afe79477b)
- Docs are now available [18129be](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/18129beb6a02da6f4b4def55d94c6df9b0ef0b28), [5f20967](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/5f20967c429dbc95d959cbb09b3b050afe79477b)

### Changes

- The cmd setup can now add tables to already existing database [5f20967](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/5f20967c429dbc95d959cbb09b3b050afe79477b)

### Fixed 

- There were some bugs related to poinetrs if the value was nil for block responsers related to validator signing [c7f229a](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/c7f229aedfbb3fb7a0fb05553898f7f2bb43f23b)


## [0.1.0-beta.2] - 2025-09-29

The indexer had some bug fixes and some small improvments. The integration test was technically successful but there seems there is some kind of bug with the indexer. The indexer is not fully tested yet only the historic process has been tested. But not any runs were made on the real data. You can try to run this version on the real data but be advised it is not fully tested and might not work as expected.

### Added

- Makefile has been added. If you feel advanterous you can try to build the indexer with greentea garbage collection. [9fdad03](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/9fdad03ca3b9fe213dced5e1ef68912cc792355a)
- Apperently the previous versions didn't had the method to insert the data for the table address_tx. Now every transaction that was executed can be tied to each address that was involved in the transaction. [9fdad03](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/9fdad03ca3b9fe213dced5e1ef68912cc792355a), [6dd764](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/6dd76464b68809ac8df63f2c66f11678e1083b14)
- The CLI for the database setup now has a new command to create a new user for the database and appoint privileges to the user. It can be a reader(for APIs and some other programs that need SELECT privileges) or a writer(example indexer for historical data). [c39c1f7](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/c39c1f7b5f992da468710a54401f73efa6611881)
- Added a retry mechanism for the query operator. [900ee4f](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/tree/900ee4ff933e1015acc7f9a80de28201075370cf)

### Changed

- Updated the go version to 1.25.1 [b3e02b0](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/b3e02b0dc4f7f1896480c6e2c80ccecf79bbb1be)

## [0.1.0-beta.1] - 2025-09-25

The indexer had some bug fixes and some small improvments. The integration test was technically successful but there seems there is some kind of bug with the indexer. The indexer is not ready for production use.

### Changed

- When the indexer decodes the data using Amino decoder it unloads the data into a map[string]any, then from there it would make 2 conversions, one for the general data struct and the second for the sql data types. The idea was to have seperated logic for the general data struct and sql types. However at this point the indexer already needs to call the copy from method where the data is again being unloaded into some sort of tuple. So the first conversion was removed. [50ca1f2](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/50ca1f2e0d3ee1a3637ca26cdd70e5b48732da8d)
- Updated all of the dependencies to the latest version [5370a5c](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/5370a5c5486be5ef3803f16f968c383598e7f033)

### Fixed

- Fixed the sql related bugs, added some missing types, switched to pgtype.Numeric for the amount type [8aea191](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/tree/8aea1919ad7c3ad16c75a4bd2d1afe934a810dc4), [2b7ed52](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/2b7ed528e23c52c2849d2731cd187e921bf6223e),[ddfdcc1](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/ddfdcc1955784ad510de7f7c847d1a8cf3009e71)
- In some instances the pgx data need to be in the pgtype.Array for instance Txs for the block need to be stored into the pgtype.Array. The indexer now uses a generics function to convert the data into the pgtype.Array [2b7ed52](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/2b7ed528e23c52c2849d2731cd187e921bf6223e)
- The chunk end height was incremented by 1 when the indexer started the historic process. This caused the chunks to overlap and the indexer to throw an error about the duplication. The indexer now correctly sets the chunk end height to the max height [ddfdcc1](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/ddfdcc1955784ad510de7f7c847d1a8cf3009e71)
- Fixed a bug where the data processor would ask the address from the regular address cache instead from the validator address cache [ddfdcc1](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/ddfdcc1955784ad510de7f7c847d1a8cf3009e71)


## [0.1.0-alpha.2] - 2025-09-21

This is a second alpha release although the indexer is not yet ready. 

### Added

- Updated all of the dependencies to the latest version [5fb6b8d](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/5fb6b8dc07bcbacd5a8a66d4eb68a66435f2d695)
- Added the generator functions for the integration test [d61cfa6](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/d61cfa64088ad5654fa2553b7c77c56007451917), [34a46fa](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/34a46fafb40d762fc4ac256fd0605da15e6cba8b)
- Added the synthetic integration test [76e42f6](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/76e42f60b4a828a075322c35d03e8ab52a1721ea)
- Moved some of the code logic to it's own package [9cc12e9](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/9cc12e9961e5c7d2e984209faa5ffda97f75eb06), [76e42f6](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/76e42f60b4a828a075322c35d03e8ab52a1721ea), [9ca2214](https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/commit/9ca221475ac90df0edadc6b1eaf028feb75b79a6)


## [0.1.0-alpha.1] - 2025-09-15

This is the first alpha release of the Spectra Gnoland indexer. Technically most of the indexer components are
done but it is not tested fully so this version is not recommended for production use. 

### Added

- CLI for the Indexer
- Config and env loaders
- RPC client with rate limiting
- PGX pool Postgres Client
- Address cache for regular and validator addresses
- Signal hook for graceful shutdown and emergency shutdown
- Amino decoder for the data from the Gnoland Chain
- Major operator/worker pattern for the indexer have been implemented
- Basic database setup 

### Known Issues

- The indexer is not tested and it is not recommended for production use.
- The setup program only sets the database and ties it to the admin user. This could be bad for security.
- The proto encoding for the events is not tested yet and might not even end in the final release.
- Zstandard compression has been added but it has only been used in some minor test nothing more. For this to work properly a synthetic dataset would need to be created and used to train the dictionary. Alternatively it can be trained on the real data but given that the chain is still in the development stage there is no gurantee it will have enough data to train a good dictionary.
