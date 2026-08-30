CREATE TABLE frames (
    id BIGSERIAL PRIMARY KEY,

    film_roll_id BIGINT NOT NULL
        REFERENCES film_rolls(id)
        ON DELETE CASCADE,

    frame_index INTEGER NOT NULL,

    frame_label TEXT,

    note TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT frames_index_check
        CHECK (frame_index > 0),

    CONSTRAINT frames_roll_index_unique
        UNIQUE (film_roll_id, frame_index)
);
