# interview-homework-b

## How to start services at local

```bash
~ docker compose --env-file .env up -d 
```

## Services

- block processor: consume block numbers from Redis stream and retrieve the data from JSON-RPC API
- tx processor: consume transactions from Redis stream and get log data from JSON-RPC API then store them to the database
- scanner: scan the block from the given number n and continuously scan for newly generated blocks
- validator:  check if the block has become an uncle block
- API server

## Configurations

Configurations are saved in a dotenv file in the root directory.

```
# postgres
POSTGRES_HOST=postgres
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=homework

# redis
REDIS_ADDRESS=redis:6379
REDIS_DB=0

# redis streams
BLOCK_STREAM_NAME=blocks
TRANSACTION_STREAM_NAME=transactions

# ethereum
ETHEREUM_RPC_URL=https://eth.llamarpc.com
# How many newly generated blocks to wait, this value will effect the validator
# when to check the block has become uncle block. If its value set to 50, the
# validator will wait until the new block exceeds 50, and then check the 51st
# 52nd, ...nth block.
BLOCK_REORG_CHECK_COUNT=50

# block processors
# The consumer group name
BLOCK_PROCESSOR_CONSUMER_GROUP=block-processors

# The concurrent worker count in an instance 
BLOCK_PROCESSOR_CONCURRENT_COUNT=2

# transaction processors
# The consumer group name
TRANSACTION_PROCESSOR_CONSUMER_GROUP=transaction-processors

# The concurrent worker count in an instance
TRANSACTION_PROCESSOR_CONCURRENT_COUNT=2

# The batch size of the number of transaction records processed at a time
TRANSACTION_PROCESSOR_BATCH_TRANSACTION_COUNT=100

# Scanner service
# Start scanning from the nth block
SCANNER_START_BLOCK_NUMBER=17310465

# The interval time for checking newly generated block
SCANNER_WATCH_INTERVAL_SECONDS=60

# Validator service
# The interval time for checking unfinalized blocks
VALIDATOR_WATCH_INTERVAL_SECONDS=60

# API port
API_PORT=8080
```
