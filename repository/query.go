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