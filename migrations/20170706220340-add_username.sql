
-- +migrate Up
ALTER TABLE users ADD COLUMN username VARCHAR UNIQUE;

-- +migrate Down
ALTER TABLE users DROP COLUMN username;