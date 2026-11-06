package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"pharmacy/internal/constant"
	"pharmacy/internal/types"
	"time"

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

// LockProductStockTx locks the products in id order, then reads their stock in a second statement,
// which (under READ COMMITTED) sees whatever committed while it waited for the locks.
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

// StockAdjustmentByKeyTx returns the adjustment saved under a retry key, or nil.
func (r *repo) StockAdjustmentByKeyTx(
	ctx context.Context, tx *sqlx.Tx, key string,
) (*types.SavedStockAdjustment, error) {
	var saved types.SavedStockAdjustment
	err := tx.GetContext(ctx, &saved, `
		SELECT id, created_by_id, COALESCE(request_hash, '') AS request_hash
		FROM stock_adjustment WHERE idempotency_key = $1`, key)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &saved, err
}

const createStockAdjustmentQuery = `
	INSERT INTO stock_adjustment (reason, note, created_by_id, idempotency_key, request_hash)
	VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (idempotency_key) WHERE idempotency_key IS NOT NULL DO NOTHING
	RETURNING id
`

// CreateStockAdjustmentTx inserts the adjustment and returns its id and true, or 0 and false when
// another save already used the key (the caller looks that one up).
func (r *repo) CreateStockAdjustmentTx(
	ctx context.Context, tx *sqlx.Tx, adj types.NewStockAdjustment,
) (int, bool, error) {
	var id int
	err := tx.GetContext(
		ctx, &id, createStockAdjustmentQuery,
		adj.Reason, adj.Note, adj.CreatedByID, adj.IdempotencyKey, adj.RequestHash,
	)
	if errors.Is(err, sql.ErrNoRows) && adj.IdempotencyKey != nil {
		return 0, false, nil
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
	if isUniqueViolation(err) {
		return 0, ErrAlreadyFlagged
	}
	return id, mapForeignKeyError(err)
}

// stockFlagColumns sums stock for the flagged products only, not every product through
// inventory_view.
var stockFlagColumns = `
	SELECT f.id, f.product_id, p.name AS product_name,
		COALESCE(p.manufacturer, '') AS manufacturer, f.note, f.counted_quantity,
		f.stock_at_flag, ` + fmt.Sprintf(productStockSQL, "f.product_id") + ` AS current_stock,
		COALESCE(u.username, '') AS flagged_by, f.created_at, f.status
	FROM stock_flag f
	JOIN product p ON p.id = f.product_id
	LEFT JOIN users u ON u.id = f.flagged_by_id
`

// openFlagFilter writes the status into the SQL text rather than binding it, so the planner can
// always use the partial indexes on status = 'open', whatever plan caching does.
var openFlagFilter = ` WHERE f.status = '` + constant.StockFlagOpen + `'`

func (r *repo) ListOpenStockFlags(ctx context.Context) ([]types.StockFlag, error) {
	flags := []types.StockFlag{}
	err := r.Data.SelectContext(ctx, &flags, stockFlagColumns+openFlagFilter+
		` ORDER BY f.created_at DESC`)
	return flags, err
}

func (r *repo) CountOpenStockFlags(ctx context.Context) (int, error) {
	var n int
	err := r.Data.GetContext(ctx, &n, `SELECT count(*) FROM stock_flag f`+openFlagFilter)
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
		SET status = $4, resolved_by_id = $2, resolved_at = NOW(), stock_adjustment_id = $3
		WHERE id = ANY($1)`, pq.Array(ids), userID, adjustmentID, constant.StockFlagResolved)
	return err
}

func (r *repo) DismissStockFlagTx(
	ctx context.Context, tx *sqlx.Tx, id, userID int, note string,
) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE stock_flag
		SET status = $4, resolved_by_id = $2, resolved_at = NOW(), resolution_note = $3
		WHERE id = $1`, id, userID, note, constant.StockFlagDismissed)
	return err
}
