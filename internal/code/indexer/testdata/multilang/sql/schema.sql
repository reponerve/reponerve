-- schema for handler table
CREATE TABLE IF NOT EXISTS handlers (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);

CREATE VIEW handler_names AS
SELECT name FROM handlers;
