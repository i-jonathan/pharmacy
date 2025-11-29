DO $$
DECLARE
    cashier_role_id INT;
BEGIN
    -- Step 1: Get the cashier role ID
    SELECT id INTO cashier_role_id FROM roles WHERE name = 'Cashier' LIMIT 1;

    -- Step 2: Update all existing users to cashier
    UPDATE users SET role_id = cashier_role_id;

    -- Step 3: Set default to cashier for future users (dynamic SQL)
    EXECUTE format('ALTER TABLE users ALTER COLUMN role_id SET DEFAULT %s;', cashier_role_id);

    -- Step 4: Make role_id NOT NULL (also dynamic SQL)
    EXECUTE 'ALTER TABLE users ALTER COLUMN role_id SET NOT NULL;';
END $$ LANGUAGE plpgsql;
