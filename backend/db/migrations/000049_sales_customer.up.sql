-- The customer a sale was made to, when the cashier attached one at the till. NULL is a walk-in,
-- and every sale made before this migration. A customer with sales is never deleted (customers
-- are only deactivated), so the plain foreign key is enough.
ALTER TABLE sales ADD COLUMN customer_id BIGINT REFERENCES customer(id);

-- A customer's own sales (history, credit, loyalty) look sales up by customer.
CREATE INDEX sales_customer_idx ON sales (customer_id) WHERE customer_id IS NOT NULL;
