ALTER TABLE rooms
    ADD COLUMN last_active_at TIMESTAMPTZ NOT NULL DEFAULT current_timestamp;