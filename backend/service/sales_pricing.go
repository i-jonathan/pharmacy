package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math"

	"pharmacy/httperror"
	"pharmacy/internal/constant"
	"pharmacy/internal/types"
	"pharmacy/model"

	"github.com/jmoiron/sqlx"
)

// maxSaleQuantity keeps a line's price × quantity far from the database's integer limit.
const maxSaleQuantity = 1_000_000

// pricedSale is a sale as the server works it out from the price list, in kobo. The numbers the
// till sent are only a check against it.
type pricedSale struct {
	items    []model.SaleItem // SaleID is set once the sale row exists
	payments []model.SalePayment
	subtotal int
	discount int
	total    int
}

// A float of naira × 100 sits a little off the whole kobo it stands for: 19.99 × 100 is
// 1998.9999999999998, and the error of a till's summed total grows with its size and its lines.
// minFloatDrift is the least that is allowed; relFloatDrift scales it up for big amounts. Both
// are far below half a kobo, so a real fraction of a kobo (0.4 of one, say) is still refused.
const (
	minFloatDrift = 1e-6
	relFloatDrift = 1e-12
)

// toKobo turns naira sent by the till into whole kobo, rounding rather than truncating (19.99 is
// 1998.999… as a float). A number that is not finite, has a fraction of a kobo, or is too big for
// the database is refused.
func toKobo(naira float64) (int, error) {
	scaled := naira * 100
	kobo := math.Round(scaled)
	if math.IsNaN(kobo) || math.IsInf(kobo, 0) || math.Abs(kobo) > math.MaxInt32 {
		return 0, badRequest("an amount in the sale is not valid")
	}
	if math.Abs(scaled-kobo) > math.Max(minFloatDrift, relFloatDrift*math.Abs(scaled)) {
		return 0, badRequest("an amount in the sale cannot have a fraction of a kobo")
	}
	return int(kobo), nil
}

// priceSale prices a new sale from the price list and checks the till's numbers against its own.
// A price that moved since the item went into the cart is a 409 `price_changed`; any other
// disagreement is a 400. Nothing is written.
func (s *saleService) priceSale(
	ctx context.Context, tx *sqlx.Tx, sale types.Sale,
) (pricedSale, error) {
	if len(sale.Items) == 0 {
		return pricedSale{}, badRequest("a sale needs at least one item")
	}

	priced := pricedSale{items: make([]model.SaleItem, len(sale.Items))}
	for i, v := range sale.Items {
		if v.Quantity <= 0 || v.Quantity > maxSaleQuantity {
			return pricedSale{}, badRequest("each item needs a quantity from 1 to 1,000,000")
		}
		price, err := s.repo.FetchPriceByID(ctx, v.PriceID)
		if errors.Is(err, sql.ErrNoRows) {
			// Product edit can delete a price a cart still holds: same remedy as a changed price.
			refusal := httperror.Conflict(
				"A price in the cart is no longer on file. Remove the item and add it again.",
				errors.New("price not found"),
			)
			refusal.Reason = "price_changed"
			return pricedSale{}, refusal
		}
		if err != nil {
			log.Println(err)
			return pricedSale{}, httperror.ServerError("failed to fetch product price", err)
		}
		if price.ProductID != v.ProductID {
			return pricedSale{}, badRequest("a price in the sale does not belong to its product")
		}

		sent, err := toKobo(v.UnitPrice)
		if err != nil {
			return pricedSale{}, err
		}
		if sent != price.SellingPriceKobo {
			return pricedSale{}, s.priceChanged(ctx, tx, v.ProductID, sent, price.SellingPriceKobo)
		}

		if price.QuantityPerUnit <= 0 {
			return pricedSale{}, badRequest(
				"a price in the sale does not say how much stock it sells",
			)
		}
		// The stock taken off must fit stock_movement.quantity, an INTEGER.
		if v.Quantity*price.QuantityPerUnit > math.MaxInt32 {
			return pricedSale{}, badRequest("the quantity of an item is too large")
		}
		gross := price.SellingPriceKobo * v.Quantity
		discount, err := toKobo(v.Discount)
		if err != nil {
			return pricedSale{}, err
		}
		if discount < 0 || discount > gross {
			return pricedSale{}, badRequest(
				"a discount must be between nothing and the item's price",
			)
		}
		lineTotal, err := toKobo(v.Total)
		if err != nil {
			return pricedSale{}, err
		}
		if lineTotal != gross-discount {
			return pricedSale{}, badRequest("the total of an item does not add up")
		}

		priced.items[i] = model.SaleItem{
			ProductID:            v.ProductID,
			Quantity:             v.Quantity,
			StockQuantityPerUnit: price.QuantityPerUnit,
			UnitPrice:            price.SellingPriceKobo,
			Discount:             discount,
			TotalPrice:           lineTotal,
		}
		priced.subtotal += gross
		priced.discount += discount
		priced.total += lineTotal
	}
	if priced.subtotal > math.MaxInt32 {
		return pricedSale{}, badRequest("the sale is too large")
	}

	for _, claim := range []struct {
		sent float64
		want int
		name string
	}{
		{sale.Subtotal, priced.subtotal, "subtotal"},
		{sale.Discount, priced.discount, "discount"},
		{sale.Total, priced.total, "total"},
	} {
		sent, err := toKobo(claim.sent)
		if err != nil {
			return pricedSale{}, err
		}
		if sent != claim.want {
			return pricedSale{}, badRequest(fmt.Sprintf("the sale %s does not add up", claim.name))
		}
	}

	priced.payments = make([]model.SalePayment, len(sale.Payments))
	for i, v := range sale.Payments {
		amount, err := toKobo(v.Amount)
		if err != nil {
			return pricedSale{}, err
		}
		if amount < 0 {
			return pricedSale{}, badRequest("a payment cannot be negative")
		}
		priced.payments[i] = model.SalePayment{
			PaymentMethod: constant.NormalizePaymentMethod(v.PaymentMethod),
			Amount:        amount,
		}
	}
	return priced, nil
}

// priceChanged is the refusal for a cart whose price is not the one on file any more.
func (s *saleService) priceChanged(
	ctx context.Context, tx *sqlx.Tx, productID, sent, current int,
) error {
	name := "an item in the cart"
	if products, err := s.repo.BulkFetchProductByIDsTx(ctx, tx, []int{productID}); err == nil &&
		len(products) > 0 {
		name = products[0].Name
	}
	refusal := httperror.Conflict(
		fmt.Sprintf(
			"The price of %s is now %s, not %s. Remove it from the cart and add it again.",
			name, model.ProductPrice{SellingPriceKobo: current}.PriceString(),
			model.ProductPrice{SellingPriceKobo: sent}.PriceString(),
		),
		errors.New("price changed since the item was added"),
	)
	refusal.Reason = "price_changed"
	return refusal
}
