CREATE TABLE IF NOT EXISTS rooms(
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(30) NOT NULL,
    password_hash BYTEA,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
    deleted_at   TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS messages(
    id BIGSERIAL PRIMARY KEY,
    nick VARCHAR(30) NOT NULL,
    text TEXT NOT NULL,
    room_id BIGINT NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT current_timestamp
);