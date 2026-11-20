package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"

	"pharmacy/httperror"
	"pharmacy/internal/types"
	"pharmacy/model"
	"pharmacy/repository"
)

func TestPriceKobo(t *testing.T) {
	accepted := map[float64]int{
		0: 0, 0.29: 29, 19.99: 1999, 1.15: 115, 0.07: 7,
		86.415:      8642, // a receiving pre-fill (70% of ₦123.45) rounds to the nearest kobo
		0.004:       0,
		1999999.99:  199999999,
		21474836.47: math.MaxInt32,
	}
	for naira, want := range accepted {
		if got, err := priceKobo("cost price", naira); err != nil || got != want {
			t.Errorf("priceKobo(%v) = %d, %v; want %d", naira, got, err, want)
		}
	}
	refused := map[float64]string{
		-0.01: "cost price cannot be negative", -5: "cost price cannot be negative",
		math.Inf(-1): "cost price cannot be negative",
		math.NaN():   "cost price is not a number",
		math.Inf(1):  "cost price is too large (the most is ₦21,474,836.47)",
		1e12:         "cost price is too large (the most is ₦21,474,836.47)",
		21474836.48:  "cost price is too large (the most is ₦21,474,836.47)",
	}
	for naira, want := range refused {
		_, err := priceKobo("cost price", naira)
		var refusal *httperror.HTTPError
		if !errors.As(err, &refusal) || refusal.Code != http.StatusBadRequest ||
			refusal.Message != want {
			t.Errorf("priceKobo(%v) = %v; want a 400 saying %q", naira, err, want)
		}
	}
}

func wantRefusedBeforeWriting(t *testing.T, err error, repo *priceRecordingRepo, msg string) {
	t.Helper()
	requireHTTPStatus(t, err, http.StatusBadRequest)
	var refusal *httperror.HTTPError
	if errors.As(err, &refusal) && refusal.Message != msg {
		t.Errorf("refusal says %q; want %q", refusal.Message, msg)
	}
	if repo.begun != 0 {
		t.Errorf("%d transactions begun; want the refusal before any", repo.begun)
	}
}

// priceRecordingRepo records the kobo each inventory write is given.
type priceRecordingRepo struct {
	repository.PharmacyRepository

	begun          int
	products       []model.Product
	createdPrices  []model.ProductPrice
	updatedPrices  []model.ProductPrice
	priceOptions   []model.ProductPrice // UpdateProductPriceByIDTx while receiving
	bulkPriceMaps  []map[string]any
	receivedBatchs []model.ProductBatch
}

var errStopHere = errors.New("test stops here")

func (r *priceRecordingRepo) BeginTx(context.Context) (*sqlx.Tx, error) {
	r.begun++
	return nil, nil
}
func (r *priceRecordingRepo) RollbackTx(*sqlx.Tx)     {}
func (r *priceRecordingRepo) CommitTx(*sqlx.Tx) error { return nil }

func (r *priceRecordingRepo) CreateProductTx(
	_ context.Context, _ *sqlx.Tx, p model.Product,
) (int, error) {
	r.products = append(r.products, p)
	return 7, nil
}

func (r *priceRecordingRepo) UpdateProductTx(_ context.Context, _ *sqlx.Tx, p model.Product) error {
	r.products = append(r.products, p)
	return nil
}

func (r *priceRecordingRepo) CreateProductPriceTx(
	_ context.Context, _ *sqlx.Tx, p model.ProductPrice,
) (int, error) {
	r.createdPrices = append(r.createdPrices, p)
	return 11, nil
}

func (r *priceRecordingRepo) UpdateProductPriceTx(
	_ context.Context, _ *sqlx.Tx, p model.ProductPrice,
) error {
	r.updatedPrices = append(r.updatedPrices, p)
	return nil
}

func (r *priceRecordingRepo) UpdateProductPriceByIDTx(
	_ context.Context, _ *sqlx.Tx, _ int, p model.ProductPrice,
) error {
	r.priceOptions = append(r.priceOptions, p)
	return nil
}

func (r *priceRecordingRepo) UpdateProductDefaultPriceTx(
	context.Context, *sqlx.Tx, int, int,
) error {
	return nil
}

func (r *priceRecordingRepo) FetchProductByIDWithPrices(
	context.Context,
	int,
) (model.Product, error) {
	return model.Product{}, nil
}

func (r *priceRecordingRepo) FindOrCreateSupplierTx(
	_ context.Context, _ *sqlx.Tx, name string,
) (types.Supplier, error) {
	return types.Supplier{ID: 1, Name: name}, nil
}

func (r *priceRecordingRepo) CreateReceivingBatchTx(
	context.Context, *sqlx.Tx, model.ReceivingBatch,
) (int, error) {
	return 3, nil
}

func (r *priceRecordingRepo) FetchDefaultPriceID(context.Context, int) (int, error) {
	return 5, nil
}

func (r *priceRecordingRepo) BulkUpdateProductPricesTx(
	_ context.Context, _ *sqlx.Tx, values []map[string]any,
) error {
	r.bulkPriceMaps = append(r.bulkPriceMaps, values...)
	return nil
}

// BulkCreateProductBatchTx ends the receipt: everything priced has been recorded by now.
func (r *priceRecordingRepo) BulkCreateProductBatchTx(
	_ context.Context, _ *sqlx.Tx, batches []model.ProductBatch,
) ([]types.BatchInsertReturn, error) {
	r.receivedBatchs = append(r.receivedBatchs, batches...)
	return nil, errStopHere
}

func TestCreateProductStoresTheKoboTyped(t *testing.T) {
	repo := &priceRecordingRepo{}
	_, err := NewInventoryService(
		repo,
	).CreateProduct(context.Background(), types.CreateProductRequest{
		Name: "Paracetamol", CategoryID: 1, CostPrice: 0.29, SellingPrice: 19.99,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := repo.products[0].CostPriceKobo; got != 29 {
		t.Errorf("cost = %d kobo; want 29", got)
	}
	if got := repo.createdPrices[0].SellingPriceKobo; got != 1999 {
		t.Errorf("selling price = %d kobo; want 1999", got)
	}
}

func TestCreateProductRefusesAPriceTooBigToStore(t *testing.T) {
	repo := &priceRecordingRepo{}
	_, err := NewInventoryService(
		repo,
	).CreateProduct(context.Background(), types.CreateProductRequest{
		Name: "Paracetamol", CategoryID: 1, CostPrice: 1, SellingPrice: 1e12,
	})
	wantRefusedBeforeWriting(
		t,
		err,
		repo,
		"selling price is too large (the most is ₦21,474,836.47)",
	)
}

// A price above zero that rounds to 0 kobo would sell the product for nothing.
func TestCreateProductRefusesAPriceThatRoundsToZero(t *testing.T) {
	cases := map[string]types.CreateProductRequest{
		"cost price must be at least ₦0.01": {
			Name: "Paracetamol", CategoryID: 1, CostPrice: 0.004, SellingPrice: 1,
		},
		"selling price must be at least ₦0.01": {
			Name: "Paracetamol", CategoryID: 1, CostPrice: 1, SellingPrice: 0.004,
		},
	}
	for msg, req := range cases {
		t.Run(msg, func(t *testing.T) {
			repo := &priceRecordingRepo{}
			_, err := NewInventoryService(repo).CreateProduct(context.Background(), req)
			wantRefusedBeforeWriting(t, err, repo, msg)
		})
	}
}

func receiveRequest(options ...types.PriceOptionChange) types.ReceiveSupplyRequest {
	return types.ReceiveSupplyRequest{
		Supplier: "Emzor",
		Products: []types.ReceiveItem{{
			ID: 2, CostPrice: 19.99, SellingPrice: 0.29, Quantity: 1,
			Expiry: time.Now().AddDate(1, 0, 0), PriceOptionsChanges: options,
		}},
	}
}

func TestReceivingStoresTheKoboTyped(t *testing.T) {
	repo := &priceRecordingRepo{}
	optionID := 9
	err := NewInventoryService(repo).ReceiveProductSupply(context.Background(), receiveRequest(
		types.PriceOptionChange{Name: "Card", SellingPrice: 1.15, QuantityPerUnit: 10},
		types.PriceOptionChange{
			ID:              &optionID,
			Name:            "Pack",
			SellingPrice:    86.415,
			QuantityPerUnit: 5,
		},
	))
	if err == nil || len(repo.receivedBatchs) == 0 {
		t.Fatalf("err = %v; want the test's stop at product batches", err)
	}
	if got := repo.receivedBatchs[0].CostPriceKobo; got != 1999 {
		t.Errorf("batch cost = %d kobo; want 1999", got)
	}
	if got := repo.bulkPriceMaps[0]; got["cost_price"] != 1999 || got["selling_price"] != 29 {
		t.Errorf("product prices = %v; want cost 1999, selling 29", got)
	}
	if got := repo.createdPrices[0].SellingPriceKobo; got != 115 {
		t.Errorf("new option = %d kobo; want 115", got)
	}
	if got := repo.priceOptions[0].SellingPriceKobo; got != 8642 {
		t.Errorf("edited option = %d kobo; want 8642", got)
	}
}

func TestReceivingRefusesANegativePriceOption(t *testing.T) {
	repo := &priceRecordingRepo{}
	err := NewInventoryService(repo).ReceiveProductSupply(context.Background(), receiveRequest(
		types.PriceOptionChange{Name: "Card", SellingPrice: -1, QuantityPerUnit: 10},
	))
	wantRefusedBeforeWriting(t, err, repo, `product 1: price of "Card" cannot be negative`)
}

func TestReceivingRefusesAPriceThatRoundsToZero(t *testing.T) {
	cases := map[string][2]float64{ // cost, selling
		"product 1: cost price must be at least ₦0.01":    {0.001, 1},
		"product 1: selling price must be at least ₦0.01": {1, 0.004},
	}
	for msg, prices := range cases {
		t.Run(msg, func(t *testing.T) {
			repo := &priceRecordingRepo{}
			req := receiveRequest()
			req.Products[0].CostPrice, req.Products[0].SellingPrice = prices[0], prices[1]
			err := NewInventoryService(repo).ReceiveProductSupply(context.Background(), req)
			wantRefusedBeforeWriting(t, err, repo, msg)
		})
	}
}

func TestUpdateProductStoresTheKoboTyped(t *testing.T) {
	repo := &priceRecordingRepo{}
	priceID := 5
	err := NewInventoryService(
		repo,
	).UpdateProduct(context.Background(), 2, types.UpdateProductRequest{
		Name: "Paracetamol", CategoryID: 1, CostPrice: 0.29,
		Prices: []types.ProductPriceUpdate{
			{ID: &priceID, Name: "Unit", SellingPrice: 19.99, QuantityPerUnit: 1, IsDefault: true},
			{Name: "Card", SellingPrice: 1.15, QuantityPerUnit: 10},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := repo.products[0].CostPriceKobo; got != 29 {
		t.Errorf("cost = %d kobo; want 29", got)
	}
	if got := repo.updatedPrices[0].SellingPriceKobo; got != 1999 {
		t.Errorf("edited price = %d kobo; want 1999", got)
	}
	if got := repo.createdPrices[0].SellingPriceKobo; got != 115 {
		t.Errorf("new price = %d kobo; want 115", got)
	}
}

func TestUpdateProductRefusesABadPrice(t *testing.T) {
	for msg, req := range map[string]types.UpdateProductRequest{
		"cost price cannot be negative": {Name: "Paracetamol", CostPrice: -1},
		`price of "Unit" is too large (the most is ₦21,474,836.47)`: {
			Name: "Paracetamol", CostPrice: 1, Prices: []types.ProductPriceUpdate{
				{Name: "Unit", SellingPrice: 1e12, QuantityPerUnit: 1},
			},
		},
	} {
		t.Run(msg, func(t *testing.T) {
			repo := &priceRecordingRepo{}
			err := NewInventoryService(repo).UpdateProduct(context.Background(), 2, req)
			wantRefusedBeforeWriting(t, err, repo, msg)
		})
	}
}

func TestAnUnnamedPriceOptionIsNamedInTheRefusal(t *testing.T) {
	repo := &priceRecordingRepo{}
	err := NewInventoryService(repo).ReceiveProductSupply(context.Background(), receiveRequest(
		types.PriceOptionChange{Name: "  ", SellingPrice: -1, QuantityPerUnit: 10},
	))
	wantRefusedBeforeWriting(t, err, repo, "product 1: price option 1 cannot be negative")

	repo = &priceRecordingRepo{}
	err = NewInventoryService(repo).UpdateProduct(context.Background(), 2,
		types.UpdateProductRequest{
			Name:      "Paracetamol",
			CostPrice: 1,
			Prices: []types.ProductPriceUpdate{
				{Name: "Unit", SellingPrice: 1, QuantityPerUnit: 1},
				{Name: "", SellingPrice: -1, QuantityPerUnit: 10},
			},
		})
	wantRefusedBeforeWriting(t, err, repo, "price option 2 cannot be negative")
}

// No path stores a price of ₦0: a product at ₦0 sells for nothing, and a cost of ₦0 hides the
// product's value from stock reports.
func TestReceivingRefusesAPriceOptionUnderAKobo(t *testing.T) {
	for _, price := range []float64{0, 0.004} {
		t.Run(fmt.Sprint(price), func(t *testing.T) {
			repo := &priceRecordingRepo{}
			err := NewInventoryService(repo).ReceiveProductSupply(context.Background(),
				receiveRequest(
					types.PriceOptionChange{Name: "Card", SellingPrice: price, QuantityPerUnit: 10},
				))
			wantRefusedBeforeWriting(
				t,
				err,
				repo,
				`product 1: price of "Card" must be at least ₦0.01`,
			)
		})
	}
}

func TestUpdateProductRefusesAPriceUnderAKobo(t *testing.T) {
	cases := []struct {
		cost, price float64
		msg         string
	}{
		{0, 1, "cost price must be at least ₦0.01"},
		{1, 0, `price of "Unit" must be at least ₦0.01`},
		{1, 0.004, `price of "Unit" must be at least ₦0.01`},
	}
	for _, c := range cases {
		t.Run(fmt.Sprint(c.cost, "/", c.price), func(t *testing.T) {
			repo := &priceRecordingRepo{}
			err := NewInventoryService(
				repo,
			).UpdateProduct(context.Background(), 2, types.UpdateProductRequest{
				Name:      "Paracetamol",
				CostPrice: c.cost,
				Prices: []types.ProductPriceUpdate{
					{Name: "Unit", SellingPrice: c.price, QuantityPerUnit: 1},
				},
			})
			wantRefusedBeforeWriting(t, err, repo, c.msg)
		})
	}
}
