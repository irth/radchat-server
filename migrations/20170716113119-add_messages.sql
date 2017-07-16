
-- +migrate Up
CREATE TABLE messages (
  id serial primary key not null,
  sender_id INTEGER REFERENCES users not null,
  target_id INTEGER REFERENCES users not null,
  created_at TIMESTAMP,
  content TEXT NOT NULL
);


-- +migrate Down

DROP TABLE messages;
