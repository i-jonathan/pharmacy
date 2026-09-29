ALTER TABLE sales_item
    ADD COLUMN stock_quantity_per_unit integer NOT NULL DEFAULT 1
    CHECK (stock_quantity_per_unit > 0);
