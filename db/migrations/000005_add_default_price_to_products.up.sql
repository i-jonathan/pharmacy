ALTER TABLE product
ADD COLUMN default_price_id integer not null;

ALTER TABLE product
ADD CONSTRAINT fk_default_price_id 
FOREIGN KEY (default_price_id) REFERENCES product_price(id);