-- +goose Up
CREATE TABLE products (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    slug TEXT NOT NULL,
    status TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX products_slug_unique_idx
    ON products (slug)
    WHERE deleted_at IS NULL;

CREATE INDEX products_status_idx
    ON products (status)
    WHERE deleted_at IS NULL;

-- +goose Down
DROP TABLE products;
