CREATE TABLE lenses (
    id BIGSERIAL PRIMARY KEY,
    manufacturer TEXT NOT NULL,
    model TEXT NOT NULL,
    focal_length_mm INTEGER NOT NULL,
    max_aperture DOUBLE PRECISION NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);