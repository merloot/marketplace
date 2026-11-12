-- +goose Up
CREATE TABLE products (
    id UUID PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    price_amount BIGINT NOT NULL,         
    currency VARCHAR(3) NOT NULL,
    sku VARCHAR(255) UNIQUE NOT NULL,
    category VARCHAR(255),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE outbox (
    id BIGSERIAL PRIMARY KEY,
    aggregate_id UUID NOT NULL,
    event_type VARCHAR(255) NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    processed_at TIMESTAMPTZ
);

CREATE INDEX outbox_unsent_idx ON outbox (id) WHERE processed_at IS NULL;

-- +goose Down
DROP TABLE outbox;
DROP TABLE products;