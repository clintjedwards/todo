CREATE TABLE IF NOT EXISTS tasks (
    id          TEXT    NOT NULL,
    title       TEXT    NOT NULL,
    description TEXT    NOT NULL,
    status      TEXT    NOT NULL,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    parent      TEXT,
    PRIMARY KEY (id),
    FOREIGN KEY (parent) REFERENCES tasks(id) ON DELETE CASCADE
) STRICT;
