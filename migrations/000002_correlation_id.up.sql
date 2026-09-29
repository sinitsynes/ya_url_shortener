ALTER TABLE resource ADD COLUMN correlation_id UUID;
ALTER TABLE resource RENAME COLUMN shortened_url TO short_url;
