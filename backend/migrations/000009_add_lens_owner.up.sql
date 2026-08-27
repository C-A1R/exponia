ALTER TABLE lenses
ADD COLUMN user_id BIGINT NOT NULL
    REFERENCES users(id);

CREATE INDEX lenses_user_id_idx
    ON lenses (user_id);
