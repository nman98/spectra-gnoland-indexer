# Database Data Model and Schema

This file will outline how the data is gathered, stored and processed by the indexer.

## Data Flow from the RPC Node

The indexer will connect to the RPC node and will start to gather the data from the node.
It collects and processes data by using batch processing. For live mode the indexer will collect the data up to the
latest block height and then it will process it in batches.

The data is gathered in the following way:

```mermaid
flowchart TD
    RPC[\RPC Node/]

    subgraph TX Pipeline
        FT([Fetch Transactions])
        D[Transactions]
        E[Transaction Messages]
        F[Gno Regular Addresses]
    end

    subgraph Validator Pipeline
        C{Validator Block Signings}
        G[Gno Validator Addresses]
    end

    BH[Block Height]
    I[(TimescaleDB)]

    RPC --> |Gather Block Height| BH
    RPC --> |Gather Validator Block Signings| C
    RPC --> FT
    BH --> |Gather TX Hashes| FT

    FT --> D
    D --> |Process Transactions| E
    E --> |Process Transaction Messages| F
    C --> |Process Validator Block Signings| G

    BH --> I
    D --> I
    E --> I
    F --> I
    C --> I
    G --> I
```

First the indexer gathers block data and validator signing for those blocks. If there are transactions the
tx hashes are gathered from the block height data and are queried from the RPC. At that moment the transaction
data is gathered and processes and all the transaction general data and messages contained in the transaction
are stored in the database. The regular and validator addresses are processed in that way that the addresses are
stored as unique int32 ids and then referenced by the integer value in the transaction tables.

## Core Schema

```mermaid
erDiagram
    blocks {
      BYTEA hash
      BIGINT height
      TIMESTAMPTZ timestamp
      TEXT chain_id
      chain_name chain_name
    }
    tx_hash_id {
      BIGINT tx_id
      BYTEA tx_hash
      TIMESTAMPTZ timestamp
      chain_name chain_name
    }
    transaction_general {
      BIGINT tx_id
      chain_name chain_name
      TIMESTAMPTZ timestamp
      BIGINT block_height
      TEXT[] msg_types
      event[] tx_events
      BYTEA tx_events_compressed
      BOOLEAN compression_on
      BIGINT gas_used
      BIGINT gas_wanted
      NUMERIC fee_amount
      TEXT fee_denom
      BOOLEAN success
      TEXT error_log
    }
    gno_addresses {
      INTEGER id
      TEXT address
      chain_name chain_name
    }
    gno_validators {
      INTEGER id
      TEXT address
      chain_name chain_name
    }
    validator_block_signing {
      BIGINT block_height
      TIMESTAMPTZ timestamp
      INTEGER proposer
      INTEGER[] signed_vals
      chain_name chain_name
    }
    address_tx {
      INTEGER address
      BIGINT tx_id
      chain_name chain_name
      TIMESTAMPTZ timestamp
    }
    blocks ||--o{ transaction_general : "contains"
    blocks ||--o{ validator_block_signing : "has"
    gno_validators ||--o{ validator_block_signing : "signs"
    gno_validators ||--o{ blocks : "proposes"
    tx_hash_id ||--o{ transaction_general : "identifies"
    transaction_general ||--o{ address_tx : "involves"
    gno_addresses ||--o{ address_tx : "participates"
```

## Message Types

```mermaid
erDiagram
    transaction_general {
      BIGINT tx_id
      chain_name chain_name
    }
    tx_hash_id {
      BIGINT tx_id
      BYTEA tx_hash
    }
    gno_addresses {
      INTEGER id
      TEXT address
    }
    bank_msg_send {
      BIGINT tx_id
      TIMESTAMPTZ timestamp
      chain_name chain_name
      INTEGER from_address
      INTEGER to_address
      amount[] amount
      INTEGER[] signers
      SMALLINT message_counter
    }
    vm_msg_call {
      BIGINT tx_id
      TIMESTAMPTZ timestamp
      chain_name chain_name
      INTEGER caller
      TEXT pkg_path
      TEXT func_name
      TEXT args
      amount[] send
      amount[] max_deposit
      INTEGER[] signers
      SMALLINT message_counter
    }
    vm_msg_add_package {
      BIGINT tx_id
      TIMESTAMPTZ timestamp
      chain_name chain_name
      INTEGER creator
      TEXT pkg_path
      TEXT pkg_name
      TEXT[] pkg_file_names
      amount[] send
      amount[] max_deposit
      INTEGER[] signers
      SMALLINT message_counter
    }
    vm_msg_run {
      BIGINT tx_id
      TIMESTAMPTZ timestamp
      chain_name chain_name
      INTEGER caller
      TEXT pkg_path
      TEXT pkg_name
      TEXT[] pkg_file_names
      amount[] send
      amount[] max_deposit
      INTEGER[] signers
      SMALLINT message_counter
    }
    bank_msg_multi_send {
      BIGINT tx_id
      TIMESTAMPTZ timestamp
      chain_name chain_name
      BOOLEAN direction
      INTEGER address_id
      amount[] coins
      INTEGER[] signers
      SMALLINT message_counter
    }
    transaction_general ||--o{ bank_msg_send : "contains"
    transaction_general ||--o{ vm_msg_call : "contains"
    transaction_general ||--o{ vm_msg_add_package : "contains"
    transaction_general ||--o{ vm_msg_run : "contains"
    transaction_general ||--o{ bank_msg_multi_send : "contains"
    tx_hash_id ||--o{ bank_msg_send : "has"
    tx_hash_id ||--o{ vm_msg_call : "has"
    tx_hash_id ||--o{ vm_msg_add_package : "has"
    tx_hash_id ||--o{ vm_msg_run : "has"
    tx_hash_id ||--o{ bank_msg_multi_send : "has"
    gno_addresses ||--o{ bank_msg_send : "from/to"
    gno_addresses ||--o{ vm_msg_call : "caller"
    gno_addresses ||--o{ vm_msg_add_package : "creator"
    gno_addresses ||--o{ vm_msg_run : "caller"
    gno_addresses ||--o{ bank_msg_multi_send : "address"
```

## Custom Types

```mermaid
erDiagram
    amount {
      NUMERIC amount
      TEXT denom
    }
    attribute {
      TEXT key
      TEXT value
    }
    event {
      TEXT at_type
      TEXT type
      attribute[] attributes
      TEXT pkg_path
    }
    transaction_general {
      BIGINT tx_id
      event[] tx_events
      BYTEA tx_events_compressed
      BOOLEAN compression_on
    }
    bank_msg_send {
      BIGINT tx_id
      amount[] amount
    }
    vm_msg_call {
      BIGINT tx_id
      amount[] send
      amount[] max_deposit
    }
    vm_msg_add_package {
      BIGINT tx_id
      amount[] send
      amount[] max_deposit
    }
    vm_msg_run {
      BIGINT tx_id
      amount[] send
      amount[] max_deposit
    }
    attribute ||--o{ event : "attributes"
    event ||--o{ transaction_general : "tx_events"
    amount ||--o{ bank_msg_send : "amount"
    amount ||--o{ vm_msg_call : "send/max_deposit"
    amount ||--o{ vm_msg_add_package : "send/max_deposit"
    amount ||--o{ vm_msg_run : "send/max_deposit"
```

## Schema versioning

`schema_migrations` is a global table (no `chain_name`) that records which migrations have been applied to
the database. Because schema changes affect all chains at once, versioning is per-database, not per-chain.

```mermaid
erDiagram
    schema_migrations {
      INTEGER version
      TEXT description
      TIMESTAMPTZ applied_at
      BOOLEAN success
    }
```

The `version` column is the primary key and should be incremented sequentially for each migration. 
`applied_at` defaults to `NOW()` so it is set automatically on insert. `success` lets you record a failed 
migration attempt without deleting the row. Since the project is still in development it won't probably be
used until a stable version of indexer is published.

## Aggregates

```mermaid
erDiagram
    blocks {
      BYTEA hash
      BIGINT height
      chain_name chain_name
    }
    transaction_general {
      BIGINT tx_id
      chain_name chain_name
    }
    validator_block_signing {
      BIGINT block_height
      chain_name chain_name
    }
    address_tx {
      INTEGER address
      BIGINT tx_id
      chain_name chain_name
    }
    block_counter {
      TIMESTAMPTZ time_bucket
      BIGINT block_count
      chain_name chain_name
    }
    tx_counter {
      TIMESTAMPTZ time_bucket
      BIGINT transaction_count
      chain_name chain_name
    }
    validator_signing_counter {
      TIMESTAMPTZ time_bucket
      INTEGER validator_id
      BIGINT blocks_signed
      chain_name chain_name
    }
    daily_active_accounts {
      TIMESTAMPTZ time_bucket
      BIGINT active_account_count
      chain_name chain_name
    }
    fee_volume {
      TIMESTAMPTZ time_bucket
      TEXT denom
      BIGINT volume
      chain_name chain_name
    }
    block_counter ||--o{ blocks : "counts"
    tx_counter ||--o{ transaction_general : "counts"
    validator_signing_counter ||--o{ validator_block_signing : "counts"
    daily_active_accounts ||--o{ address_tx : "counts"
    fee_volume ||--o{ transaction_general : "sums fees"
```
