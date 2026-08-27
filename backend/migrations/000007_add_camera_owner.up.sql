ALTER TABLE cameras
ADD COLUMN user_id BIGINT
    REFERENCES users(id);

CREATE INDEX cameras_user_id_idx
    ON cameras (user_id);
