CREATE TABLE exposures (
    id BIGSERIAL PRIMARY KEY,

    frame_id BIGINT NOT NULL
        REFERENCES frames(id)
        ON DELETE CASCADE,

    exposure_index INTEGER NOT NULL,

    camera_id BIGINT NOT NULL
        REFERENCES cameras(id),

    lens_id BIGINT
        REFERENCES lenses(id),

    aperture DOUBLE PRECISION,

    shutter_speed_us BIGINT,

    shot_at TIMESTAMPTZ,

    note TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT exposures_index_check
        CHECK (exposure_index > 0),

    CONSTRAINT exposures_aperture_check
        CHECK (aperture > 0),

    CONSTRAINT exposures_shutter_speed_check
        CHECK (shutter_speed_us > 0),

    CONSTRAINT exposures_frame_index_unique
        UNIQUE (frame_id, exposure_index)
);
