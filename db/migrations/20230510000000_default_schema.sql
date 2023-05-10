-- +goose Up
ALTER TABLE MODEL ADD COLUMN default_schema INT, ADD FOREIGN KEY (default_schema) REFERENCES SCHEMA (id) ON DELETE SET NULL;

-- +goose Down
ALTER TABLE MODEL DROP COLUMN default_schema;
