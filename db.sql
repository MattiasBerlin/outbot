CREATE TABLE IF NOT EXISTS events (
    description text,
    time timestamp,
    expired boolean NOT NULL DEFAULT false,
    PRIMARY KEY (description, time)
);

CREATE TABLE IF NOT EXISTS participants (
    name text PRIMARY KEY,
    participating boolean NOT NULL
)