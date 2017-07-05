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