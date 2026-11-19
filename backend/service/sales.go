package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"pharmacy/httperror"
	"pharmacy/internal/constant"
	"pharmacy/internal/types"
	"pharmacy/model"
	"pharmacy/repository"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jmoiron/sqlx"
)

type saleService struct {
	repo repository.PharmacyRepository
}

func NewSaleService(repo repository.PharmacyRepository) *saleService {
	return &saleService{repo: repo}
}

func (s *saleService) CreateSale(ctx context.Context, saleParams types.Sale) (string, error) {
	if saleParams.CustomerID < 0 {
		return "", badRequest("customer was not found")
	}
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		log.Println(err)
		return "", httperror.ServerError("could not begin transcation", err)
	}
	committed := false
	defer func() {
		if !committed {
			s.repo.RollbackTx(tx)
		}
	}()

	idempotencyKey := strings.TrimSpace(saleParams.IdempotencyKey)

	// A retry of a saved sale is answered with that sale, whatever has happened since to the
	// rights of the user, the customer or the prices, so the key is looked up first.
	keyUsed, err := s.saleKeyUsed(ctx, tx, idempotencyKey)
	if err != nil {
		return "", err
	}

	// Seeing customers is what lets a cashier pick one; a user without it sells to walk-ins. Only a
	// new sale is refused.
	if saleParams.CustomerID > 0 && !keyUsed &&
		!HasPermission(ctx, constant.ViewCustomersPermissionKey) {
		return "", httperror.Forbidden("you are not allowed to attach customers to a sale", nil)
	}

	// The server prices a new sale; what the till sent is only a check against it.
	var priced pricedSale
	if !keyUsed {
		priced, err = s.priceSale(ctx, tx, saleParams)
		if err != nil {
			// A twin of this request may have saved the sale while this one was pricing it, and a
			// price edit since would then read as a refusal. Say the sale was saved instead.
			priceErr := err
			if keyUsed, err = s.saleKeyUsed(ctx, tx, idempotencyKey); !keyUsed {
				if err != nil {
					log.Println("re-checking the sale key after a pricing refusal:", err)
				}
				return "", priceErr
			}
		}
	}

	sale := model.Sale{
		Status:         constant.CompleteSaleStatus,
		CashierID:      saleParams.CashierID,
		Discount:       priced.discount,
		Subtotal:       priced.subtotal,
		Total:          priced.total,
		IdempotencyKey: nil,
	}
	if idempotencyKey != "" {
		sale.IdempotencyKey = &idempotencyKey
	}
	if saleParams.CustomerID > 0 {
		customerID := saleParams.CustomerID
		sale.CustomerID = &customerID
	}
	sale.GenerateReceiptNumber()

	saleID, err := s.repo.CreateSaleTx(ctx, tx, sale)
	if errors.Is(err, repository.ErrInvalidReference) {
		return "", badRequest("customer was not found")
	}
	if err != nil {
		log.Println(err)
		return "", httperror.ServerError("failed to create sale", err)
	}
	if saleID == 0 {
		log.Printf("duplicate sale idempotency key ignored: %s", idempotencyKey)
		// The sale is saved, so a retry succeeds as it always did; only the number may be missing.
		saved, err := s.repo.FetchSaleByKeyTx(ctx, tx, idempotencyKey)
		if err != nil {
			log.Println(err)
			return "", nil
		}
		// A retry that names a different customer is not the sale that was saved: answering it with
		// the first save would leave the receipt and the record disagreeing.
		if !sameCustomer(saved.CustomerID, saleParams.CustomerID) {
			return "", alreadySaved(
				"this sale was already saved with a different customer; check Sales History before selling again",
			)
		}
		return saved.ReceiptNumber, nil
	}

	// Only a new sale is checked: a retry of a saved sale must still be answered with that sale,
	// whatever has happened to its customer since.
	if saleParams.CustomerID > 0 {
		customer, err := s.repo.GetCustomerRefTx(ctx, tx, saleParams.CustomerID)
		if errors.Is(err, sql.ErrNoRows) {
			return "", badRequest("customer was not found")
		}
		if err != nil {
			log.Println(err)
			return "", httperror.ServerError("failed to check the customer", err)
		}
		if !customer.Active {
			refusal := httperror.Conflict(
				"this customer has been deactivated; remove them from the sale, or ask a manager to reactivate them",
				errors.New("customer is deactivated"),
			)
			refusal.Reason = "customer_inactive"
			return "", refusal
		}
	}

	saleItems := priced.items
	stockMovement := make([]model.StockMovement, len(saleItems))
	for i := range saleItems {
		saleItems[i].SaleID = saleID
		stockMovement[i] = model.StockMovement{
			ProductID:    saleItems[i].ProductID,
			Quantity:     saleItems[i].Quantity * saleItems[i].StockQuantityPerUnit,
			ReferenceID:  saleID,
			MovementType: model.MovementTypeOutSale,
		}
	}

	err = s.repo.BulkCreateSaleItemsTx(ctx, tx, saleItems)
	if err != nil {
		log.Println(err)
		return "", httperror.ServerError("Failed to create sale items", err)
	}

	salePayments := priced.payments
	for i := range salePayments {
		salePayments[i].SaleID = saleID
	}

	err = s.repo.BulkCreateSalePaymentsTX(ctx, tx, salePayments)
	if err != nil {
		log.Println(err)
		return "", httperror.ServerError("Failed to bulk create payments", err)
	}

	// create bulk stock movements
	err = s.repo.BulkCreateStockMovementTx(ctx, tx, stockMovement)
	if err != nil {
		log.Println(err)
		return "", httperror.ServerError("Failed to bulk create stock movements", err)
	}

	// Delete held sale tx
	log.Printf("Sale reference: %s\n", saleParams.HeldSaleReference)
	if saleParams.HeldSaleReference != "" {
		err = s.repo.DeleteHeldTransactionByReferenceTx(ctx, tx, saleParams.HeldSaleReference)
		if err != nil {
			log.Println(err)
		}
	}
	err = s.repo.CommitTx(tx)
	if err != nil {
		log.Println(err)
		return "", httperror.ServerError("failed to commit transaction", err)
	}
	committed = true
	return sale.ReceiptNumber, nil
}

func (s *saleService) FetchSalesHistory(ctx context.Context, filter types.SaleFilter) (types.SaleHistory, error) {
	var (
		totalCount        int
		responses         []types.SaleResponse
		canViewTotal      bool
		salesHistoryTotal int
	)
	err := runInTx(ctx, s.repo, "failed to commit transaction", func(tx *sqlx.Tx) error {
		// fetch sales
		sales, err := s.repo.FetchSalesTx(ctx, tx, filter)
		if err != nil {
			log.Println(err)
			return httperror.ServerError("fetching sales failed", err)
		}

		totalCount, err = s.repo.CountSales(ctx, filter)
		if err != nil {
			log.Println(err)
			return httperror.ServerError("counting sales failed", err)
		}

		// index sales by sale id and fetch all sale id's to use in getting sale items and payments
		salesByID := make(map[int]*model.Sale)
		saleIDs := make([]int, len(sales))
		cashierIDs := make(map[int]struct{})
		for i, s := range sales {
			salesByID[sales[i].ID] = &sales[i]
			saleIDs[i] = sales[i].ID
			cashierIDs[s.CashierID] = struct{}{}
		}

		// fetch sale items, from sale items, extract product id's.
		// Then, assign sale items to the already indexed sales map created earlier
		saleItems, err := s.repo.BulkFetchSaleItems(ctx, tx, saleIDs)
		if err != nil {
			log.Println(err)
			return httperror.ServerError("fetching sale items failed", err)
		}

		productIDs := make([]int, len(saleItems))
		for i, item := range saleItems {
			salesByID[item.SaleID].SaleItems = append(salesByID[item.SaleID].SaleItems, item)
			productIDs[i] = item.ProductID
		}

		// fetch all sales payments, and assign to the already indexed sales. Again.
		payments, err := s.repo.BulkFetchSalePayments(ctx, tx, saleIDs)
		if err != nil {
			log.Println(err)
			return httperror.ServerError("fetching sale payments failed", err)
		}

		for i := range payments {
			salesByID[payments[i].SaleID].Payments = append(salesByID[payments[i].SaleID].Payments, payments[i])
		}

		// fetch return items (joined with returns table to get sale_id)
		returns, err := s.repo.BulkFetchReturnItemsBySaleIDs(ctx, saleIDs)
		if err != nil {
			log.Println(err)
			return httperror.ServerError("fetching return items failed", err)
		}

		// group return items by sale_id
		returnsBySaleID := make(map[int][]model.ReturnItemWithSale)
		returnedQuantityBySaleItemID := make(map[int]int)
		for _, r := range returns {
			returnsBySaleID[r.SaleID] = append(returnsBySaleID[r.SaleID], r)
			returnedQuantityBySaleItemID[r.SaleItemID] += r.Quantity
		}

		// fetch & index products to prepare for final structuring
		products, err := s.repo.BulkFetchProductByIDsTx(ctx, tx, productIDs)
		if err != nil {
			log.Println(err)
			return httperror.ServerError("fetching products failed", err)
		}

		productsByID := make(map[int]model.Product)
		for _, p := range products {
			productsByID[p.ID] = p
		}

		ids := make([]int, 0, len(cashierIDs))
		for id := range cashierIDs {
			ids = append(ids, id)
		}

		cashiers, err := s.repo.BulkFetchUserByIDTx(ctx, tx, ids)
		if err != nil {
			log.Println(err)
			return httperror.ServerError("fetching cashiers failed", err)
		}

		cashiersByID := make(map[int]string)
		for _, c := range cashiers {
			cashiersByID[c.ID] = c.UserName
		}

		// Customers are personal data: only users who may see customers get them.
		changeRows, err := s.customerChangeRowsFor(ctx, tx, saleIDs)
		if err != nil {
			log.Println(err)
			return httperror.ServerError("fetching customer changes failed", err)
		}
		changes, changedCustomerIDs := decodeCustomerChanges(changeRows)
		customerNames, err := s.customerNamesFor(ctx, tx, sales, changedCustomerIDs)
		if err != nil {
			log.Println(err)
			return httperror.ServerError("fetching customers failed", err)
		}

		responses = make([]types.SaleResponse, 0, len(sales))
		canViewTotal = HasPermission(ctx, constant.ViewSalesTotalPermissionKey)

		if canViewTotal {
			t, err := s.repo.SumSalesTotal(ctx, filter)
			if err != nil {
				log.Println(err)
				return httperror.ServerError("summing sales total failed", err)
			}
			salesHistoryTotal = t

			// Subtract return refunds so the period total reflects net sales
			refundTotal, err := s.repo.SumReturnTotal(ctx, filter)
			if err != nil {
				log.Println(err)
				return httperror.ServerError("summing return total failed", err)
			}
			salesHistoryTotal -= refundTotal
		}

		for _, s := range sales {
			saleItemsByID := make(map[int]model.SaleItem)

			// build items
			items := make([]types.SaleItemResponse, 0, len(s.SaleItems))
			for _, item := range s.SaleItems {
				p := productsByID[item.ProductID]
				items = append(items, types.SaleItemResponse{
					ID:               item.ID,
					ProductName:      p.Name,
					Manufacturer:     manufacturerName(p),
					Quantity:         item.Quantity,
					ReturnedQuantity: returnedQuantityBySaleItemID[item.ID],
					UnitPrice:        float64(item.UnitPrice) / 100,
					Discount:         float64(item.Discount) / 100,
				})

				saleItemsByID[item.ID] = item
			}

			// build payments
			payments := make([]types.SalePaymentResponse, 0, len(s.Payments))
			for _, pay := range s.Payments {
				payments = append(payments, types.SalePaymentResponse{
					MethodName: pay.PaymentMethod,
					Amount:     float64(pay.Amount) / 100,
				})
			}

			// build returns
			returnsResp := make([]types.ReturnItemResponse, 0, len(returnsBySaleID[s.ID]))
			for _, r := range returnsBySaleID[s.ID] {
				item := saleItemsByID[r.SaleItemID]
				p := productsByID[item.ProductID]
				returnsResp = append(returnsResp, types.ReturnItemResponse{
					SaleItemID:   r.SaleItemID,
					Name:         p.Name,
					Manufacturer: manufacturerName(p),
					UnitPrice:    float64(item.UnitPrice) / 100,
					Quantity:     r.Quantity,
				})

			}

			// build the final response
			resp := types.SaleResponse{
				ID:            s.ID,
				ReceiptNumber: s.ReceiptNumber,
				Cashier:       cashiersByID[s.CashierID],
				CreatedAt:     s.CreatedAt.Format(time.RFC3339),
				Subtotal:      float64(s.Subtotal) / 100,
				Discount:      float64(s.Discount) / 100,
				Total:         float64(s.Total) / 100,
				Items:         items,
				Payments:      payments,
				Returns:       returnsResp,
				IsRecent:      s.IsRecent,
			}
			if s.CustomerID != nil {
				if name, ok := customerNames[*s.CustomerID]; ok {
					resp.Customer = &types.SaleCustomer{ID: *s.CustomerID, Name: name}
				}
			}
			resp.CustomerChanges = customerChangesOf(changes[s.ID], customerNames)
			responses = append(responses, resp)
		}
		return nil
	})
	if err != nil {
		return types.SaleHistory{}, err
	}

	history := types.SaleHistory{
		Data:       responses,
		Page:       filter.Page,
		PerPage:    filter.PerPage,
		TotalCount: totalCount,
	}

	if canViewTotal {
		history.TotalAmount = float64(salesHistoryTotal) / 100
	}

	return history, nil
}

// maxSaleCustomerNote is the longest note given with a change of a sale's customer; the page's
// box allows the same.
const maxSaleCustomerNote = 250

// saleCustomerChangeReasons are the reasons given for changing the customer on a saved sale;
// "other" needs a note.
var saleCustomerChangeReasons = map[string]bool{
	"picked_by_mistake": true, "customer_asked": true, "other": true,
}

// ChangeSaleCustomer changes or removes the customer on a saved sale. The sale is locked while it
// is read, so two people fixing it cannot both succeed blind: the page says which customer it saw.
func (s *saleService) ChangeSaleCustomer(
	ctx context.Context, req types.ChangeSaleCustomerRequest, userID int,
) (*types.SaleCustomer, error) {
	if !HasPermission(ctx, constant.ManageCustomersPermissionKey) {
		return nil, httperror.Forbidden("you are not allowed to change a sale's customer", nil)
	}
	// Seeing customers comes with the right to change them: without it the customer on a sale is
	// never shown, and a refusal would reveal one.
	if !HasPermission(ctx, constant.ViewCustomersPermissionKey) {
		return nil, httperror.Forbidden(
			"you are not allowed to see or change a sale's customer", nil,
		)
	}
	if req.SaleID <= 0 {
		return nil, badRequest("sale was not found")
	}
	if req.CustomerID < 0 || req.ExpectedCustomerID < 0 {
		return nil, badRequest("customer was not found")
	}
	req.Reason = strings.ToLower(strings.TrimSpace(req.Reason))
	if !saleCustomerChangeReasons[req.Reason] {
		return nil, badRequest("choose a reason for the change")
	}
	note := strings.TrimSpace(req.Note)
	if note == "" && req.Reason == "other" {
		return nil, badRequest("say why the customer is changed")
	}
	if utf8.RuneCountInString(note) > maxSaleCustomerNote {
		return nil, badRequest(
			fmt.Sprintf("the note is limited to %d characters", maxSaleCustomerNote),
		)
	}

	var result *types.SaleCustomer
	err := runInTx(ctx, s.repo, "failed to save the customer", func(tx *sqlx.Tx) error {
		sale, err := s.repo.FetchSaleCustomerForUpdateTx(ctx, tx, req.SaleID)
		if errors.Is(err, sql.ErrNoRows) {
			return httperror.NotFound("sale was not found", err)
		}
		if err != nil {
			return httperror.ServerError("failed to load the sale", err)
		}
		// A sale that already has the requested customer is answered as saved, whatever the page
		// expected: a retry after a lost reply (or a double click) sends the old expectation and
		// must not be told that someone else changed it. It writes and audits nothing, and shows
		// only the customer Sales History already shows, so the age of the sale does not matter: a
		// retry that arrives after the sale has aged out is still the change that was saved.
		unchanged := sameCustomer(sale.CustomerID, req.CustomerID)
		// For a real change the age rule comes before the stale check, so a cashier learns nothing
		// about an older sale that Sales History does not already show.
		if !unchanged && !sale.IsRecent &&
			!HasPermission(ctx, constant.ChangeOldSaleCustomerPermissionKey) {
			refusal := httperror.Forbidden(
				"only a manager can change the customer on a sale more than 24 hours old", nil,
			)
			refusal.Reason = "sale_too_old"
			return refusal
		}
		if !unchanged && !sameCustomer(sale.CustomerID, req.ExpectedCustomerID) {
			msg := "someone else changed this sale's customer; check who is on it now before changing it again"
			refusal := httperror.Conflict(msg, errors.New(msg))
			refusal.Reason = "customer_changed"
			return refusal
		}

		if req.CustomerID > 0 {
			customer, err := s.repo.GetCustomerRefTx(ctx, tx, req.CustomerID)
			if errors.Is(err, sql.ErrNoRows) {
				return badRequest("customer was not found")
			}
			if err != nil {
				return httperror.ServerError("failed to check the customer", err)
			}
			// Only a new attachment needs an active customer: a repeat of a change that already
			// happened is answered as saved even if the customer has been deactivated since.
			if !unchanged && !customer.Active {
				refusal := httperror.Conflict(
					"this customer has been deactivated; choose another, or ask a manager to reactivate them",
					errors.New("customer is deactivated"),
				)
				refusal.Reason = "customer_inactive"
				return refusal
			}
			result = &types.SaleCustomer{ID: customer.ID, Name: customer.Name}
		}
		if unchanged {
			return nil
		}

		var target *int
		if req.CustomerID > 0 {
			target = &req.CustomerID
		}
		err = s.repo.UpdateSaleCustomerTx(ctx, tx, req.SaleID, target)
		if errors.Is(err, repository.ErrInvalidReference) {
			return badRequest("customer was not found")
		}
		if err != nil {
			return httperror.ServerError("failed to save the customer", err)
		}
		reason := req.Reason
		if note != "" {
			reason += ": " + note
		}
		from := 0
		if sale.CustomerID != nil {
			from = *sale.CustomerID
		}
		return appendAudit(
			ctx, s.repo, tx, userID, constant.AuditEntitySale, req.SaleID,
			constant.AuditActionSaleCustomerChanged, reason,
			map[string]any{
				"receipt_number":   sale.ReceiptNumber,
				"from_customer_id": from,
				"to_customer_id":   req.CustomerID,
			},
		)
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// saleKeyUsed says whether a sale was already saved under the retry key. An empty key was never
// used.
func (s *saleService) saleKeyUsed(
	ctx context.Context, tx *sqlx.Tx, key string,
) (bool, error) {
	if key == "" {
		return false, nil
	}
	_, err := s.repo.FetchSaleByKeyTx(ctx, tx, key)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, httperror.ServerError("failed to check the sale", err)
	}
	return true, nil
}

// sameCustomer says whether a saved sale's customer (nil for a walk-in) is the one a request names
// (0 for a walk-in).
func sameCustomer(saved *int, requested int) bool {
	if saved == nil {
		return requested <= 0
	}
	return *saved == requested
}

// customerNamesFor looks up, in one query, the customers of the given sales and the extra ones
// (customers named in changes); nothing is looked up for a user who may not see customers.
func (s *saleService) customerNamesFor(
	ctx context.Context, tx *sqlx.Tx, sales []model.Sale, extra []int,
) (map[int]string, error) {
	if !HasPermission(ctx, constant.ViewCustomersPermissionKey) {
		return nil, nil
	}
	seen := map[int]bool{}
	var ids []int
	add := func(id int) {
		if id > 0 && !seen[id] {
			seen[id] = true
			ids = append(ids, id)
		}
	}
	for _, sale := range sales {
		if sale.CustomerID != nil {
			add(*sale.CustomerID)
		}
	}
	for _, id := range extra {
		add(id)
	}
	if len(ids) == 0 {
		return nil, nil
	}
	return s.repo.FetchCustomerNames(ctx, tx, ids)
}

// customerChangeRowsFor reads the customer changes of the given sales in one query. Only users who
// may see customers get them: the entries name customers.
func (s *saleService) customerChangeRowsFor(
	ctx context.Context, tx *sqlx.Tx, saleIDs []int,
) ([]types.SaleCustomerChangeRow, error) {
	if len(saleIDs) == 0 || !HasPermission(ctx, constant.ViewCustomersPermissionKey) {
		return nil, nil
	}
	return s.repo.FetchSaleCustomerChanges(ctx, tx, saleIDs)
}

// customerChangeDetails is what ChangeSaleCustomer stores in an audit entry's details: customer
// ids, 0 for a walk-in.
type customerChangeDetails struct {
	From int `json:"from_customer_id"`
	To   int `json:"to_customer_id"`
}

// decodedCustomerChange is a stored change with its customer ids read out.
type decodedCustomerChange struct {
	row      types.SaleCustomerChangeRow
	from, to int
}

// decodeCustomerChanges groups the stored changes by sale, keeping their order (newest first), and
// lists the customer ids they name. An entry whose details cannot be read is left out and logged,
// so one bad row never costs the whole page.
func decodeCustomerChanges(
	rows []types.SaleCustomerChangeRow,
) (map[int][]decodedCustomerChange, []int) {
	bySale := map[int][]decodedCustomerChange{}
	var ids []int
	for _, row := range rows {
		var details customerChangeDetails
		if err := json.Unmarshal(row.Details, &details); err != nil {
			log.Printf("skipping a customer change on sale %d: %v", row.SaleID, err)
			continue
		}
		bySale[row.SaleID] = append(
			bySale[row.SaleID], decodedCustomerChange{row: row, from: details.From, to: details.To},
		)
		ids = append(ids, details.From, details.To)
	}
	return bySale, ids
}

// customerChangesOf shapes one sale's changes for the page. The list is never nil, so the page
// can read its length.
func customerChangesOf(
	changes []decodedCustomerChange, names map[int]string,
) []types.SaleCustomerChange {
	out := make([]types.SaleCustomerChange, 0, len(changes))
	who := func(id int) *types.SaleCustomer {
		if id <= 0 {
			return nil
		}
		return &types.SaleCustomer{ID: id, Name: names[id]}
	}
	for _, c := range changes {
		reason, note := splitChangeReason(c.row.Reason)
		out = append(out, types.SaleCustomerChange{
			At:     c.row.OccurredAt.Format(time.RFC3339),
			By:     c.row.Actor,
			From:   who(c.from),
			To:     who(c.to),
			Reason: reason,
			Note:   note,
		})
	}
	return out
}

// splitChangeReason undoes how ChangeSaleCustomer stores its reason: the reason code, then ": " and
// the note when there is one. Text that does not start with a known code is all note.
func splitChangeReason(stored string) (reason, note string) {
	if saleCustomerChangeReasons[stored] {
		return stored, ""
	}
	if code, rest, found := strings.Cut(stored, ": "); found && saleCustomerChangeReasons[code] {
		return code, strings.TrimSpace(rest)
	}
	return "", stored
}

// manufacturerName is a product's manufacturer, or empty when none was recorded.
func manufacturerName(p model.Product) string {
	if p.Manufacturer == nil {
		return ""
	}
	return *p.Manufacturer
}

func (s *saleService) ReprintReceipt(
	ctx context.Context, saleID, userID int,
) (types.ReceiptReprint, error) {
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return types.ReceiptReprint{}, httperror.ServerError("could not begin transaction", err)
	}
	committed := false
	defer func() {
		if !committed {
			s.repo.RollbackTx(tx)
		}
	}()

	receipt, err := s.repo.FetchSaleReceiptNumberByID(ctx, tx, saleID)
	if errors.Is(err, sql.ErrNoRows) {
		return types.ReceiptReprint{}, httperror.NotFound("sale was not found", err)
	}
	if err != nil {
		return types.ReceiptReprint{}, httperror.ServerError("failed to load the sale", err)
	}

	details, err := json.Marshal(map[string]any{"receipt_number": receipt})
	if err != nil {
		return types.ReceiptReprint{}, httperror.ServerError("failed to record the reprint", err)
	}
	err = s.repo.InsertAuditLogTx(ctx, tx, types.AuditEntry{
		ActorID: userID, EntityType: constant.AuditEntitySale, EntityID: saleID,
		Action: constant.AuditActionReceiptReprinted, Details: details,
	})
	if err != nil {
		return types.ReceiptReprint{}, httperror.ServerError("failed to record the reprint", err)
	}
	reprintedAt := time.Now()
	if err := s.repo.CommitTx(tx); err != nil {
		return types.ReceiptReprint{}, httperror.ServerError("failed to commit transaction", err)
	}
	committed = true

	return types.ReceiptReprint{
		ReceiptNumber: receipt,
		ReprintedAt:   reprintedAt.Format(time.RFC3339),
	}, nil
}

func (s *saleService) HoldSale(ctx context.Context, holdSaleRequest types.HoldTransactionRequest) error {
	// Validate payload has items
	// The frontend sends payload as JSON.stringify(...) so RawMessage is a JSON string.
	// Unwrap it first.
	payloadBytes := []byte(holdSaleRequest.Payload)
	if len(payloadBytes) > 0 && payloadBytes[0] == '"' {
		var s string
		if err := json.Unmarshal(payloadBytes, &s); err != nil {
			return httperror.BadRequest("invalid payload string", err)
		}
		payloadBytes = []byte(s)
	}

	var payload struct {
		Cart []any `json:"cart"`
	}
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return httperror.BadRequest("invalid payload", err)
	}
	if len(payload.Cart) == 0 {
		return httperror.BadRequest("cannot hold an empty sale", fmt.Errorf("cart is empty"))
	}

	reference := holdSaleRequest.Reference
	if reference == "" {
		reference = fmt.Sprintf("%s-%s-%04d",
			string(constant.HoldSaleType),
			time.Now().Format("20060102"),
			rand.Intn(10000),
		)
	}

	holdTransaction := model.HeldTransaction{
		Type:      string(constant.HoldSaleType),
		Reference: reference,
		Payload:   holdSaleRequest.Payload,
	}

	err := s.repo.SaveHeldTransaction(ctx, holdTransaction)
	if err != nil {
		log.Println(err)
		return httperror.ServerError("failed to hold sale", err)
	}
	return nil
}

func (s *saleService) FetchHeldSaleTransactions(ctx context.Context) ([]types.HeldTransactionResponse, error) {
	transactions, err := s.repo.FetchHeldTransactionsByType(ctx, constant.HoldSaleType)
	if err != nil {
		log.Println(err)
		return nil, httperror.ServerError("failed to fetch held transactions", err)
	}

	var response []types.HeldTransactionResponse
	for _, sale := range transactions {
		response = append(response, types.HeldTransactionResponse{
			Reference: sale.Reference,
			Payload:   sale.Payload,
			CreatedAt: sale.CreatedAt,
			UpdatedAt: sale.UpdatedAt,
		})
	}

	return response, nil
}

func (s *saleService) DeleteHeldTransaction(ctx context.Context, reference string) error {
	err := s.repo.DeleteHeldTransactionByReference(ctx, reference)
	if err != nil {
		log.Println(err)
		return httperror.ServerError("failed to delete held transaction", err)
	}
	return nil
}

func (s *saleService) ReturnItems(ctx context.Context, returnParams types.ReturnSale) error {
	if returnParams.SaleID <= 0 {
		return httperror.BadRequest("sale id must be positive", nil)
	}
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		log.Println(err)
		return httperror.ServerError("could not begin transcation", err)
	}
	committed := false
	defer func() {
		if !committed {
			s.repo.RollbackTx(tx)
		}
	}()

	sale, err := s.repo.FetchSaleByID(ctx, tx, returnParams.SaleID)
	if err != nil {
		log.Println("fetching sale failed: ", err)
		return httperror.ServerError("failed to fetch sale by id", err)
	}

	returnRecord, err := s.repo.FetchAllSaleReturns(ctx, tx, returnParams.SaleID)
	if err != nil {
		log.Println("fetching return record failed: ", err)
		return httperror.ServerError("failed to fetch return records by sale id", err)
	}
	returnMap := make(map[int]int)
	for _, rr := range returnRecord {
		returnMap[rr.SaleItemID] += rr.Quantity
	}

	validItems := make(map[int]model.SaleItem)
	for _, item := range sale.SaleItems {
		validItems[item.ID] = item
	}

	refundTotal := 0
	var returnItems []model.ReturnItems
	var stockMovement []model.StockMovement
	requestedBySaleItemID := make(map[int]int)
	if len(returnParams.ReturnItems) == 0 {
		return httperror.BadRequest("select at least one item to return", nil)
	}

	for _, r := range returnParams.ReturnItems {
		if _, exists := validItems[r.SaleItemID]; !exists {
			return httperror.BadRequest("invalid sale item id passed in", fmt.Errorf("bad sale item id: %d", r.SaleItemID))
		}
		if r.Quantity <= 0 {
			return httperror.BadRequest(
				"return quantity must be > 0",
				fmt.Errorf("invalid return quantity for sale item: %d. quantity: %d", r.SaleItemID, r.Quantity),
			)
		}
		requestedBySaleItemID[r.SaleItemID] += r.Quantity
	}

	for saleItemID, quantity := range requestedBySaleItemID {
		saleItem := validItems[saleItemID]
		if quantity+returnMap[saleItemID] > saleItem.Quantity {
			err = fmt.Errorf(
				"cannot return %d (already returned %d of %d sold)",
				quantity, returnMap[saleItemID], saleItem.Quantity,
			)
			return httperror.BadRequest(err.Error(), err)
		}

		returnItems = append(returnItems, model.ReturnItems{
			SaleItemID: saleItemID,
			Quantity:   quantity,
		})

		stockMovement = append(stockMovement, model.StockMovement{
			ProductID:    saleItem.ProductID,
			Quantity:     quantity * saleItem.StockQuantityPerUnit,
			ReferenceID:  returnParams.SaleID,
			MovementType: model.MovementTypeInSaleReturn,
		})

		previouslyReturned := returnMap[saleItemID]
		if saleItem.Quantity <= 0 {
			return httperror.BadRequest("sale item has an invalid quantity", nil)
		}
		refundBefore := saleItem.TotalPrice * previouslyReturned / saleItem.Quantity
		refundAfter := saleItem.TotalPrice * (previouslyReturned + quantity) / saleItem.Quantity
		refundTotal += refundAfter - refundBefore
	}

	rtn := model.Return{
		SaleID:        returnParams.SaleID,
		CashierID:     returnParams.CashierID,
		TotalRefunded: refundTotal,
		Notes:         returnParams.Notes,
	}

	rID, err := s.repo.CreateReturnTx(ctx, tx, rtn)
	if err != nil {
		log.Println("failed to create return: ", err)
		return httperror.ServerError("failed to create return", err)
	}

	for i := range returnItems {
		returnItems[i].ReturnID = rID
	}

	err = s.repo.BulkCreateReturnItemsTx(ctx, tx, returnItems)
	if err != nil {
		log.Println("failed to bulk create return items: ", err)
		return httperror.ServerError("failed to bulk create return items", err)
	}

	// do inventory update
	err = s.repo.BulkCreateStockMovementTx(ctx, tx, stockMovement)
	if err != nil {
		log.Println(err)
		return httperror.ServerError("Failed to bulk create stock movements", err)
	}

	err = tx.Commit()
	if err != nil {
		log.Println(err)
		return httperror.ServerError("failed to commit transaction", err)
	}
	committed = true
	return nil
}
