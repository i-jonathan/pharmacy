package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"pharmacy/internal/types"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// ErrAlreadyFlagged is returned when someone flags a product they already have an open flag on.
var ErrAlreadyFlagged = errors.New("product already flagged by this user")

// FOR NO KEY UPDATE queues adjustments of a product behind each other without blocking the
// foreign-key checks (FOR KEY SHARE) that sales, receipts and flags take on the product row.
const lockProductsQuery = `
	SELECT id FROM product WHERE id = ANY($1) ORDER BY id FOR NO KEY UPDATE`

var productStockQuery = `
	SELECT p.id, p.name, COALESCE(p.cost_price, 0) AS cost_price, p.current_expiry,
		` + fmt.Sprintf(productStockSQL, "p.id") + ` AS stock
	FROM product p
	WHERE p.id = ANY($1)
	ORDER BY p.id
`

// LockProductStockTx locks the products' rows (in id order, so concurrent adjustments queue
// instead of deadlocking) and then returns their stock. The stock is read in a second statement:
// under READ COMMITTED it then sees everything committed while this one waited for the lock,
// which a sum computed in the locking statement would not.
func (r *repo) LockProductStockTx(
	ctx context.Context, tx *sqlx.Tx, productIDs []int,
) ([]types.ProductStock, error) {
	if _, err := tx.ExecContext(ctx, lockProductsQuery, pq.Array(productIDs)); err != nil {
		return nil, err
	}
	var rows []types.ProductStock
	err := tx.SelectContext(ctx, &rows, productStockQuery, pq.Array(productIDs))
	return rows, err
}

// ProductsInActiveStockTakingTx returns the products that already have a count (and so a stock
// snapshot) in an in-progress stock taking.
func (r *repo) ProductsInActiveStockTakingTx(
	ctx context.Context, tx *sqlx.Tx, productIDs []int,
) ([]int, error) {
	var ids []int
	err := tx.SelectContext(ctx, &ids, `
		SELECT DISTINCT sti.product_id
		FROM stock_taking_item sti
		JOIN stock_taking st ON st.id = sti.stock_taking_id
		WHERE st.status = 'in_progress' AND sti.product_id = ANY($1)
		ORDER BY sti.product_id`, pq.Array(productIDs))
	return ids, err
}

// StockAdjustmentIDByKeyTx returns the id of the user's adjustment with that retry key, or 0.
func (r *repo) StockAdjustmentIDByKeyTx(
	ctx context.Context, tx *sqlx.Tx, userID int, key string,
) (int, error) {
	var id int
	err := tx.GetContext(ctx, &id, stockAdjustmentByKeyQuery, userID, key)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	return id, err
}

const createStockAdjustmentQuery = `
	INSERT INTO stock_adjustment (reason, note, created_by_id, idempotency_key)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (created_by_id, idempotency_key) WHERE idempotency_key IS NOT NULL DO NOTHING
	RETURNING id
`

const stockAdjustmentByKeyQuery = `
	SELECT id FROM stock_adjustment WHERE created_by_id = $1 AND idempotency_key = $2`

// CreateStockAdjustmentTx inserts the adjustment and returns its id. When the idempotency key
// was already used it returns the existing adjustment's id and false.
func (r *repo) CreateStockAdjustmentTx(
	ctx context.Context, tx *sqlx.Tx, adj types.NewStockAdjustment,
) (int, bool, error) {
	var id int
	err := tx.GetContext(
		ctx, &id, createStockAdjustmentQuery,
		adj.Reason, adj.Note, adj.CreatedByID, adj.IdempotencyKey,
	)
	if errors.Is(err, sql.ErrNoRows) && adj.IdempotencyKey != nil {
		err = tx.GetContext(
			ctx, &id, stockAdjustmentByKeyQuery, adj.CreatedByID, *adj.IdempotencyKey,
		)
		return id, false, err
	}
	return id, err == nil, err
}

const insertStockAdjustmentItemQuery = `
	INSERT INTO stock_adjustment_item (
		stock_adjustment_id, product_id, quantity_before, quantity_change, unit_cost_kobo,
		expiry_before, expiry_after
	) VALUES ($1, $2, $3, $4, $5, $6, $7)
`

func (r *repo) InsertStockAdjustmentItemTx(
	ctx context.Context, tx *sqlx.Tx, adjustmentID int, item types.NewStockAdjustmentItem,
) error {
	_, err := tx.ExecContext(
		ctx, insertStockAdjustmentItemQuery, adjustmentID, item.ProductID, item.QuantityBefore,
		item.QuantityChange, item.UnitCostKobo, item.ExpiryBefore, item.ExpiryAfter,
	)
	return err
}

func (r *repo) SetProductCurrentExpiryTx(
	ctx context.Context, tx *sqlx.Tx, productID int, expiry time.Time,
) error {
	_, err := tx.ExecContext(ctx, updateProductCurrentExpiry, productID, expiry)
	return err
}

const listStockAdjustmentsQuery = `
	SELECT sa.id, sa.reason, sa.note, COALESCE(u.username, '') AS created_by, sa.created_at
	FROM stock_adjustment sa
	LEFT JOIN users u ON u.id = sa.created_by_id
	ORDER BY sa.created_at DESC, sa.id DESC
	LIMIT $1 OFFSET $2
`

const stockAdjustmentItemsQuery = `
	SELECT i.stock_adjustment_id, i.product_id, p.name AS product_name,
		COALESCE(p.manufacturer, '') AS manufacturer, i.quantity_before, i.quantity_change,
		i.expiry_before, i.expiry_after
	FROM stock_adjustment_item i
	JOIN product p ON p.id = i.product_id
	WHERE i.stock_adjustment_id = ANY($1)
	ORDER BY p.name
`

// ListStockAdjustments returns one page of adjustments, newest first, with their items, and
// the total number of adjustments.
func (r *repo) ListStockAdjustments(
	ctx context.Context, limit, offset int,
) ([]types.StockAdjustment, int, error) {
	adjustments := []types.StockAdjustment{}
	var total int
	err := r.readSnapshot(ctx, func(tx *sqlx.Tx) error {
		if err := tx.GetContext(ctx, &total, `SELECT count(*) FROM stock_adjustment`); err != nil {
			return err
		}
		if err := tx.SelectContext(
			ctx, &adjustments, listStockAdjustmentsQuery, limit, offset,
		); err != nil {
			return err
		}
		ids := make([]int, len(adjustments))
		index := make(map[int]int, len(adjustments))
		for i, a := range adjustments {
			ids[i], index[a.ID] = a.ID, i
			adjustments[i].Items = []types.StockAdjustmentItem{}
		}
		var items []struct {
			AdjustmentID int `db:"stock_adjustment_id"`
			types.StockAdjustmentItem
		}
		if err := tx.SelectContext(
			ctx,
			&items,
			stockAdjustmentItemsQuery,
			pq.Array(ids),
		); err != nil {
			return err
		}
		for _, item := range items {
			i := index[item.AdjustmentID]
			adjustments[i].Items = append(adjustments[i].Items, item.StockAdjustmentItem)
		}
		return nil
	})
	return adjustments, total, err
}

// productStockSQL is the stock of the product whose id is in the expression %s.
const productStockSQL = `COALESCE((
	SELECT SUM(CASE
		WHEN sm.movement_type LIKE 'IN%%' THEN sm.quantity
		WHEN sm.movement_type LIKE 'OUT%%' THEN -sm.quantity
		ELSE 0
	END)
	FROM stock_movement sm WHERE sm.product_id = %s
), 0)`

var createStockFlagQuery = `
	INSERT INTO stock_flag (product_id, note, counted_quantity, flagged_by_id, stock_at_flag)
	VALUES ($1, $2, $3, $4, ` + fmt.Sprintf(productStockSQL, "$1") + `)
	RETURNING id
`

func (r *repo) CreateStockFlagTx(
	ctx context.Context, tx *sqlx.Tx, flag types.StockFlagRequest, userID int,
) (int, error) {
	var id int
	err := tx.GetContext(
		ctx, &id, createStockFlagQuery, flag.ProductID, flag.Note, flag.CountedQuantity, userID,
	)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return 0, ErrAlreadyFlagged
	}
	return id, mapForeignKeyError(err)
}

// stockFlagColumns sums stock only for the flagged products, not through inventory_view, which
// would aggregate every product's movements. moved_since counts only shelf movements (sales,
// receipts, returns) after the flag; a stock taking or adjustment since then changed the system's
// number, not the shelf, so it must not move the cashier's count.
var stockFlagColumns = `
	SELECT f.id, f.product_id, p.name AS product_name,
		COALESCE(p.manufacturer, '') AS manufacturer, f.note, f.counted_quantity,
		f.stock_at_flag, ` + fmt.Sprintf(productStockSQL, "f.product_id") + ` AS current_stock,
		COALESCE((
			SELECT SUM(CASE WHEN sm.movement_type LIKE 'IN%' THEN sm.quantity ELSE -sm.quantity END)
			FROM stock_movement sm
			WHERE sm.product_id = f.product_id AND sm.created_at > f.created_at
				AND sm.movement_type NOT LIKE '%STOCK_TAKING'
				AND sm.movement_type NOT LIKE '%MANUAL_ADJUSTMENT'
		), 0) AS moved_since,
		COALESCE(u.username, '') AS flagged_by, f.created_at, f.status
	FROM stock_flag f
	JOIN product p ON p.id = f.product_id
	LEFT JOIN users u ON u.id = f.flagged_by_id
`

func (r *repo) ListOpenStockFlags(ctx context.Context) ([]types.StockFlag, error) {
	flags := []types.StockFlag{}
	err := r.Data.SelectContext(
		ctx, &flags, stockFlagColumns+` WHERE f.status = 'open' ORDER BY f.created_at DESC`,
	)
	return flags, err
}

func (r *repo) CountOpenStockFlags(ctx context.Context) (int, error) {
	var n int
	err := r.Data.GetContext(ctx, &n, `SELECT count(*) FROM stock_flag WHERE status = 'open'`)
	return n, err
}

// GetStockFlagsForUpdateTx locks the flags so two managers can't handle the same one.
func (r *repo) GetStockFlagsForUpdateTx(
	ctx context.Context, tx *sqlx.Tx, ids []int,
) ([]types.StockFlag, error) {
	var flags []types.StockFlag
	err := tx.SelectContext(ctx, &flags, `
		SELECT id, product_id, note, counted_quantity, stock_at_flag, created_at, status
		FROM stock_flag WHERE id = ANY($1) ORDER BY id FOR UPDATE`, pq.Array(ids))
	return flags, err
}

func (r *repo) ResolveStockFlagsTx(
	ctx context.Context, tx *sqlx.Tx, ids []int, adjustmentID, userID int,
) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE stock_flag
		SET status = 'resolved', resolved_by_id = $2, resolved_at = NOW(),
			stock_adjustment_id = $3
		WHERE id = ANY($1)`, pq.Array(ids), userID, adjustmentID)
	return err
}

func (r *repo) DismissStockFlagTx(
	ctx context.Context, tx *sqlx.Tx, id, userID int, note string,
) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE stock_flag
		SET status = 'dismissed', resolved_by_id = $2, resolved_at = NOW(), resolution_note = $3
		WHERE id = $1`, id, userID, note)
	return err
}
