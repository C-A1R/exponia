DROP INDEX cameras_user_id_idx;

ALTER TABLE cameras
DROP COLUMN user_id;
