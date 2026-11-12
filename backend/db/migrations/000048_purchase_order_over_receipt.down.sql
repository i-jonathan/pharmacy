-- Fails if any line has received more than it ordered. Correct those lines first: clamping them
-- here would rewrite what actually arrived. Restores the name Postgres gave the check in 000040.
ALTER TABLE purchase_order_item
    ADD CONSTRAINT purchase_order_item_check
    CHECK (received_quantity <= quantity);
