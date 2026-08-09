CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    email TEXT NOT NULL,
    display_name TEXT NOT NULL,
    auth_issuer TEXT NOT NULL,
    auth_subject TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT users_email_not_blank
        CHECK (BTRIM(email) <> ''),

    CONSTRAINT users_display_name_not_blank
        CHECK (BTRIM(display_name) <> ''),

    CONSTRAINT users_auth_issuer_not_blank
        CHECK (BTRIM(auth_issuer) <> ''),

    CONSTRAINT users_auth_subject_not_blank
        CHECK (BTRIM(auth_subject) <> ''),

    CONSTRAINT users_auth_identity_unique
        UNIQUE (auth_issuer, auth_subject)
);

CREATE UNIQUE INDEX users_email_unique
    ON users (LOWER(email));
