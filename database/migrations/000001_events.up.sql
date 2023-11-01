-- Create a table named events
CREATE TABLE IF NOT EXISTS events (
    id serial PRIMARY KEY,
    listener_one VARCHAR NOT NULL,
    listener_two VARCHAR NOT NULL,
    listener_three VARCHAR NOT NULL,
    event_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);