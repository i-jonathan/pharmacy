package repository

const createUserQuery = `INSERT INTO users (username, password) VALUES ($1, $2)`
const usernameExistsQuery = `SELECT 1 FROM users WHERE username = $1 LIMIT 1`
const fetchUserByNameQuery = `
	SELECT
	    u.id,
	    u.username,
	    u.password,
	    u.role_id,
	    COALESCE(
	        json_agg(
	            json_build_object(
	                'id', p.id,
	                'resource', p.resource,
	                'action', p.action
	            )
	        ) FILTER (WHERE p.id IS NOT NULL),
	        '[]'
	    ) AS permissions
	FROM users u
	JOIN roles r ON r.id = u.role_id
	LEFT JOIN role_permissions rp ON rp.role_id = r.id
	LEFT JOIN permissions p ON p.id = rp.permission_id
	WHERE u.username = $1
	GROUP BY u.id, u.username, u.password, u.role_id;
`
const bulkFetchUserByIDQuery = `SELECT id, username FROM users WHERE id = ANY($1)`
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
          'name', ppo.name,
          'quantity_per_unit', ppo.quantity_per_unit
        )
      ) FILTER (WHERE ppo.id IS NOT NULL), '[]'
    ) AS price_options
	FROM product p
	LEFT JOIN product_price pp ON p.default_price_id = pp.id
	LEFT JOIN product_price ppo ON p.id = ppo.product_id
	WHERE p.name ILIKE $1 || '%'
	OR p.barcode ILIKE $1 || '%'
	OR p.manufacturer ILIKE $1 || '%'
	GROUP BY p.id, pp.id, pp.selling_price, p.name, p.barcode, p.cost_price, p.manufacturer
	ORDER BY
    CASE
        WHEN p.name ILIKE $1 || '%' THEN 1
        WHEN p.barcode ILIKE $1 || '%' THEN 2
        WHEN p.manufacturer ILIKE $1 || '%' THEN 3
        ELSE 4
    END,
    p.name,
    p.barcode,
    p.manufacturer;`
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
const createStockMovementQuery = `
	INSERT INTO stock_movement (product_id, reference_id, movement_type, quantity)
	VALUES ($1, $2, $3, $4)
`
const createMovementFromBatchQuery = `
	INSERT INTO stock_movement (product_id, reference_id, movement_type, quantity)
	VALUES (:product_id, :reference_id, :movement_type, :quantity)
`
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
const createSaleQuery = `
	INSERT INTO sales (receipt_number, status, cashier_id, subtotal, discount, total)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING id;
`
const createSaleItemQuery = `
	INSERT INTO sales_item (sale_id, product_id, quantity, unit_price, discount, total_price)
	VALUES ($1, $2, $3, $4, $5, $6);
`
const createSalePaymentQuery = `
	INSERT INTO sales_payment (sale_id, amount, payment_method)
	VALUES ($1, $2, $3);
`
const bulkCreateSaleItemQuery = `
	INSERT INTO sales_item (sale_id, product_id, quantity, unit_price, discount, total_price)
	VALUES (:sale_id, :product_id, :quantity, :unit_price, :discount, :total_price);
`
const bulkCreateSalePaymentQuery = `
	INSERT INTO sales_payment (sale_id, amount, payment_method)
	VALUES (:sale_id, :amount, :payment_method);
`
const fetchSalesQuery = `
  SELECT id, receipt_number, cashier_id, subtotal, discount, total, created_at
  FROM sales
  WHERE
    ($1::date IS NULL OR created_at::date >= $1::date)
    AND ($2::date IS NULL OR created_at::date <= $2::date)
  ORDER BY created_at DESC
`
const fetchSalesByIDQuery = `
	SELECT id, receipt_number, cashier_id, subtotal, discount, total, created_at
	FROM sales WHERE id = $1;
`
const fetchSaleItemsBySaleIDQuery = `
	SELECT id, product_id, quantity, unit_price, discount, total_price
	FROM sales_item
	WHERE sale_id = $1
	ORDER BY created_at
`
const bulkFetchSaleItemsQuery = `
	SELECT id, sale_id, product_id, quantity, unit_price, discount, total_price
	FROM sales_item WHERE sale_id = ANY($1)
	ORDER BY created_at DESC
`
const bulkFetchSalePaymentsQuery = `
	SELECT sale_id, amount, payment_method
	FROM sales_payment WHERE sale_id = ANY($1)
	ORDER BY payment_method ASC
`
const bulkFetchProductByIDQuery = `
	SELECT id, name, manufacturer
	FROM product WHERE id = ANY($1)
`
const fetchInventoryViewQuery = `
	SELECT * from inventory_view ORDER BY name ASC;
`
const fetchPriceByIDQuery = `SELECT * from product_price where id = $1`
const insertIntoHeldTransactionQuery = `
	INSERT INTO held_transaction (type, reference, payload)
	VALUES ($1, $2, $3);
`
const upsertHeldTransactionQuery = `
	INSERT INTO held_transaction (type, reference, payload)
	VALUES ($1, $2, $3)
	ON CONFLICT (reference)
	DO UPDATE SET
		payload = EXCLUDED.payload,
		updated_at = NOW();
`
const fetchHeldTransactionByTypeQuery = `
	SELECT * from held_transaction
	WHERE type = $1
	ORDER BY updated_at DESC
`
const deleteHeldTransactionByReferenceQuery = `
	DELETE from held_transaction
	WHERE reference = $1;
`
const createReturnQuery = `
	INSERT INTO returns (sale_id, cashier_id, total_refunded, notes)
	VALUES ($1, $2, $3, $4)
	RETURNING id;
`
const bulkCreateReturnItemQuery = `
	INSERT INTO return_items (return_id, sale_item_id, quantity)
	VALUES (:return_id, :sale_item_id, :quantity);
`
const fetchReturnsForSaleBySaleIDQuery = `
	SELECT ri.sale_item_id as sale_item_id, SUM(ri.quantity) as quantity
	FROM return_items ri
	JOIN returns r ON r.id = ri.return_id
	WHERE r.sale_id = $1
	GROUP BY ri.sale_item_id;
`
const bulkFetchReturnsForSaleBySaleIDQuery = `
	SELECT
	    r.sale_id,
	    ri.sale_item_id,
	    SUM(ri.quantity) AS quantity
	FROM return_items ri
	JOIN returns r ON r.id = ri.return_id
	WHERE r.sale_id = ANY($1)
	GROUP BY r.sale_id, ri.sale_item_id;
`
const createStockTakingQuery = `
	INSERT INTO stock_taking (name, status, created_by_id)
	VALUES ($1, $2, $3)
	RETURNING id;
`
const createStockTakingItemQuery = `
	INSERT INTO stock_taking_item
	(stock_taking_id, product_id, snapshot_quantity)
	VALUES ($1, $2, $3)
	RETURNING id;
`
const updateStockTakingItemQuery = `
	UPDATE stock_taking_item
	SET
	    dispensary_count = $1,
	    store_count = $2,
	    notes = $3,
	    last_updated_by_id = $4,
	    last_updated_at = NOW()
	WHERE id = $5
`
const getStockTakingByIDQuery = `
	SELECT
	    st.id,
	    st.name,
	    st.status,
	    st.created_at,
	    st.started_at,
	    st.completed_at,
	    u.id   AS created_by_id,
	    u.username AS created_by_name
	FROM stock_taking st
	JOIN users u ON u.id = st.created_by_id
	WHERE st.id = $1;
`
const getStockTakingItemsQuery = `
	SELECT
	    p.id AS product_id,
	    sti.id AS stock_taking_item_id,
	    p.name AS product_name,
	    p.manufacturer,
	    COALESCE(sti.snapshot_quantity,
	        COALESCE(SUM(
	            CASE
	                WHEN sm.movement_type LIKE 'IN%'  THEN sm.quantity
	                WHEN sm.movement_type LIKE 'OUT%' THEN -sm.quantity
	                ELSE 0
	            END
	        ), 0)
	    ) AS snapshot_quantity,
	    sti.dispensary_count,
	    sti.store_count,
	    MIN(pb.expiry_date) AS earliest_expiry,
	    ARRAY_AGG(DISTINCT pb.expiry_date ORDER BY pb.expiry_date) FILTER (WHERE pb.expiry_date IS NOT NULL) AS expiry_options,
	    sti.notes,
	    u.username AS last_updated_by,
	    sti.last_updated_at AS last_updated_at
	FROM product p
	LEFT JOIN stock_taking_item sti
	    ON sti.product_id = p.id
	   AND sti.stock_taking_id = $1
	LEFT JOIN product_batch pb
	    ON pb.product_id = p.id
	LEFT JOIN stock_movement sm
    	ON sm.product_id = p.id
	LEFT JOIN users u
	    ON u.id = sti.last_updated_by_id
	GROUP BY
	    p.id,
	    sti.id,
	    p.name,
	    p.manufacturer,
	    sti.snapshot_quantity,
	    sti.dispensary_count,
	    sti.store_count,
	    sti.notes,
	    sti.last_updated_at,
	    u.username
	ORDER BY p.name
`
const checkIfActiveStockTaking = `
	SELECT EXISTS (
		SELECT 1 FROM stock_taking WHERE status = $1
	);
`
const currentProductStockQuery = `
	SELECT COALESCE(SUM(
        CASE
            WHEN movement_type LIKE 'IN%'  THEN quantity
            WHEN movement_type LIKE 'OUT%' THEN -quantity
            ELSE 0
        END
    ), 0) AS stock
    FROM stock_movement
    WHERE product_id = $1;
`
const getStockTakingItemByProductIDQuery = `
	SELECT * from stock_taking_item
	WHERE stock_taking_id = $1
	AND product_id = $2;
`
const updateProductCurrentExpiry = `
	UPDATE product
	SET current_expiry = $2::date
	WHERE id = $1;
`
const completeStockTakingQuery = `
	UPDATE stock_taking
	SET status = $2, completed_at = $3, completed_by_id = $4
	WHERE id = $1;
`
const listAllStockTakingsQuery = `
	SELECT
	    st.id,
	    st.name,
	    st.status,
	    st.started_at,
	    st.completed_at,
	    cu.username AS created_by_name,
	    cbu.username AS completed_by_name
	FROM stock_taking st
	JOIN users cu ON cu.id = st.created_by_id
	LEFT JOIN users cbu ON cbu.id = st.completed_by_id
	ORDER BY st.started_at DESC;
`

const updateProductQuery = `
	UPDATE product
	SET name = $1, barcode = $2, category_id = $3, reorder_level = $4, manufacturer = $5, cost_price = $6
	WHERE id = $7
`

const updateProductPriceQuery = `
	UPDATE product_price
	SET quantity_per_unit = $1, selling_price = $2, name = $3
	WHERE id = $4
`

const deleteProductPriceQuery = `
	DELETE FROM product_price WHERE id = $1
`

const fetchProductByIDWithPricesQuery = `
	SELECT
		p.id, p.name, p.barcode, p.cost_price, p.manufacturer, p.category_id, p.reorder_level, p.default_price_id,
		COALESCE(
			(SELECT SUM(
				CASE
					WHEN movement_type LIKE 'IN%'  THEN quantity
					WHEN movement_type LIKE 'OUT%' THEN -quantity
					ELSE 0
				END
			) FROM stock_movement WHERE product_id = p.id), 0
		) AS stock,
		pp.id as "default_price.id", pp.selling_price as "default_price.selling_price",
		pp.name as "default_price.name", pp.quantity_per_unit as "default_price.quantity_per_unit",
		COALESCE(
	      json_agg(
	        json_build_object(
	          'id', ppo.id,
	          'selling_price', ppo.selling_price,
	          'name', ppo.name,
	          'quantity_per_unit', ppo.quantity_per_unit
	        )
	      ) FILTER (WHERE ppo.id IS NOT NULL), '[]'
	    ) AS price_options
	FROM product p
	LEFT JOIN product_price pp ON p.default_price_id = pp.id
	LEFT JOIN product_price ppo ON p.id = ppo.product_id
	WHERE p.id = $1
	GROUP BY p.id, p.name, p.barcode, p.cost_price, p.manufacturer, p.category_id, p.reorder_level, 
	p.default_price_id, pp.id, pp.selling_price, pp.name, pp.quantity_per_unit;
`
