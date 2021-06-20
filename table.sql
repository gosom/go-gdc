CREATE TABLE individuals (
    registration_number INTEGER PRIMARY KEY,
    data JSONB NOT NULL
);

CREATE INDEX individuals_fts ON individuals
USING gin (( to_tsvector('english', data) ));
