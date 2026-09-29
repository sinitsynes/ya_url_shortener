ALTER TABLE resource DROP COLUMN correlation_id;
ALTER TABLE resource RENAME COLUMN short_url TO shortened_url;
