package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"pharmacy/httperror"
	"pharmacy/internal/constant"
	"pharmacy/internal/types"
	"pharmacy/model"
	"pharmacy/repository"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jmoiron/sqlx"
)

const (
	maxStockAdjustmentItems = 200
	maxStockNote            = 500
	stockAdjustmentsPerPage = 20
	// maxStockQuantity keeps quantities well inside the INTEGER columns they are stored in.
	maxStockQuantity = 1_000_000
	// maxStockAdjustmentPage keeps the history offset from overflowing.
	maxStockAdjustmentPage = 1_000_000
)

// stockAdjustmentReasons are the reasons a manager can give; "other" needs a note.
var stockAdjustmentReasons = map[string]bool{
	"expired": true, "damaged": true, "theft_loss": true, "count_correction": true,
	"opening_stock": true, "internal_use": true, "other": true,
}

func requireStockAdjuster(ctx context.Context) error {
	if !HasPermission(ctx, constant.AdjustStockPermissionKey) {
		return httperror.Forbidden("only a manager can adjust stock", nil)
	}
	return nil
}

// plannedStockItem is one validated line of an adjustment request.
type plannedStockItem struct {
	productID int
	mode      string
	quantity  int
	expiry    *time.Time
}

// normalizeStockNote trims a note and checks its length; required notes must not be empty, and
// missing is the message when one is.
func normalizeStockNote(note string, required bool, missing string) (string, error) {
	note = strings.TrimSpace(note)
	if required && note == "" {
		return "", badRequest(missing)
	}
	if utf8.RuneCountInString(note) > maxStockNote {
		return "", badRequest(fmt.Sprintf("the note is limited to %d characters", maxStockNote))
	}
	return note, nil
}

func normalizeStockAdjustment(
	req types.StockAdjustmentRequest,
) (types.StockAdjustmentRequest, []plannedStockItem, error) {
	req.Reason = strings.ToLower(strings.TrimSpace(req.Reason))
	req.IdempotencyKey = strings.TrimSpace(req.IdempotencyKey)
	if !stockAdjustmentReasons[req.Reason] {
		return req, nil, badRequest("choose a reason for the adjustment")
	}
	note, err := normalizeStockNote(
		req.Note, req.Reason == "other", "say what the adjustment is for",
	)
	if err != nil {
		return req, nil, err
	}
	req.Note = note
	if len(req.IdempotencyKey) > 100 {
		return req, nil, badRequest("invalid idempotency key")
	}
	if len(req.Items) == 0 {
		return req, nil, badRequest("add at least one product")
	}
	if len(req.Items) > maxStockAdjustmentItems {
		return req, nil, badRequest(
			fmt.Sprintf("adjust at most %d products at a time", maxStockAdjustmentItems),
		)
	}

	items := make([]plannedStockItem, 0, len(req.Items))
	seen := make(map[int]bool, len(req.Items))
	for i, in := range req.Items {
		line := fmt.Sprintf("line %d", i+1)
		if in.ProductID <= 0 {
			return req, nil, badRequest(line + ": choose a product")
		}
		if seen[in.ProductID] {
			return req, nil, badRequest(line + ": that product is already in this adjustment")
		}
		seen[in.ProductID] = true
		switch in.Mode {
		case "set":
			if in.Quantity < 0 {
				return req, nil, badRequest(line + ": the counted quantity can't be negative")
			}
		case "add", "remove":
			if in.Quantity < 1 {
				return req, nil, badRequest(line + ": enter a quantity of at least 1")
			}
		default:
			return req, nil, badRequest(line + ": choose set, add or remove")
		}
		if in.Quantity > maxStockQuantity {
			return req, nil, badRequest(
				fmt.Sprintf("%s: the quantity is limited to %d", line, maxStockQuantity),
			)
		}
		item := plannedStockItem{productID: in.ProductID, mode: in.Mode, quantity: in.Quantity}
		if in.Expiry != nil && strings.TrimSpace(*in.Expiry) != "" {
			expiry, err := time.Parse("2006-01-02", strings.TrimSpace(*in.Expiry))
			if err != nil {
				return req, nil, badRequest(line + ": the expiry date must be YYYY-MM-DD")
			}
			// Expired stock can be corrected to a past date, but not to a typo decades off.
			now := time.Now()
			if expiry.Before(now.AddDate(-10, 0, 0)) || expiry.After(now.AddDate(20, 0, 0)) {
				return req, nil, badRequest(line + ": check the expiry date")
			}
			item.expiry = &expiry
		}
		items = append(items, item)
	}
	return req, items, nil
}

// CreateStockAdjustment corrects the stock of one or more products for one reason. The products
// are locked and their stock read inside the transaction, so concurrent adjustments of a product
// queue up and each sees the last one's result. Sales, receiving and stock taking don't take the
// lock yet (T-022), so a sale committing at the same moment is not counted. A retry with the same
// idempotency key returns the first adjustment.
func (s *inventoryService) CreateStockAdjustment(
	ctx context.Context, req types.StockAdjustmentRequest, userID int,
) (int, error) {
	if err := requireStockAdjuster(ctx); err != nil {
		return 0, err
	}
	req, planned, err := normalizeStockAdjustment(req)
	if err != nil {
		return 0, err
	}

	var id int
	err = s.inTx(ctx, "failed to save stock adjustment", func(tx *sqlx.Tx) error {
		var err error
		id, err = s.createStockAdjustmentTx(ctx, tx, req, planned, userID)
		return err
	})
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (s *inventoryService) createStockAdjustmentTx(
	ctx context.Context, tx *sqlx.Tx, req types.StockAdjustmentRequest,
	planned []plannedStockItem, userID int,
) (int, error) {
	ids := make([]int, len(planned))
	for i, p := range planned {
		ids[i] = p.productID
	}
	sort.Ints(ids)
	locked, err := s.repo.LockProductStockTx(ctx, tx, ids)
	if err != nil {
		return 0, httperror.ServerError("failed to read stock", err)
	}
	// A retry of a save that already went through returns it, before the items are checked
	// against stock that already includes it. Checked after the locks, so a copy that committed
	// while this one waited is visible.
	if req.IdempotencyKey != "" {
		existing, err := s.repo.StockAdjustmentIDByKeyTx(ctx, tx, userID, req.IdempotencyKey)
		if err != nil {
			return 0, httperror.ServerError("failed to check for a repeated request", err)
		}
		if existing != 0 {
			return existing, nil
		}
	}
	products := make(map[int]types.ProductStock, len(locked))
	for _, p := range locked {
		products[p.ID] = p
	}
	if err := s.refuseProductsBeingCounted(ctx, tx, ids, products); err != nil {
		return 0, err
	}

	items := make([]types.NewStockAdjustmentItem, 0, len(planned))
	for i, p := range planned {
		product, ok := products[p.productID]
		if !ok {
			return 0, badRequest(fmt.Sprintf("line %d: product not found", i+1))
		}
		change := p.quantity
		switch p.mode {
		case "set":
			change = p.quantity - product.Stock
		case "remove":
			change = -p.quantity
		}
		if product.Stock+change < 0 {
			return 0, badRequest(
				fmt.Sprintf("%s: only %d in stock", product.Name, product.Stock),
			)
		}
		item := types.NewStockAdjustmentItem{
			ProductID:      product.ID,
			QuantityBefore: product.Stock,
			QuantityChange: change,
			UnitCostKobo:   product.CostPriceKobo,
			ExpiryBefore:   product.CurrentExpiry,
		}
		if p.expiry != nil && !sameDay(product.CurrentExpiry, *p.expiry) {
			item.ExpiryAfter = p.expiry
		}
		if item.QuantityChange == 0 && item.ExpiryAfter == nil {
			if p.expiry != nil {
				return 0, badRequest(fmt.Sprintf(
					"%s: nothing to change, the expiry is already %s",
					product.Name, p.expiry.Format("2006-01-02"),
				))
			}
			return 0, badRequest(fmt.Sprintf("%s: nothing to change", product.Name))
		}
		items = append(items, item)
	}

	flags, err := s.lockFlagsForAdjustment(ctx, tx, req.FlagIDs, products)
	if err != nil {
		return 0, err
	}

	header := types.NewStockAdjustment{Reason: req.Reason, Note: req.Note, CreatedByID: userID}
	if req.IdempotencyKey != "" {
		header.IdempotencyKey = &req.IdempotencyKey
	}
	adjustmentID, created, err := s.repo.CreateStockAdjustmentTx(ctx, tx, header)
	if err != nil {
		return 0, httperror.ServerError("failed to save stock adjustment", err)
	}
	if !created {
		// The same request already went through; don't apply it twice.
		return adjustmentID, nil
	}

	var movements []model.StockMovement
	summary := make([]map[string]any, 0, len(items))
	for _, item := range items {
		if err = s.repo.InsertStockAdjustmentItemTx(ctx, tx, adjustmentID, item); err != nil {
			return 0, httperror.ServerError("failed to save stock adjustment", err)
		}
		if item.QuantityChange != 0 {
			movement := model.StockMovement{
				ProductID:    item.ProductID,
				Quantity:     item.QuantityChange,
				ReferenceID:  adjustmentID,
				MovementType: model.MovementTypeInManualAdjustment,
			}
			if item.QuantityChange < 0 {
				movement.Quantity = -item.QuantityChange
				movement.MovementType = model.MovementTypeOutManualAdjustment
			}
			movements = append(movements, movement)
		}
		if item.ExpiryAfter != nil {
			err = s.repo.SetProductCurrentExpiryTx(ctx, tx, item.ProductID, *item.ExpiryAfter)
			if err != nil {
				return 0, httperror.ServerError("failed to update expiry", err)
			}
		}
		line := map[string]any{
			"product_id": item.ProductID,
			"product":    products[item.ProductID].Name,
			"before":     item.QuantityBefore,
			"change":     item.QuantityChange,
			"after":      item.QuantityBefore + item.QuantityChange,
		}
		if item.ExpiryAfter != nil {
			line["expiry"] = change{From: item.ExpiryBefore, To: item.ExpiryAfter}
		}
		summary = append(summary, line)
	}
	if len(movements) > 0 {
		if err = s.repo.BulkCreateStockMovementTx(ctx, tx, movements); err != nil {
			log.Println(err)
			return 0, httperror.ServerError("failed to record stock movements", err)
		}
	}

	if len(flags) > 0 {
		if err = s.repo.ResolveStockFlagsTx(
			ctx,
			tx,
			req.FlagIDs,
			adjustmentID,
			userID,
		); err != nil {
			return 0, httperror.ServerError("failed to resolve flags", err)
		}
		for _, f := range flags {
			details := map[string]any{"stock_adjustment_id": adjustmentID}
			err = s.audit(ctx, tx, userID, constant.AuditEntityStockFlag, f.ID,
				constant.AuditResolved, "", details)
			if err != nil {
				return 0, err
			}
		}
	}

	details := map[string]any{"reason": req.Reason, "items": summary, "flag_ids": req.FlagIDs}
	err = s.audit(ctx, tx, userID, constant.AuditEntityStockAdjustment, adjustmentID,
		constant.AuditCreated, req.Note, details)
	if err != nil {
		return 0, err
	}

	return adjustmentID, nil
}

// lockFlagsForAdjustment locks the flags an adjustment says it answers. Each must exist, still be
// open, and be about a product in the adjustment.
func (s *inventoryService) lockFlagsForAdjustment(
	ctx context.Context, tx *sqlx.Tx, ids []int, products map[int]types.ProductStock,
) ([]types.StockFlag, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	unique := make(map[int]bool, len(ids))
	for _, id := range ids {
		if unique[id] {
			return nil, badRequest("a flag is listed twice")
		}
		unique[id] = true
	}
	flags, err := s.repo.GetStockFlagsForUpdateTx(ctx, tx, ids)
	if err != nil {
		return nil, httperror.ServerError("failed to read flags", err)
	}
	if len(flags) != len(ids) {
		return nil, httperror.NotFound("flag not found", nil)
	}
	for _, f := range flags {
		if f.Status != "open" {
			return nil, httperror.NewHttpError(
				http.StatusConflict, "a flag was already handled by someone else", nil,
			)
		}
		if _, ok := products[f.ProductID]; !ok {
			return nil, badRequest("a flag is for a product that isn't in this adjustment")
		}
	}
	return flags, nil
}

func sameDay(a *time.Time, b time.Time) bool {
	if a == nil {
		return false
	}
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}

// refuseProductsBeingCounted stops an adjustment of a product already counted in an open stock
// taking: completing the taking applies (count - snapshot), so a change made after the snapshot
// would be applied twice.
func (s *inventoryService) refuseProductsBeingCounted(
	ctx context.Context, tx *sqlx.Tx, ids []int, products map[int]types.ProductStock,
) error {
	counting, err := s.repo.ProductsInActiveStockTakingTx(ctx, tx, ids)
	if err != nil {
		return httperror.ServerError("failed to check the stock taking", err)
	}
	if len(counting) == 0 {
		return nil
	}
	names := make([]string, len(counting))
	for i, id := range counting {
		names[i] = products[id].Name
	}
	return httperror.NewHttpError(http.StatusConflict, fmt.Sprintf(
		"%s already counted in the open stock taking; correct the count there, "+
			"or adjust after the stock taking is completed", strings.Join(names, ", "),
	), nil)
}

func (s *inventoryService) ListStockAdjustments(
	ctx context.Context, page int,
) (types.StockAdjustmentPage, error) {
	if err := requireStockAdjuster(ctx); err != nil {
		return types.StockAdjustmentPage{}, err
	}
	page = min(max(page, 1), maxStockAdjustmentPage)
	rows, total, err := s.repo.ListStockAdjustments(
		ctx, stockAdjustmentsPerPage, (page-1)*stockAdjustmentsPerPage,
	)
	if err != nil {
		return types.StockAdjustmentPage{}, httperror.ServerError(
			"failed to fetch stock adjustments", err,
		)
	}
	return types.StockAdjustmentPage{
		Data: rows, Page: page, PerPage: stockAdjustmentsPerPage, TotalCount: total,
	}, nil
}

// FlagStock asks a manager to check a product's stock. Anyone who can flag or adjust may.
func (s *inventoryService) FlagStock(
	ctx context.Context, req types.StockFlagRequest, userID int,
) (int, error) {
	if !HasPermission(ctx, constant.FlagStockPermissionKey) &&
		!HasPermission(ctx, constant.AdjustStockPermissionKey) {
		return 0, httperror.Forbidden("you cannot flag stock", nil)
	}
	note, err := normalizeStockNote(req.Note, true, "say what looks wrong")
	if err != nil {
		return 0, err
	}
	req.Note = note
	if req.CountedQuantity != nil && *req.CountedQuantity < 0 {
		return 0, badRequest("the counted quantity can't be negative")
	}
	if req.CountedQuantity != nil && *req.CountedQuantity > maxStockQuantity {
		return 0, badRequest(
			fmt.Sprintf("the counted quantity is limited to %d", maxStockQuantity),
		)
	}
	if req.ProductID <= 0 {
		return 0, badRequest("choose a product")
	}

	var id int
	err = s.inTx(ctx, "failed to flag stock", func(tx *sqlx.Tx) error {
		var err error
		id, err = s.flagStockTx(ctx, tx, req, userID)
		return err
	})
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (s *inventoryService) flagStockTx(
	ctx context.Context, tx *sqlx.Tx, req types.StockFlagRequest, userID int,
) (int, error) {
	id, err := s.repo.CreateStockFlagTx(ctx, tx, req, userID)
	if errors.Is(err, repository.ErrInvalidReference) {
		return 0, badRequest("product not found")
	}
	if errors.Is(err, repository.ErrAlreadyFlagged) {
		return 0, httperror.NewHttpError(http.StatusConflict,
			"you already flagged this product; it's waiting for a manager to check", err)
	}
	if err != nil {
		return 0, httperror.ServerError("failed to flag stock", err)
	}
	details := map[string]any{"product_id": req.ProductID, "counted_quantity": req.CountedQuantity}
	err = s.audit(ctx, tx, userID, constant.AuditEntityStockFlag, id,
		constant.AuditCreated, req.Note, details)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (s *inventoryService) ListStockFlags(ctx context.Context) ([]types.StockFlag, error) {
	if err := requireStockAdjuster(ctx); err != nil {
		return nil, err
	}
	flags, err := s.repo.ListOpenStockFlags(ctx)
	if err != nil {
		return nil, httperror.ServerError("failed to fetch flags", err)
	}
	return flags, nil
}

func (s *inventoryService) CountStockFlags(ctx context.Context) (int, error) {
	if err := requireStockAdjuster(ctx); err != nil {
		return 0, err
	}
	n, err := s.repo.CountOpenStockFlags(ctx)
	if err != nil {
		return 0, httperror.ServerError("failed to count flags", err)
	}
	return n, nil
}

// DismissStockFlag closes a flag without adjusting stock, with a note saying why.
func (s *inventoryService) DismissStockFlag(
	ctx context.Context, id int, note string, userID int,
) error {
	if err := requireStockAdjuster(ctx); err != nil {
		return err
	}
	note, err := normalizeStockNote(note, true, "say why the flag is dismissed")
	if err != nil {
		return err
	}

	return s.inTx(ctx, "failed to dismiss flag", func(tx *sqlx.Tx) error {
		return s.dismissStockFlagTx(ctx, tx, id, note, userID)
	})
}

func (s *inventoryService) dismissStockFlagTx(
	ctx context.Context, tx *sqlx.Tx, id int, note string, userID int,
) error {
	flags, err := s.repo.GetStockFlagsForUpdateTx(ctx, tx, []int{id})
	if err != nil {
		return httperror.ServerError("failed to read flag", err)
	}
	if len(flags) == 0 {
		return httperror.NotFound("flag not found", nil)
	}
	if flags[0].Status != "open" {
		return httperror.NewHttpError(
			http.StatusConflict, "this flag was already handled by someone else", nil,
		)
	}
	if err = s.repo.DismissStockFlagTx(ctx, tx, id, userID, note); err != nil {
		return httperror.ServerError("failed to dismiss flag", err)
	}
	err = s.audit(ctx, tx, userID, constant.AuditEntityStockFlag, id,
		constant.AuditDismissed, note, map[string]any{"product_id": flags[0].ProductID})
	if err != nil {
		return err
	}
	return nil
}
