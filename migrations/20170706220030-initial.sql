-- +migrate Up
CREATE TABLE users (
  id serial primary key not null,
  display_name VARCHAR not null
);

CREATE TABLE auth_tokens (
  token VARCHAR primary key,
  user_id INTEGER REFERENCES users
);

CREATE TABLE remote_users (
  remote_id VARCHAR primary key,
  user_id INTEGER REFERENCES users
);

CREATE TABLE friendships (
  id serial PRIMARY KEY,
  user_id INTEGER REFERENCES users,
  friend_id INTEGER REFERENCES users
);

-- +migrate Down
DROP TABLE friendships;
DROP TABLE remote_users;
DROP TABLE auth_tokens;
DROP TABLE users;
