package service

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"net/http"
	"strings"
	"testing"

	"pharmacy/httperror"
	"pharmacy/internal/types"
	"pharmacy/model"

	"github.com/jmoiron/sqlx"
)

// pricingRepo is a till's repository with a price list: it keeps the lines a sale saves.
type pricingRepo struct {
	customerSaleRepo

	prices map[int]model.ProductPrice
	saved  []model.SaleItem
	pays   []model.SalePayment
}

// Product 1 is sold singly at ₦100.00 (price 1) and in a pack of 10 at ₦900.00 (price 2).
// Product 2 costs ₦19.99 (price 3).
func newPricingRepo() *pricingRepo {
	r := &pricingRepo{prices: map[int]model.ProductPrice{}}
	add := func(id, product, perUnit, kobo int) {
		p := model.ProductPrice{
			ProductID:        product,
			QuantityPerUnit:  perUnit,
			SellingPriceKobo: kobo,
		}
		p.ID = id
		r.prices[id] = p
	}
	add(1, 1, 1, 10000)
	add(2, 1, 10, 90000)
	add(3, 2, 1, 1999)
	return r
}

func (r *pricingRepo) FetchPriceByID(_ context.Context, id int) (model.ProductPrice, error) {
	p, ok := r.prices[id]
	if !ok {
		return model.ProductPrice{}, sql.ErrNoRows
	}
	return p, nil
}

func (r *pricingRepo) BulkFetchProductByIDsTx(
	_ context.Context, _ *sqlx.Tx, ids []int,
) ([]model.Product, error) {
	var out []model.Product
	for _, id := range ids {
		p := model.Product{Name: "Paracetamol 500mg"}
		p.ID = id
		out = append(out, p)
	}
	return out, nil
}

func (r *pricingRepo) BulkCreateSaleItemsTx(
	_ context.Context, _ *sqlx.Tx, items []model.SaleItem,
) error {
	r.saved = append(r.saved, items...)
	return nil
}

func (r *pricingRepo) BulkCreateSalePaymentsTX(
	_ context.Context, _ *sqlx.Tx, pays []model.SalePayment,
) error {
	r.pays = append(r.pays, pays...)
	return nil
}

// honestSale is what the till sends for 2 × price 1 with ₦50 off: ₦200 - ₦50 = ₦150.
func honestSale() types.Sale {
	return types.Sale{
		CashierID:      1,
		IdempotencyKey: "price-key-1",
		Subtotal:       200,
		Discount:       50,
		Total:          150,
		Items: []types.SaleItem{
			{ProductID: 1, PriceID: 1, Quantity: 2, UnitPrice: 100, Discount: 50, Total: 150},
		},
		Payments: []types.SalePayment{{Amount: 150, PaymentMethod: "Cash"}},
	}
}

func createSale(t *testing.T, repo *pricingRepo, sale types.Sale) error {
	t.Helper()
	_, err := NewSaleService(repo).CreateSale(contextWith(), sale)
	return err
}

func wantRefusal(t *testing.T, err error, status int, code string) *httperror.HTTPError {
	t.Helper()
	got, ok := err.(*httperror.HTTPError)
	if !ok {
		t.Fatalf("want an HTTP refusal %d, got %v", status, err)
	}
	if got.Code != status || got.Reason != code {
		t.Fatalf("want %d %q, got %d %q (%v)", status, code, got.Code, got.Reason, err)
	}
	return got
}

// wantNothingSaved checks a refused sale left no trace: not even the sale row.
func wantNothingSaved(t *testing.T, repo *pricingRepo) {
	t.Helper()
	if len(repo.sales) != 0 || len(repo.saved) != 0 || repo.commits != 0 {
		t.Fatalf("a refused sale saved %d sales, %d lines, %d commits",
			len(repo.sales), len(repo.saved), repo.commits)
	}
}

func TestAnHonestSaleSavesTheServersNumbers(t *testing.T) {
	repo := newPricingRepo()
	if err := createSale(t, repo, honestSale()); err != nil {
		t.Fatal(err)
	}
	sale := repo.sales[0]
	if sale.Subtotal != 20000 || sale.Discount != 5000 || sale.Total != 15000 {
		t.Fatalf("sale numbers in kobo: %+v", sale)
	}
	line := repo.saved[0]
	if line.UnitPrice != 10000 || line.Discount != 5000 || line.TotalPrice != 15000 {
		t.Fatalf("line numbers in kobo: %+v", line)
	}
	if len(repo.pays) != 1 || repo.pays[0].Amount != 15000 {
		t.Fatalf("payments: %+v", repo.pays)
	}
}

func TestACheaperPriceThanOnFileIsRefusedAsPriceChanged(t *testing.T) {
	repo := newPricingRepo()
	sale := honestSale()
	sale.Subtotal, sale.Discount, sale.Total = 2, 0, 2
	sale.Items = []types.SaleItem{
		{ProductID: 1, PriceID: 1, Quantity: 2, UnitPrice: 1, Total: 2},
	}
	err := createSale(t, repo, sale)
	got := wantRefusal(t, err, http.StatusConflict, "price_changed")
	if want := "Paracetamol 500mg"; !strings.Contains(got.Message, want) {
		t.Fatalf("the refusal should name the product: %q", got.Message)
	}
	wantNothingSaved(t, repo)
}

func TestTotalsThatDoNotAddUpAreRefused(t *testing.T) {
	cases := map[string]func(*types.Sale){
		"total":    func(s *types.Sale) { s.Total = 1 },
		"subtotal": func(s *types.Sale) { s.Subtotal = 1 },
		"discount": func(s *types.Sale) { s.Discount = 0 },
		"line":     func(s *types.Sale) { s.Items[0].Total = 1 },
	}
	for name, tamper := range cases {
		t.Run(name, func(t *testing.T) {
			repo := newPricingRepo()
			sale := honestSale()
			tamper(&sale)
			wantRefusal(t, createSale(t, repo, sale), http.StatusBadRequest, "")
			wantNothingSaved(t, repo)
		})
	}
}

func TestAPriceOfAnotherProductIsRefused(t *testing.T) {
	repo := newPricingRepo()
	sale := honestSale()
	sale.Items[0].ProductID = 2 // price 1 belongs to product 1
	wantRefusal(t, createSale(t, repo, sale), http.StatusBadRequest, "")
	wantNothingSaved(t, repo)
}

// Product edit can delete a price, so a cart may hold one that is gone: the till is told what it
// is told for a changed price, which makes it read the product list again.
func TestAPriceThatIsGoneIsRefusedAsPriceChangedNotAServerError(t *testing.T) {
	repo := newPricingRepo()
	sale := honestSale()
	sale.Items[0].PriceID = 99
	wantRefusal(t, createSale(t, repo, sale), http.StatusConflict, "price_changed")
	wantNothingSaved(t, repo)
}

// A price id that cannot exist is a broken request, not a price that changed: no remedy to offer.
func TestAPriceIdThatCannotExistIsABadRequest(t *testing.T) {
	for _, id := range []int{0, -1} {
		repo := newPricingRepo()
		sale := honestSale()
		sale.Items[0].PriceID = id
		wantRefusal(t, createSale(t, repo, sale), http.StatusBadRequest, "")
		wantNothingSaved(t, repo)
	}
}

func TestAPriceThatTakesNoStockOrAddsStockIsRefused(t *testing.T) {
	for name, perUnit := range map[string]int{"zero": 0, "negative": -5} {
		t.Run(name, func(t *testing.T) {
			repo := newPricingRepo()
			p := repo.prices[1]
			p.QuantityPerUnit = perUnit
			repo.prices[1] = p
			wantRefusal(t, createSale(t, repo, honestSale()), http.StatusBadRequest, "")
			wantNothingSaved(t, repo)
		})
	}
}

func TestASaleNeedsItemsWithAPositiveWholeQuantity(t *testing.T) {
	for name, items := range map[string][]types.SaleItem{
		"no items": nil,
		"zero":     {{ProductID: 1, PriceID: 1, Quantity: 0, UnitPrice: 100}},
		"negative": {{ProductID: 1, PriceID: 1, Quantity: -3, UnitPrice: 100, Total: -300}},
		"huge":     {{ProductID: 1, PriceID: 1, Quantity: math.MaxInt32, UnitPrice: 100}},
	} {
		t.Run(name, func(t *testing.T) {
			repo := newPricingRepo()
			sale := honestSale()
			sale.Items = items
			wantRefusal(t, createSale(t, repo, sale), http.StatusBadRequest, "")
			wantNothingSaved(t, repo)
		})
	}
}

func TestADiscountMustBeBetweenNothingAndTheLine(t *testing.T) {
	for name, discount := range map[string]float64{
		"negative":   -10,
		"over":       201,
		"not number": math.NaN(),
		"infinite":   math.Inf(1),
	} {
		t.Run(name, func(t *testing.T) {
			repo := newPricingRepo()
			sale := honestSale()
			sale.Items[0].Discount = discount
			sale.Discount = discount
			wantRefusal(t, createSale(t, repo, sale), http.StatusBadRequest, "")
			wantNothingSaved(t, repo)
		})
	}
}

func TestAWholeLineDiscountIsAllowed(t *testing.T) {
	repo := newPricingRepo()
	sale := honestSale()
	sale.Discount, sale.Total = 200, 0
	sale.Items[0].Discount, sale.Items[0].Total = 200, 0
	sale.Payments = nil
	if err := createSale(t, repo, sale); err != nil {
		t.Fatal(err)
	}
	if repo.sales[0].Total != 0 || repo.saved[0].TotalPrice != 0 {
		t.Fatalf("a free line stores nothing owed: %+v %+v", repo.sales[0], repo.saved[0])
	}
}

func TestKoboAreRoundedNotTruncated(t *testing.T) {
	repo := newPricingRepo()
	sale := types.Sale{
		CashierID:      1,
		IdempotencyKey: "price-key-2",
		Subtotal:       59.97, // 3 × ₦19.99, which floats read as 5996.999…
		Total:          59.97,
		Items: []types.SaleItem{
			{ProductID: 2, PriceID: 3, Quantity: 3, UnitPrice: 19.99, Total: 59.97},
		},
		Payments: []types.SalePayment{{Amount: 59.97, PaymentMethod: "Cash"}},
	}
	if err := createSale(t, repo, sale); err != nil {
		t.Fatal(err)
	}
	if repo.saved[0].UnitPrice != 1999 || repo.saved[0].TotalPrice != 5997 ||
		repo.sales[0].Total != 5997 || repo.pays[0].Amount != 5997 {
		t.Fatalf("kobo: line %+v sale %+v pay %+v", repo.saved[0], repo.sales[0], repo.pays[0])
	}
}

func TestAPackSellsAtItsOwnPriceAndTakesItsOwnStock(t *testing.T) {
	repo := newPricingRepo()
	sale := honestSale()
	sale.Subtotal, sale.Discount, sale.Total = 900, 0, 900
	sale.Items = []types.SaleItem{
		{ProductID: 1, PriceID: 2, Quantity: 1, UnitPrice: 900, Total: 900},
	}
	sale.Payments = []types.SalePayment{{Amount: 900, PaymentMethod: "Cash"}}
	if err := createSale(t, repo, sale); err != nil {
		t.Fatal(err)
	}
	if got := repo.saved[0]; got.UnitPrice != 90000 || got.StockQuantityPerUnit != 10 {
		t.Fatalf("a pack line: %+v", got)
	}
}

func TestANegativePaymentIsRefused(t *testing.T) {
	for name, amount := range map[string]float64{
		"negative": -50, "not number": math.NaN(), "infinite": math.Inf(1),
	} {
		t.Run(name, func(t *testing.T) {
			repo := newPricingRepo()
			sale := honestSale()
			sale.Payments = []types.SalePayment{{Amount: amount, PaymentMethod: "Cash"}}
			wantRefusal(t, createSale(t, repo, sale), http.StatusBadRequest, "")
			wantNothingSaved(t, repo)
		})
	}
}

func TestARetryOfASavedSaleIsNotPricedAgain(t *testing.T) {
	repo := newPricingRepo()
	repo.duplicate = true
	p := repo.prices[1] // the price moved since the first save
	p.SellingPriceKobo = 20000
	repo.prices[1] = p
	receipt, err := NewSaleService(repo).CreateSale(contextWith(), honestSale())
	if err != nil || receipt != "2490-7777" {
		t.Fatalf("a retry is answered with the first save: %q, %v", receipt, err)
	}
	if len(repo.saved) != 0 {
		t.Fatalf("a retry writes no lines: %+v", repo.saved)
	}
}

func TestAFractionOfAKoboIsRefusedWithAClearMessage(t *testing.T) {
	repo := newPricingRepo()
	sale := honestSale()
	sale.Discount, sale.Total = 10.555, 189.445
	sale.Items[0].Discount, sale.Items[0].Total = 10.555, 189.445
	err := createSale(t, repo, sale)
	got := wantRefusal(t, err, http.StatusBadRequest, "")
	if !strings.Contains(got.Message, "kobo") {
		t.Fatalf("say what is wrong with the amount: %q", got.Message)
	}
	wantNothingSaved(t, repo)
}

func TestToKobo(t *testing.T) {
	accepted := map[float64]int{
		0: 0, 19.99: 1999, 0.07: 7, -0.5: -50,
		3.4499999999999997: 345, // float error from 1.15 × 3
		0.1 + 0.2:          30,  // 0.30000000000000004
		21474836.47:        math.MaxInt32,
		21474836.470000003: math.MaxInt32, // a big sum with float error is still that sum
	}
	for naira, want := range accepted {
		if got, err := toKobo(naira); err != nil || got != want {
			t.Errorf("toKobo(%v) = %d, %v; want %d", naira, got, err, want)
		}
	}
	// A till's float sum of many lines drifts further from its kobo the bigger it gets.
	sum := 0.0
	for i := 0; i < 1000; i++ {
		sum += 19999.99
	}
	if got, err := toKobo(sum); err != nil || got != 1999999000 {
		t.Errorf("toKobo(%v) = %d, %v; want 1999999000", sum, got, err)
	}
	refused := map[float64]string{
		math.NaN(): "not valid", math.Inf(1): "not valid", math.Inf(-1): "not valid",
		1e12: "not valid", 21474836.48: "not valid",
		10.555: "fraction of a kobo", 0.004: "fraction of a kobo", 0.0049999: "fraction of a kobo",
	}
	for naira, why := range refused {
		_, err := toKobo(naira)
		var refusal *httperror.HTTPError
		if !errors.As(err, &refusal) || refusal.Code != http.StatusBadRequest ||
			!strings.Contains(refusal.Message, why) {
			t.Errorf("toKobo(%v) = %v; want a 400 saying %q", naira, err, why)
		}
	}
}

func TestAMovementTooBigForTheDatabaseIsABadRequest(t *testing.T) {
	repo := newPricingRepo()
	p := repo.prices[3]
	p.QuantityPerUnit = 2200 // 1,000,000 × 2200 units of stock does not fit an INTEGER column
	repo.prices[3] = p
	sale := types.Sale{
		CashierID: 1, IdempotencyKey: "big", Subtotal: 19990000, Total: 19990000,
		Items: []types.SaleItem{
			{ProductID: 2, PriceID: 3, Quantity: 1_000_000, UnitPrice: 19.99, Total: 19990000},
		},
		Payments: []types.SalePayment{{Amount: 19990000, PaymentMethod: "Cash"}},
	}
	got := wantRefusal(t, createSale(t, repo, sale), http.StatusBadRequest, "")
	if !strings.Contains(got.Message, "too large") {
		t.Fatalf("message: %q", got.Message)
	}
	wantNothingSaved(t, repo)
}

// raceRepo says the key is free when the sale starts checking it, and taken afterwards: the first
// request saved the sale while this one was pricing it.
type raceRepo struct {
	*pricingRepo
	lookups int
}

func (r *raceRepo) FetchSaleByKeyTx(
	context.Context, *sqlx.Tx, string,
) (types.SavedSale, error) {
	r.lookups++
	if r.lookups == 1 {
		return types.SavedSale{}, sql.ErrNoRows
	}
	return types.SavedSale{ReceiptNumber: "2490-7777"}, nil
}

func TestARetryThatMeetsAPriceEditWhileTheFirstSaveCommitsIsAnsweredAsSaved(t *testing.T) {
	inner := newPricingRepo()
	inner.duplicate = true // the insert finds the key taken
	p := inner.prices[1]
	p.SellingPriceKobo = 20000 // the price was edited before this request read it
	inner.prices[1] = p
	repo := &raceRepo{pricingRepo: inner}

	receipt, err := NewSaleService(repo).CreateSale(contextWith(), honestSale())
	if err != nil || receipt != "2490-7777" {
		t.Fatalf("the sale was saved, so say so: %q, %v", receipt, err)
	}
	if len(inner.saved) != 0 {
		t.Fatalf("a retry writes no lines: %+v", inner.saved)
	}
}
