CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- -----------------------------------------------------------------------------

DROP TABLE IF EXISTS merchants CASCADE;
CREATE TABLE IF NOT EXISTS merchants
(
    id         text PRIMARY KEY,
    email      text        NOT NULL,
    created_at timestamptz NOT NULL
);


DROP TABLE IF EXISTS api_keys CASCADE;
CREATE TABLE IF NOT EXISTS api_keys
(
    id              text PRIMARY KEY,
    merchant_id     text REFERENCES merchants (id) NOT NULL,
    secret_key_hash text                           NOT NULL,
    active          boolean                        NOT NULL,
    created_at      timestamptz                    NOT NULL
);
CREATE INDEX "api_keys__secret_key_hash" ON api_keys USING HASH (secret_key_hash) WHERE active = 't';



DROP TABLE IF EXISTS payments CASCADE;
CREATE TABLE IF NOT EXISTS payments
(
    id          text PRIMARY KEY,
    merchant_id text        NOT NULL,
    amount      bigint      NOT NULL,
    currency    text        NOT NULL,

    card_number text        NOT NULL,
    expiry_date text        NOT NULL,
    card_holder text        NOT NULL,
    cvv         text        NOT NULL,

    state       text        NOT NULL,
    "version"   text        NOT NULL,

    created_at  timestamptz NOT NULL,
    updated_at  timestamptz NOT NULL
);

CREATE INDEX "payments__id__version" ON payments (id, version);



DROP TABLE IF EXISTS payment_updates CASCADE;
CREATE TABLE IF NOT EXISTS payment_updates
(
    id         text PRIMARY KEY,
    payment_id text    NOT NULL,
    state      text    NOT NULL,
    "version"  text    NOT NULL,
    is_applied boolean NOT NULL
);

CREATE INDEX "payment_updates__payment_id__version" ON payment_updates (payment_id, version) WHERE is_applied = 'f';
