CREATE TABLE IF NOT EXISTS events (
    description text,
    time timestamp,
    expired boolean NOT NULL DEFAULT false,
    PRIMARY KEY (description, time)
);