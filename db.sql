CREATE TABLE IF NOT EXISTS events (
    id SERIAL PRIMARY KEY,
    description text,
    time timestamp,
    expired boolean NOT NULL DEFAULT false,
    PRIMARY KEY (description, time)
);

CREATE TYPE participant_instance AS ENUM ('Main', 'Academy');

CREATE TABLE IF NOT EXISTS participants (
    instance participant_instance NOT NULL,
    name text NOT NULL,
    participating boolean NOT NULL,
    preferred_role text NOT NULL DEFAULT 'No preference',
    user_id text NOT NULL DEFAULT '',
    PRIMARY KEY (instance, name)
);