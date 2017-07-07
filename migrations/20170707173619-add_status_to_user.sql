
-- +migrate Up
CREATE TYPE status AS ENUM('unavailable', 'available', 'busy', 'away');

ALTER TABLE users ADD COLUMN status status;
UPDATE users SET status='unavailable' WHERE status IS NULL;
ALTER TABLE users ALTER COLUMN status SET NOT NULL;


-- +migrate Down
ALTER TABLE users DROP COLUMN status;

DROP TYPE status;