ALTER TABLE auth.users ALTER COLUMN created_at TYPE timestamptz USING created_at AT TIME ZONE 'UTC';
ALTER TABLE auth.users ALTER COLUMN updated_at TYPE timestamptz USING updated_at AT TIME ZONE 'UTC';

ALTER TABLE auth.roles ALTER COLUMN created_at TYPE timestamptz USING created_at AT TIME ZONE 'UTC';
ALTER TABLE auth.roles ALTER COLUMN updated_at TYPE timestamptz USING updated_at AT TIME ZONE 'UTC';

ALTER TABLE auth.permissions ALTER COLUMN created_at TYPE timestamptz USING created_at AT TIME ZONE 'UTC';
ALTER TABLE auth.permissions ALTER COLUMN updated_at TYPE timestamptz USING updated_at AT TIME ZONE 'UTC';