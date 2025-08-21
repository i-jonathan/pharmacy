package repository

const createUserQuery = `INSERT INTO users (username, password) VALUES ($1, $2)`
const usernameExistsQuery = `SELECT 1 FROM users WHERE username = $1 LIMIT 1`
const fetchUserByNameQuery = `SELECT id, username, password FROM users WHERE username = $1`
const createProductQuery = `INSERT INTO product
	(name, barcode, category_id, reorder_level, manufacturer, cost_price)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING id`
const createProductPriceQuery = `INSERT INTO product_price
	(product_id, quantity_per_unit, selling_price, name)
	VALUES ($1, $2, $3, $4)
	RETURNING id`
const updateProductDefaultPrice = `UPDATE product SET default_price_id = $1 WHERE id = $2`
const fetchCategoriesQuery = `SELECT id, name, created_at FROM category ORDER BY name ASC;`
const searchProductsQuery = `SELECT
	p.id, p.name, p.barcode, p.cost_price, p.manufacturer,
	pp.id as "default_price.id", pp.selling_price as "default_price.selling_price",
	COALESCE(
      json_agg(
        json_build_object(
          'id', ppo.id,
          'selling_price', ppo.selling_price,
          'name', ppo.name
        )
      ) FILTER (WHERE ppo.id IS NOT NULL), '[]'
    ) AS price_options
	FROM product p
	LEFT JOIN product_price pp ON p.default_price_id = pp.id
	LEFT JOIN product_price ppo ON p.id = ppo.product_id
	WHERE p.name ILIKE '%' || $1 || '%'
	OR p.barcode ILIKE '%' || $1 || '%'
	GROUP BY p.id, pp.id, pp.selling_price, p.name, p.barcode, p.cost_price, p.manufacturer`
const searchSupplierQuery = `SELECT
	DISTINCT supplier_name FROM receiving_batch
	WHERE supplier_name ILIKE '%' || $1 || '%'`
const createReceivingBatchQuery = `INSERT INTO receiving_batch
	(supplier_name, received_by_id)
	VALUES ($1, $2)
	RETURNING id`
const createProductBatchQuery = `INSERT INTO product_batch (
		product_id, price_id, receiving_batch_id, quantity, cost_price, expiry_date
	)
	VALUES (
		:product_id, :price_id, :receiving_batch_id, :quantity, :cost_price, :expiry_date
	)
	RETURNING id, product_id, quantity`
const createMovementFromBatchQuery = `INSERT INTO stock_movement (product_id, batch_id, movement_type, quantity)
	VALUES (:product_id, :batch_id, :movement_type, :quantity)`
const fetchDefaultPriceIDQuery = `
	SELECT COALESCE(
        p.default_price_id,
        (
            SELECT pp.id
            FROM product_price pp
            WHERE pp.product_id = p.id
            ORDER BY pp.created_at ASC
            LIMIT 1
        )
    ) AS price_id
	FROM product p
	WHERE p.id = $1;
`
const updateProductPricesQuery = `
	WITH updated_product AS (
	    UPDATE product
	    SET cost_price = :cost_price
	    WHERE id = :product_id
	    RETURNING default_price_id
	)
	UPDATE product_price
	SET selling_price = :selling_price
	WHERE id = (SELECT default_price_id FROM updated_product);
`
