package repository

import (
	"context"
	"database/sql"
	"errors"
	"pharmacy/internal/types"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// ErrInvalidReference is returned when an insert or update points at a row that does not exist
// (a foreign key violation), for example a purchase order line's product or a new user's role.
var ErrInvalidReference = errors.New("referenced row does not exist")

// ErrPurchaseOrderItemNotReceivable is returned when a receipt does not fit a purchase order
// line: wrong order, different product, or more than the outstanding quantity.
var ErrPurchaseOrderItemNotReceivable = errors.New("purchase order item cannot receive quantity")

type purchaseOrderItemRow struct {
	types.PurchaseOrderItem
	PurchaseOrderID int `db:"purchase_order_id"`
}

func mapForeignKeyError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23503" {
		return ErrInvalidReference
	}
	return err
}

// FindOrCreateSupplierTx returns the supplier with this name, ignoring case and surrounding
// spaces, creating it when it does not exist yet. Insert and select are separate statements
// so a concurrent creator that committed first is visible to the select.
func (r *repo) FindOrCreateSupplierTx(
	ctx context.Context, tx *sqlx.Tx, name string,
) (types.Supplier, error) {
	var supplier types.Supplier
	if _, err := tx.ExecContext(ctx, insertSupplierQuery, name); err != nil {
		return supplier, err
	}
	err := tx.GetContext(ctx, &supplier, getSupplierByNameQuery, name)
	return supplier, err
}

func (r *repo) CreatePurchaseOrderTx(
	ctx context.Context, tx *sqlx.Tx, order types.NewPurchaseOrder,
) (int, error) {
	var id int
	err := tx.QueryRowContext(
		ctx, createPurchaseOrderQuery, order.SupplierName, order.SupplierID,
		order.SourceOrderID, order.Note, order.CreatedByID, order.Status,
	).Scan(&id)
	return id, err
}

func (r *repo) GetPurchaseOrderForUpdateTx(
	ctx context.Context, tx *sqlx.Tx, id int,
) (types.PurchaseOrder, error) {
	var order types.PurchaseOrder
	if err := tx.GetContext(ctx, &order, lockPurchaseOrderQuery, id); err != nil {
		return order, err
	}
	items, err := r.fetchPurchaseOrderItems(ctx, tx, []int{id})
	order.Items = items[id]
	return order, err
}

func (r *repo) UpdatePurchaseOrderHeaderTx(
	ctx context.Context, tx *sqlx.Tx, id int, header types.NewPurchaseOrder,
) error {
	_, err := tx.ExecContext(
		ctx,
		updatePurchaseOrderHeaderQuery,
		header.SupplierName,
		header.SupplierID,
		header.Note,
		id,
	)
	return err
}

func (r *repo) InsertPurchaseOrderItemTx(
	ctx context.Context, tx *sqlx.Tx, orderID int, item types.PurchaseOrderItem,
) error {
	_, err := tx.ExecContext(
		ctx, insertPurchaseOrderItemQuery, orderID, item.ProductID, item.Name,
		item.Manufacturer, item.Note, item.Quantity,
	)
	return mapForeignKeyError(err)
}

func (r *repo) UpdatePurchaseOrderItemTx(
	ctx context.Context, tx *sqlx.Tx, orderID int, item types.PurchaseOrderItem,
) error {
	_, err := tx.ExecContext(
		ctx, updatePurchaseOrderItemQuery, item.ID, orderID, item.ProductID, item.Name,
		item.Manufacturer, item.Note, item.Quantity,
	)
	return mapForeignKeyError(err)
}

func (r *repo) DeletePurchaseOrderItemTx(
	ctx context.Context, tx *sqlx.Tx, orderID, itemID int,
) error {
	_, err := tx.ExecContext(ctx, deletePurchaseOrderItemQuery, itemID, orderID)
	return err
}

func (r *repo) ReceivePurchaseOrderItemTx(
	ctx context.Context, tx *sqlx.Tx, orderID, itemID, productID, quantity int,
) error {
	res, err := tx.ExecContext(
		ctx, receivePurchaseOrderItemQuery, quantity, itemID, orderID, productID,
	)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return ErrPurchaseOrderItemNotReceivable
	}
	return nil
}

func (r *repo) SetPurchaseOrderStatusTx(
	ctx context.Context, tx *sqlx.Tx, id int, status string,
) error {
	_, err := tx.ExecContext(ctx, setPurchaseOrderStatusQuery, id, status)
	return err
}

func (r *repo) ApprovePurchaseOrderTx(ctx context.Context, tx *sqlx.Tx, id, approverID int) error {
	_, err := tx.ExecContext(ctx, approvePurchaseOrderQuery, id, approverID)
	return err
}

// CountPendingPurchaseOrders counts drafts still being written and orders submitted for
// approval, keyed by status.
func (r *repo) CountPendingPurchaseOrders(ctx context.Context) (map[string]int, error) {
	var rows []struct {
		Status string `db:"status"`
		Count  int    `db:"count"`
	}
	if err := r.Data.SelectContext(ctx, &rows, countPendingPurchaseOrdersQuery); err != nil {
		return nil, err
	}
	counts := map[string]int{}
	for _, row := range rows {
		counts[row.Status] = row.Count
	}
	return counts, nil
}

func (r *repo) RefreshPurchaseOrderStatusTx(ctx context.Context, tx *sqlx.Tx, id int) error {
	_, err := tx.ExecContext(ctx, refreshPurchaseOrderStatusQuery, id)
	return err
}

func (r *repo) ClosePurchaseOrderTx(ctx context.Context, tx *sqlx.Tx, id int) error {
	_, err := tx.ExecContext(ctx, closePurchaseOrderQuery, id)
	return err
}

func (r *repo) PurchaseOrderExists(ctx context.Context, id int) (bool, error) {
	var exists bool
	err := r.Data.GetContext(ctx, &exists, purchaseOrderExistsQuery, id)
	return exists, err
}

// readSnapshot runs fn in one repeatable-read transaction, so the orders and their items it
// reads belong to the same moment even while receipts are committing.
func (r *repo) readSnapshot(ctx context.Context, fn func(tx *sqlx.Tx) error) error {
	tx, err := r.Data.BeginTxx(
		ctx,
		&sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true},
	)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	return fn(tx)
}

// ListPurchaseOrders returns orders newest first. A nil statuses slice means every status
// and a limit of 0 means no limit.
func (r *repo) ListPurchaseOrders(
	ctx context.Context, statuses []string, limit int, createdBy *int,
) ([]types.PurchaseOrder, error) {
	var orders []types.PurchaseOrder
	err := r.readSnapshot(ctx, func(tx *sqlx.Tx) error {
		err := tx.SelectContext(
			ctx, &orders, listPurchaseOrdersQuery, pq.StringArray(statuses), limit, createdBy,
		)
		if err != nil {
			return err
		}
		orders, err = r.attachPurchaseOrderItems(ctx, tx, orders)
		return err
	})
	return orders, err
}

func (r *repo) GetPurchaseOrder(ctx context.Context, id int) (types.PurchaseOrder, error) {
	var order types.PurchaseOrder
	err := r.readSnapshot(ctx, func(tx *sqlx.Tx) error {
		if err := tx.GetContext(ctx, &order, getPurchaseOrderQuery, id); err != nil {
			return err
		}
		orders, err := r.attachPurchaseOrderItems(ctx, tx, []types.PurchaseOrder{order})
		if err != nil {
			return err
		}
		order = orders[0]
		return nil
	})
	return order, err
}

func (r *repo) attachPurchaseOrderItems(
	ctx context.Context, q sqlx.QueryerContext, orders []types.PurchaseOrder,
) ([]types.PurchaseOrder, error) {
	ids := make([]int, len(orders))
	for i, order := range orders {
		ids[i] = order.ID
	}
	items, err := r.fetchPurchaseOrderItems(ctx, q, ids)
	if err != nil {
		return nil, err
	}
	for i := range orders {
		orders[i].Items = items[orders[i].ID]
	}
	return orders, nil
}

// fetchPurchaseOrderItems loads the lines of several orders in one query. Orders without
// lines map to an empty, non-nil slice so they serialize as [] rather than null.
func (r *repo) fetchPurchaseOrderItems(
	ctx context.Context, q sqlx.QueryerContext, orderIDs []int,
) (map[int][]types.PurchaseOrderItem, error) {
	var rows []purchaseOrderItemRow
	if err := sqlx.SelectContext(
		ctx,
		q,
		&rows,
		purchaseOrderItemsQuery,
		pq.Array(orderIDs),
	); err != nil {
		return nil, err
	}
	items := make(map[int][]types.PurchaseOrderItem, len(orderIDs))
	for _, id := range orderIDs {
		items[id] = []types.PurchaseOrderItem{}
	}
	for _, row := range rows {
		items[row.PurchaseOrderID] = append(items[row.PurchaseOrderID], row.PurchaseOrderItem)
	}
	return items, nil
}

// InsertAuditLogTx appends to the audit log inside the caller's transaction, so the record
// commits or rolls back together with the change it describes.
func (r *repo) InsertAuditLogTx(ctx context.Context, tx *sqlx.Tx, entry types.AuditEntry) error {
	details := entry.Details
	if len(details) == 0 {
		details = []byte("{}")
	}
	// NoActor is stored as a null actor; any other ActorID must be a real user.
	actor := sql.NullInt64{Int64: int64(entry.ActorID), Valid: !entry.NoActor}
	_, err := tx.ExecContext(
		ctx, insertAuditLogQuery, actor, entry.EntityType, entry.EntityID,
		entry.Action, entry.Reason, details,
	)
	return err
}

func (r *repo) FetchAuditLog(
	ctx context.Context, entityType string, entityID int,
) ([]types.AuditRecord, error) {
	records := []types.AuditRecord{}
	err := r.Data.SelectContext(ctx, &records, fetchAuditLogQuery, entityType, entityID)
	return records, err
}

// NewStore wraps an existing connection, for callers that manage their own (tests).
func NewStore(db *sqlx.DB) *repo {
	return &repo{Data: db}
}
