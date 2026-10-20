-- +goose Up
CREATE TABLE Files (
    id            UUID PRIMARY KEY,
    original_name VARCHAR(255) NOT NULL ,
    extension     VARCHAR(255) NULL,
    object_key    VARCHAR(512) NOT NULL UNIQUE ,
    bucket        VARCHAR(255) NOT NULL ,
    content_type  VARCHAR(255) NOT NULL ,
    size_bytes    BIGINT NOT NULL CHECK ( size_bytes > 0 ),
    created_at    TIMESTAMPZ NOT NULL DEFAULT NOW()
)

CREATE INDEX IDX_files_created_at ON Files (created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS Files;