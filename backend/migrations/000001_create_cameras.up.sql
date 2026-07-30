CREATE TABLE cameras (
    id BIGSERIAL PRIMARY KEY,
    manufacturer TEXT NOT NULL,
    model TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);