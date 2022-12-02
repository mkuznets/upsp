DROP TABLE IF EXISTS payments CASCADE;
CREATE TABLE IF NOT EXISTS payments
(
    id                text PRIMARY KEY,
    amount            bigint      NOT NULL,
    currency          text        NOT NULL,

    card_number       text        NOT NULL,
    expiry_date       text        NOT NULL,
    card_holder       text        NOT NULL,
    cvv               text        NOT NULL,

    state             text        NOT NULL,

    acquiring_id      text        NOT NULL default '',
    acquiring_state   text        NOT NULL default '',
    acquiring_version text        NOT NULL default '',

    created_at        timestamptz NOT NULL,
    updated_at        timestamptz NOT NULL
);

CREATE INDEX "payments__state" ON payments (state);
