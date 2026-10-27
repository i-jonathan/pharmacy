-- Suppliers: one row per supplier name, unique ignoring case and surrounding spaces.
CREATE TABLE supplier (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL CHECK (btrim(name) <> ''),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX supplier_name_unique_idx ON supplier (lower(btrim(name)));

-- Backfill from every name already used on a receipt. Spellings that differ only by case or
-- spaces become one supplier, named by whichever spelling was used most.
INSERT INTO supplier (name)
SELECT DISTINCT ON (lower(spelling)) spelling
FROM (
    SELECT btrim(supplier_name) AS spelling, count(*) AS uses
    FROM receiving_batch
    WHERE btrim(supplier_name) <> ''
    GROUP BY btrim(supplier_name)
) counted
ORDER BY lower(spelling), uses DESC, spelling COLLATE "C";

ALTER TABLE receiving_batch ADD COLUMN supplier_id INTEGER REFERENCES supplier(id);

UPDATE receiving_batch rb SET supplier_id = s.id
FROM supplier s WHERE lower(btrim(rb.supplier_name)) = lower(btrim(s.name));

CREATE INDEX receiving_batch_supplier_idx ON receiving_batch(supplier_id);

-- Purchase orders. supplier_name is kept as written on the document; supplier_id is optional
-- because the supplier can be chosen after the order is drafted.
CREATE TABLE purchase_order (
    id BIGSERIAL PRIMARY KEY,
    supplier_name TEXT NOT NULL DEFAULT '',
    supplier_id INTEGER REFERENCES supplier(id),
    status TEXT NOT NULL DEFAULT 'open'
        CHECK (status IN ('open', 'partial', 'received', 'closed')),
    note TEXT NOT NULL DEFAULT '',
    -- Set when this order took over the outstanding items of another order.
    source_purchase_order_id BIGINT REFERENCES purchase_order(id),
    -- Incremented on every material change; edits must present the version they started from.
    version INTEGER NOT NULL DEFAULT 1,
    created_by_id INTEGER REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX purchase_order_status_created_idx ON purchase_order(status, created_at DESC);
CREATE INDEX purchase_order_supplier_idx ON purchase_order(supplier_id);
-- An order's outstanding items can be taken over by at most one other order.
CREATE UNIQUE INDEX purchase_order_source_unique_idx ON purchase_order(source_purchase_order_id)
    WHERE source_purchase_order_id IS NOT NULL;

-- product_id stays empty for items that are not products yet; it is filled in when the item
-- is registered and received.
CREATE TABLE purchase_order_item (
    id BIGSERIAL PRIMARY KEY,
    purchase_order_id BIGINT NOT NULL REFERENCES purchase_order(id) ON DELETE CASCADE,
    product_id INTEGER REFERENCES product(id),
    item_name TEXT NOT NULL,
    manufacturer TEXT NOT NULL DEFAULT '',
    note TEXT NOT NULL DEFAULT '',
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    received_quantity INTEGER NOT NULL DEFAULT 0 CHECK (received_quantity >= 0),
    CHECK (received_quantity <= quantity)
);

CREATE INDEX purchase_order_item_order_idx ON purchase_order_item(purchase_order_id);
CREATE INDEX purchase_order_item_product_idx ON purchase_order_item(product_id)
    WHERE product_id IS NOT NULL;

-- Which purchase order line a stock batch was received against.
ALTER TABLE product_batch ADD COLUMN purchase_order_item_id BIGINT REFERENCES purchase_order_item(id);
CREATE INDEX product_batch_purchase_order_item_idx ON product_batch(purchase_order_item_id)
    WHERE purchase_order_item_id IS NOT NULL;
