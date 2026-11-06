// Package demoseed fills an empty database with made-up data (users, products, a day of sales,
// a purchase order waiting for approval, a stock flag and an adjustment) so the app can be run
// and photographed without real data. It goes through the same services the app uses.
//
// It refuses to run against anything but a database on this machine that has no users yet, so
// it cannot touch a real shop's data by mistake.
package demoseed

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"time"

	"github.com/jmoiron/sqlx"

	"pharmacy/internal/constant"
	"pharmacy/internal/types"
	"pharmacy/repository"
	"pharmacy/service"
)

// Password is the password of every demo user.
const Password = "Demo-Pharmacy-2026"

// Users are the demo accounts, by role.
var Users = []struct{ Name, Role string }{
	{"owner", "Admin"},
	{"amaka", "Manager"},
	{"tunde", "Cashier"},
}

// CheckTarget refuses any database that is not on this machine.
func CheckTarget(dsn string) error {
	u, err := url.Parse(dsn)
	if err != nil || u.Host == "" {
		return fmt.Errorf("cannot read the database address: %q", dsn)
	}
	// pgx lets query parameters such as host= or hostaddr= override the address above, so the
	// only one allowed is sslmode.
	for key := range u.Query() {
		if key != "sslmode" {
			return fmt.Errorf("refusing connection setting %q: only sslmode may be set", key)
		}
	}
	host := u.Hostname()
	if host != "localhost" {
		if ip := net.ParseIP(host); ip == nil || !ip.IsLoopback() {
			return fmt.Errorf("refusing to seed %q: demo data only goes into a local database", host)
		}
	}
	return nil
}

// Summary says what was created.
type Summary struct {
	Products, Sales int
}

type product struct {
	price       float64
	id, priceID int
}

var catalogue = []struct {
	name, maker, category string
	cost, price           float64
	reorder, qty, months  int
}{
	{"Paracetamol 500mg x20", "Emzor", "Pain relief", 350, 600, 20, 120, 20},
	{"Ibuprofen 400mg x10", "Fidson", "Pain relief", 500, 850, 15, 80, 18},
	{"Diclofenac Gel 30g", "Neimeth", "Pain relief", 900, 1500, 10, 36, 14},
	{"Amoxicillin 500mg x21", "May & Baker", "Antibiotics", 1800, 2800, 15, 60, 12},
	{"Azithromycin 500mg x3", "Fidson", "Antibiotics", 1500, 2400, 10, 40, 10},
	{"Ciprofloxacin 500mg x10", "Emzor", "Antibiotics", 1100, 1800, 10, 45, 16},
	{"Metronidazole 400mg x21", "May & Baker", "Antibiotics", 600, 1000, 12, 70, 22},
	{"Vitamin C 1000mg x30", "Swipha", "Vitamins", 1200, 2000, 12, 50, 24},
	{"Multivitamin Adult x30", "Swipha", "Vitamins", 1600, 2600, 10, 42, 24},
	{"Zinc Sulphate 20mg x30", "Neimeth", "Vitamins", 700, 1200, 10, 55, 20},
	{"Folic Acid 5mg x28", "Emzor", "Vitamins", 400, 700, 10, 65, 26},
	{"Loratadine 10mg x10", "Fidson", "Allergy", 450, 800, 12, 48, 15},
	{"Cetirizine 10mg x10", "Emzor", "Allergy", 350, 650, 12, 58, 18},
	{"Oral Rehydration Salts", "Neimeth", "First aid", 150, 300, 25, 90, 28},
	{"Antiseptic Solution 100ml", "Swipha", "First aid", 500, 900, 8, 30, 30},
	{"Plasters (box of 100)", "Hansaplast", "First aid", 800, 1400, 8, 25, 36},
	{"Cotton Wool 100g", "Hansaplast", "First aid", 400, 700, 8, 28, 36},
	{"Hydrocortisone Cream 1%", "Neimeth", "Skin care", 700, 1200, 8, 32, 14},
	{"Clotrimazole Cream 20g", "Fidson", "Skin care", 600, 1000, 8, 34, 16},
	{"Petroleum Jelly 100g", "Vaseline", "Skin care", 500, 900, 8, 40, 30},
	{"Digital Thermometer", "Omron", "Devices", 1800, 3000, 5, 14, 60},
	{"Blood Glucose Strips x50", "Accu-Chek", "Devices", 5500, 8500, 6, 18, 9},
	{"Amlodipine 5mg x30", "Emzor", "Chronic care", 800, 1400, 12, 52, 20},
	{"Metformin 500mg x30", "May & Baker", "Chronic care", 600, 1100, 12, 56, 20},
}

type line struct {
	item, qty int
	discount  float64
}

// demoSale pays its computed total by one method; a cash sale can be paid with a larger note, so
// the slip shows change.
type demoSale struct {
	lines  []line
	method string
	tender float64
}

var sales = []demoSale{
	{[]line{{0, 2, 0}, {13, 3, 0}}, "Cash", 2100},
	{[]line{{3, 1, 0}, {7, 1, 0}}, "Card", 0},
	{[]line{{1, 1, 0}}, "Cash", 1000},
	{[]line{{11, 2, 0}, {12, 1, 0}}, "Transfer", 0},
	{[]line{{21, 1, 0}}, "Card", 0},
	{[]line{{9, 2, 100}, {10, 2, 0}}, "Cash", 0},
	{[]line{{15, 1, 0}, {16, 1, 0}, {14, 1, 0}}, "Cash", 3000},
	{[]line{{22, 2, 0}, {23, 2, 0}}, "Transfer", 0},
	{[]line{{4, 1, 0}}, "Card", 0},
	{[]line{{0, 1, 0}, {1, 1, 0}, {8, 1, 0}}, "Cash", 0},
}

func as(keys ...string) context.Context {
	perms := map[string]bool{}
	for _, k := range keys {
		perms[k] = true
	}
	return context.WithValue(context.Background(), constant.PermissionsSessionKey, perms)
}

// Run seeds db, which must be migrated and have no users.
func Run(db *sqlx.DB) (Summary, error) {
	var users int
	if err := db.Get(&users, `SELECT COUNT(*) FROM users`); err != nil {
		return Summary{}, fmt.Errorf("is the database migrated? %w", err)
	}
	if users > 0 {
		return Summary{}, errors.New("refusing to seed a database that already has users " +
			"(if an earlier run failed part-way, drop and re-create the demo database)")
	}

	store := repository.NewStore(db)
	inv := service.NewInventoryService(store)
	saleSvc := service.NewSaleService(store)
	adminSvc := service.NewAdminService(store)

	// Users
	ids := map[string]int{}
	for _, u := range Users {
		var role int
		if err := db.Get(&role, `SELECT id FROM roles WHERE name = $1`, u.Role); err != nil {
			return Summary{}, fmt.Errorf("role %s: %w", u.Role, err)
		}
		actor := 0 // the first account has no one to be created by
		if u.Name != "owner" {
			actor = ids["owner"]
		}
		id, err := adminSvc.CreateUser(context.Background(),
			types.CreateUserRequest{Username: u.Name, Password: Password, RoleID: role}, actor)
		if err != nil {
			return Summary{}, fmt.Errorf("user %s: %w", u.Name, err)
		}
		ids[u.Name] = id
	}
	manager, cashier := ids["amaka"], ids["tunde"]
	admin := as(constant.AdminAccessPermissionKey, constant.ApprovePurchaseOrderPermissionKey,
		constant.AdjustStockPermissionKey, constant.EditInventoryPermissionKey,
		constant.ViewSalesTotalPermissionKey)
	mgr := as(constant.ApprovePurchaseOrderPermissionKey, constant.AdjustStockPermissionKey,
		constant.EditInventoryPermissionKey, constant.ViewSalesTotalPermissionKey)
	clerk := as(constant.DraftPurchaseOrderPermissionKey, constant.FlagStockPermissionKey)

	// Products and opening stock, received from one supplier.
	categories := map[string]int{}
	products := make([]product, len(catalogue))
	var received []types.ReceiveItem
	for i, c := range catalogue {
		cat, ok := categories[c.category]
		if !ok {
			err := db.Get(&cat, `INSERT INTO category (name) VALUES ($1) RETURNING id`, c.category)
			if err != nil {
				return Summary{}, err
			}
			categories[c.category] = cat
		}
		barcode := fmt.Sprintf("600100%05d", i+1)
		res, err := inv.CreateProduct(admin, types.CreateProductRequest{
			Name: c.name, Barcode: barcode, Manufacturer: c.maker, CategoryID: cat,
			ReorderLevel: c.reorder, CostPrice: c.cost, SellingPrice: c.price,
		})
		if err != nil {
			return Summary{}, fmt.Errorf("product %s: %w", c.name, err)
		}
		products[i] = product{id: res.ID, price: c.price}
		received = append(received, types.ReceiveItem{
			ID: res.ID, Barcode: barcode, CostPrice: c.cost, SellingPrice: c.price,
			Quantity: c.qty, Expiry: time.Now().AddDate(0, c.months, 0),
		})
		if err := db.Get(&products[i].priceID,
			`SELECT id FROM product_price WHERE product_id = $1 ORDER BY id LIMIT 1`, res.ID); err != nil {
			return Summary{}, err
		}
	}
	if err := inv.ReceiveProductSupply(mgr, types.ReceiveSupplyRequest{
		Supplier: "MedSupply Ltd", UserID: manager, IdempotencyKey: "demo-opening-stock",
		Products: received,
	}); err != nil {
		return Summary{}, fmt.Errorf("opening stock: %w", err)
	}

	// A day of sales by the cashier.
	for n, s := range sales {
		sale := types.Sale{CashierID: cashier, IdempotencyKey: fmt.Sprintf("demo-sale-%d", n+1)}
		for _, l := range s.lines {
			p := products[l.item]
			total := p.price*float64(l.qty) - l.discount
			sale.Items = append(sale.Items, types.SaleItem{
				ProductID: p.id, Quantity: l.qty, PriceID: p.priceID, UnitPrice: p.price,
				Discount: l.discount, Total: total,
			})
			sale.Subtotal += p.price * float64(l.qty)
			sale.Discount += l.discount
			sale.Total += total
		}
		paid := sale.Total
		if s.tender > paid {
			paid = s.tender
		}
		sale.Payments = []types.SalePayment{{Amount: paid, PaymentMethod: s.method}}
		if _, err := saleSvc.CreateSale(mgr, sale); err != nil {
			return Summary{}, fmt.Errorf("sale %d: %w", n+1, err)
		}
	}

	// One item from the first sale comes back.
	var saleID, itemID int
	if err := db.QueryRow(`
		SELECT s.id, i.id FROM sales s JOIN sales_item i ON i.sale_id = s.id
		ORDER BY s.id, i.id LIMIT 1`).Scan(&saleID, &itemID); err != nil {
		return Summary{}, err
	}
	if err := saleSvc.ReturnItems(mgr, types.ReturnSale{
		SaleID: saleID, CashierID: manager, Notes: "Customer changed their mind",
		ReturnItems: []types.ReturnSaleItem{{SaleItemID: itemID, Quantity: 1}},
	}); err != nil {
		return Summary{}, fmt.Errorf("return: %w", err)
	}

	// Purchase orders: one the cashier submitted, one the manager placed directly.
	low := func(i int) types.PurchaseOrderItem {
		id := products[i].id
		return types.PurchaseOrderItem{
			ProductID: &id, Name: catalogue[i].name, Manufacturer: catalogue[i].maker, Quantity: 40,
		}
	}
	draftID, err := inv.SavePurchaseOrder(clerk, types.PurchaseOrderRequest{
		Supplier: "MedSupply Ltd", Note: "Running low on antibiotics",
		Items: []types.PurchaseOrderItem{low(3), low(4), low(6)},
	}, cashier)
	if err != nil {
		return Summary{}, fmt.Errorf("draft order: %w", err)
	}
	draft, err := inv.GetPurchaseOrder(clerk, draftID, cashier)
	if err != nil {
		return Summary{}, err
	}
	if err := inv.SubmitPurchaseOrder(clerk, draftID, draft.Version, cashier); err != nil {
		return Summary{}, fmt.Errorf("submit order: %w", err)
	}
	if _, err := inv.SavePurchaseOrder(mgr, types.PurchaseOrderRequest{
		Supplier: "HealthPlus Distributors", Note: "Weekly vitamins",
		Items: []types.PurchaseOrderItem{low(7), low(8), low(9)},
	}, manager); err != nil {
		return Summary{}, fmt.Errorf("order: %w", err)
	}

	// The cashier flags a shelf that looks short; the manager writes off damaged stock.
	counted := 9
	if _, err := inv.FlagStock(clerk, types.StockFlagRequest{
		ProductID: products[2].id, Note: "Shelf looks short", CountedQuantity: &counted,
	}, cashier); err != nil {
		return Summary{}, fmt.Errorf("flag: %w", err)
	}
	if _, err := inv.CreateStockAdjustment(mgr, types.StockAdjustmentRequest{
		Reason: "damaged", Note: "Dropped carton", IdempotencyKey: "demo-adjustment-1",
		Items: []types.StockAdjustmentItemRequest{
			{ProductID: products[15].id, Mode: "remove", Quantity: 3},
		},
	}, manager); err != nil {
		return Summary{}, fmt.Errorf("adjustment: %w", err)
	}

	return Summary{Products: len(products), Sales: len(sales)}, nil
}
