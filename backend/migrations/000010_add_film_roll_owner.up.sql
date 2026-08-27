ALTER TABLE film_rolls
ADD COLUMN user_id BIGINT NOT NULL
    REFERENCES users(id);

CREATE INDEX film_rolls_user_id_idx
    ON film_rolls (user_id);

ALTER TABLE cameras
ADD CONSTRAINT cameras_id_user_id_unique
    UNIQUE (id, user_id);

ALTER TABLE film_rolls
DROP CONSTRAINT film_rolls_camera_id_fkey;

ALTER TABLE film_rolls
ADD CONSTRAINT film_rolls_camera_owner_fkey
    FOREIGN KEY (camera_id, user_id)
    REFERENCES cameras (id, user_id);
