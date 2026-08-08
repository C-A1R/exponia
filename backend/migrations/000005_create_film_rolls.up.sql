CREATE TABLE film_rolls (
    id BIGSERIAL PRIMARY KEY,

    film_stock_id BIGINT NOT NULL
        REFERENCES film_stocks(id),

    format_id BIGINT NOT NULL
        REFERENCES film_formats(id),

    camera_id BIGINT
        REFERENCES cameras(id),

    exposure_iso INTEGER NOT NULL,

    status TEXT NOT NULL DEFAULT 'unused',

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT film_rolls_exposure_iso_check
        CHECK (exposure_iso > 0),

    CONSTRAINT film_rolls_status_check
        CHECK (
            status IN (
                'unused',
                'ready',
                'in_use',
                'exposed',
                'developed'
            )
        )
);