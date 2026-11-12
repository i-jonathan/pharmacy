-- A supplier may deliver more than was ordered (bonus stock, pack rounding), and the whole
-- delivery is received against the order. Drop the rule that a line can never receive more than
-- it ordered.
--
-- 000040 created the check unnamed; Postgres named it purchase_order_item_check. A database
-- whose name differs is covered by the second step, which finds the check by what it tests.
ALTER TABLE purchase_order_item DROP CONSTRAINT IF EXISTS purchase_order_item_check;

DO $$
DECLARE
    found_name TEXT;
BEGIN
    FOR found_name IN
        SELECT conname
        FROM pg_constraint
        WHERE conrelid = 'purchase_order_item'::regclass
          AND contype = 'c'
          AND replace(pg_get_constraintdef(oid), ' ', '') LIKE '%received\_quantity<=quantity%'
    LOOP
        EXECUTE format('ALTER TABLE purchase_order_item DROP CONSTRAINT %I', found_name);
    END LOOP;
END
$$;
