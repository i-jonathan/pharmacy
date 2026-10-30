package service

import (
	"context"
	"encoding/json"
	"errors"
	"pharmacy/httperror"
	"pharmacy/internal/constant"
	"pharmacy/internal/types"
	"pharmacy/model"

	"github.com/jmoiron/sqlx"
)

func HasPermission(ctx context.Context, permissionKey string) bool {
	perms, ok := ctx.Value(constant.PermissionsSessionKey).(map[string]bool)
	if !ok {
		return false
	}

	return perms[permissionKey]
}

func calculateStockDifference(item types.StockTakingItemData) (int, model.MovementType, bool) {
	total := 0
 
	if item.DispensaryCount != nil {
		total += *item.DispensaryCount
	}
	if item.StoreCount != nil {
		total += *item.StoreCount
	}
 
	snapshot := 0
	if item.SnapshotQuantity != nil {
		snapshot = *item.SnapshotQuantity
	}
 
	diff := total - snapshot
	if diff == 0 {
		return 0, "", true
	}
 
	movementType := model.MovementTypeInStockTaking
	if diff < 0 {
		movementType = model.MovementTypeOutStockTaking
		diff = -diff
	}
 
	return diff, movementType, false
}

// badRequest is a 400 whose message is shown to the user as is.
func badRequest(msg string) error {
	return httperror.BadRequest(msg, errors.New(msg))
}

// inTx runs fn in a transaction, committing when it returns nil and rolling back otherwise.
// failure is the message for a failed commit.
func (s *inventoryService) inTx(
	ctx context.Context, failure string, fn func(tx *sqlx.Tx) error,
) error {
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return httperror.ServerError("failed to start transaction", err)
	}
	committed := false
	defer func() {
		if !committed {
			s.repo.RollbackTx(tx)
		}
	}()
	if err = fn(tx); err != nil {
		return err
	}
	if err = s.repo.CommitTx(tx); err != nil {
		return httperror.ServerError(failure, err)
	}
	committed = true
	return nil
}

// audit appends an audit record for any entity inside the caller's transaction.
func (s *inventoryService) audit(
	ctx context.Context, tx *sqlx.Tx, userID int, entity string, entityID int,
	action, reason string, details any,
) error {
	body, err := json.Marshal(details)
	if err != nil {
		return httperror.ServerError("failed to record audit entry", err)
	}
	err = s.repo.InsertAuditLogTx(ctx, tx, types.AuditEntry{
		ActorID: userID, EntityType: entity, EntityID: entityID,
		Action: action, Reason: reason, Details: body,
	})
	if err != nil {
		return httperror.ServerError("failed to record audit entry", err)
	}
	return nil
}
