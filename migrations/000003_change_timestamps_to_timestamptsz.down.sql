ALTER TABLE auth.users ALTER COLUMN created_at TYPE timestamp;
ALTER TABLE auth.users ALTER COLUMN updated_at TYPE timestamp;

ALTER TABLE auth.roles ALTER COLUMN created_at TYPE timestamp;
ALTER TABLE auth.roles ALTER COLUMN updated_at TYPE timestamp;

ALTER TABLE auth.permissions ALTER COLUMN created_at TYPE timestamp;
ALTER TABLE auth.permissions ALTER COLUMN updated_at TYPE timestamp;