package repository

const createUserQuery = `INSERT INTO users (username, password) VALUES ($1, $2)`
const usernameExistsQuery = `SELECT 1 FROM users WHERE username = $1 LIMIT 1`
const fetchUserByNameQuery = `SELECT id, username, password FROM users WHERE username = $1`
const createProductQuery = `INSERT INTO product
	(name, barcode, category_id, reorder_level, manufacturer, cost_price)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING id`
const createProductPriceQuery = `INSERT INTO product_price
	(product_id, quantity_per_unit, selling_price)
	VALUES ($1, $2, $3)
	RETURNING id`
const updateProductDefaultPrice = `UPDATE product SET default_price_id = $1 WHERE id = $2`
const fetchCategoriesQuery = `SELECT id, name, created_at FROM category ORDER BY name ASC;`
const searchProductsQuery = `SELECT
	p.id, p.name, p.barcode, p.cost_price, p.manufacturer,
	pp.id as "default_price.id", pp.selling_price as "default_price.selling_price"
	FROM product p
	LEFT JOIN product_price pp ON p.default_price_id = pp.id
	WHERE p.name ILIKE '%' || $1 || '%'
	OR p.barcode ILIKE '%' || $1 || '%'`
