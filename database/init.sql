-- Create a role 'eda_user' with a password 'eda_password' if it doesn't exist
DO $$ BEGIN
  IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'eda_user') THEN
    CREATE ROLE eda_user WITH LOGIN PASSWORD 'eda_password';
  END IF;
END $$;

--- Drop table events
-- DROP TABLE IF EXISTS events;

-- Create a table named events
CREATE TABLE IF NOT EXISTS events (
    id serial PRIMARY KEY,
    listener_one VARCHAR NOT NULL,
    listener_two VARCHAR NOT NULL,
    listener_three VARCHAR NOT NULL,
    event_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);