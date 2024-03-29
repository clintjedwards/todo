CREATE TABLE IF NOT EXISTS tasks (
    id                 TEXT    NOT NULL,
    title              TEXT    NOT NULL,
    description        TEXT    NOT NULL,
    state              TEXT    NOT NULL,
    created            INTEGER NOT NULL,
    modified           INTEGER NOT NULL,
    parent             TEXT,
    PRIMARY KEY (id)
) STRICT;

CREATE TABLE IF NOT EXISTS scheduled_tasks (
    id                 TEXT  NOT NULL,
    title              TEXT  NOT NULL,
    description        TEXT  NOT NULL,
    expression         TEXT  NOT NULL,
    parent             TEXT,
    PRIMARY KEY (id)
) STRICT;
