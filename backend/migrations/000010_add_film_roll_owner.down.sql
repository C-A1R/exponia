ALTER TABLE film_rolls
DROP CONSTRAINT film_rolls_camera_owner_fkey;

ALTER TABLE film_rolls
ADD CONSTRAINT film_rolls_camera_id_fkey
    FOREIGN KEY (camera_id)
    REFERENCES cameras (id);

ALTER TABLE cameras
DROP CONSTRAINT cameras_id_user_id_unique;

DROP INDEX film_rolls_user_id_idx;

ALTER TABLE film_rolls
DROP COLUMN user_id;
