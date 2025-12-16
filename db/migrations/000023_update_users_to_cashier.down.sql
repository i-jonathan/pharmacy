DO $$
BEGIN
    -- Step 1: Remove NOT NULL constraint
    ALTER TABLE users ALTER COLUMN role_id DROP NOT NULL;

    -- Step 2: Remove the default
    ALTER TABLE users ALTER COLUMN role_id DROP DEFAULT;

    UPDATE users SET role_id = NULL;
END $$ LANGUAGE plpgsql;
