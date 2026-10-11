-- +goose Up
CREATE TABLE organizations (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE organization_members (
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role TEXT NOT NULL CHECK (role IN ('owner', 'member')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (organization_id, user_id)
);

CREATE INDEX organization_members_user_id_idx
    ON organization_members (user_id);

ALTER TABLE products
    ADD COLUMN organization_id UUID NOT NULL REFERENCES organizations(id);

DROP INDEX products_slug_unique_idx;
CREATE UNIQUE INDEX products_organization_slug_unique_idx
    ON products (organization_id, slug)
    WHERE deleted_at IS NULL;

-- +goose Down
DROP INDEX products_organization_slug_unique_idx;
CREATE UNIQUE INDEX products_slug_unique_idx
    ON products (slug)
    WHERE deleted_at IS NULL;

ALTER TABLE products DROP COLUMN organization_id;
DROP TABLE organization_members;
DROP TABLE organizations;
