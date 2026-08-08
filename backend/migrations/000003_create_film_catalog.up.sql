CREATE TABLE film_color_types (
    id BIGSERIAL PRIMARY KEY,
    code TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL
);

CREATE TABLE film_formats (
    id BIGSERIAL PRIMARY KEY,
    code TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL
);

CREATE TABLE film_stocks (
    id BIGSERIAL PRIMARY KEY,

    manufacturer TEXT NOT NULL,
    name TEXT NOT NULL,
    iso INTEGER NOT NULL,

    color_type_id BIGINT NOT NULL
        REFERENCES film_color_types(id),

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT film_stocks_iso_check
        CHECK (iso > 0),

    CONSTRAINT film_stocks_unique_name
        UNIQUE (manufacturer, name)
);

CREATE TABLE film_stock_formats (
    film_stock_id BIGINT NOT NULL
        REFERENCES film_stocks(id)
        ON DELETE CASCADE,

    format_id BIGINT NOT NULL
        REFERENCES film_formats(id)
        ON DELETE CASCADE,

    PRIMARY KEY (film_stock_id, format_id)
);