# REST API

With this indexer you can use the REST API from the api directory. It is built with chi framework and huma.

Out of the box you get a [Spotlight UI](https://stoplight.io/) to interact with the API on the /docs route.
You can also use some other API UIs but you will need to make the changes yourself. See docs about changing the [UI docs](https://huma.rocks/features/api-docs/).

The huma is framework agnostic and you could modify the API to use some other framework, use another middleware
or maybe use the stdlib http package. This API provides the most basic and necessary features for querying the database.

## API routes

There are total of 5 routes available. This are the basic routes that are needed for the most part. This might be expanded in the future.

### Blocks

- /blocks/{height} - Get a specific block data by height
- /blocks/{from_height}/{to_height} - Get a range of blocks data by height range
- /blocks/{block_height}/signers - Get all of the validators that signed that block + the proposer
- /blocks/latest - Get the latest block data
- /blocks - Get a list of blocks by setting the limit and using cursor.
- /blocks/stats/count/recent - Get the total number of blocks produced in the last 24 hours.
- /blocks/stats/count/daily - Get the block count per day within the given date range. Max range is 30 days.
- /blocks/stats/avg_time - Get Average Block production time.

### Transactions

- /transactions/{tx_hash} - Get a specific basic transaction data by hash, this gives the basic data about the transaction like hash, timestamp, block height, gas used, gas wanted, fee and more.
- /transactions/{tx_hash}/message - Get a specific transaction message data by hash, this gives more detailed data about type of transaction, specific data for that message type and more.
- /transactions - Get a list of transactions by setting the limit and using cursor.
- /transactions/stats/count/recent - Get the total transaction count for the last 24 hours.
- /transactions/stats/count/daily - Get the transaction count per day within the given date range. Max range is 30 days.
- /transactions/stats/count/hourly - Get the transaction count per hour within the given datetime range. Max range is 7 days.
- /transactions/stats/volume/daily - Get the transaction volume grouped by denom per day. Max range is 30 days.
- /transactions/stats/volume/hourly - Get the transaction volume grouped by denom per hour. Max range is 7 days.

### Addresses

- /address/{address}/transactions - Get all of the transactions for a given address for a certain time period
- /addresses/stats/active/daily - Get the number of daily active addresses within the given date range.
- /addresses/stats/total - Get the total number of addresses active since indexer started to process the data.

### Utilities

These endpoints can be queried via POST method.

- /utils/base64url/decode - Convert a base64 encoded tx hash to a base64url encoded tx hash
- /utils/base64url/encode - Convert a base64url encoded tx hash to a base64 encoded tx hash

### Validators

- /validators/{validator_address}/signing/recent - Get the signing performance of a validator over the last 24 hours.
- /validators/{validator_address}/signing/hourly - Get the per-hour signing performance of a validator within the given datetime range. Max range is 7 days.
- /validators/list - Get a list of all validators that were at any point active on the network.
- /validators/signing/24h - Get all validators that signed at least one block in the last 24 hours.

## Setup API

To setup the API you can use the config file. The example config file is in the root under config-api.yml.example.

```yaml
# Example config file for the API
host: 127.0.0.1
port: 8080
cors_allowed_origins:
  - "*"
cors_allowed_methods:
  - "GET"
cors_allowed_headers:
  - "Origin"
  - "Content-Type"
  - "Accept"
cors_max_age: 600
chain_name: gnoland
trusted_proxies:
  - "127.0.0.1/32"
  - "192.168.1.1/32"
  - "10.0.0.1/32"
disable_rate_limit: false
ip_rpm_limit: 30
key_refresh_interval: 5m
```

Some of the environment variables are located under the .env file. The example .env file is in the root under .env.example.

```env
# Example .env file for the API
# do not use password default unless for development or testing!!!
API_DB_HOST=127.0.0.1
API_DB_PORT=5432
API_DB_USER=reader
API_DB_SSLMODE=disable
API_DB_PASSWORD=12345678
API_DB_NAME=gnoland

# these are the default values for the database connection pool
# if they are not filled the API will load the default values
API_DB_POOL_MAX_CONNS=50
API_DB_POOL_MIN_CONNS=10
API_DB_POOL_MAX_CONN_LIFETIME=10s
API_DB_POOL_MAX_CONN_IDLE_TIME=5m
API_DB_POOL_HEALTH_CHECK_PERIOD=1m
API_DB_POOL_MAX_CONN_LIFETIME_JITTER=1m
```

You can make the API by running the following command from the project root:

```bash
make build-api
```

This command will build the API and it will be located in the build directory.

To run the API you can use the following command:

```bash
./build/api -c config-api.yml
```

You can also use the following command to run the API with HTTPS if you have the cert and key files:

```bash
./build/api -c config-api.yml -t cert.pem -k key.pem
```

## Adding API keys

To add API keys you can use the following command:

```bash
./build/api key add <api_key>
```

This will add the API key to the database. You can also use the following command to list all of the API keys:

```bash
./build/api key list
```

You can also use the following command to disable an API key:

```bash
./build/api key disable <api_key>
```

You can also use the following command to enable an API key:

```bash
./build/api key enable <api_key>
```

If you do decide to run the API with the rate limiting enabled you will need to run it with valkey also, which is
present in the docker-compose files.

If you plan to run the APi behind API gateway disable in the config file the usage or rate limits altogether and
in the docker file you can remove the valkey service, if you are running it in docker.
