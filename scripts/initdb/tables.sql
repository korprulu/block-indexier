CREATE TABLE blocks (
    number BIGINT,
    hash VARCHAR(66),
    parent_hash VARCHAR(66) NOT NULL,
    timestamp BIGINT NOT NULL,
    status VARCHAR(15) NOT NULL,
    is_uncle BOOLEAN DEFAULT FALSE,
    PRIMARY KEY (number, hash)
);

CREATE TABLE transactions (
    hash VARCHAR(66) PRIMARY KEY,
    index SMALLINT NOT NULL,
    block_hash VARCHAR(66) NOT NULL,
    block_number BIGINT NOT NULL,
    from_address VARCHAR(42) NOT NULL,
    to_address VARCHAR(42),
    nonce BIGINT,
    data BYTEA,
    value VARCHAR,
    logs JSONB
);
