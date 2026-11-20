package service

import (
	"fmt"
	"math"
	"strings"
)

// koboFits reports whether a rounded kobo amount is a finite number the database's INTEGER
// money columns can hold.
func koboFits(kobo float64) bool {
	return !math.IsNaN(kobo) && !math.IsInf(kobo, 0) && math.Abs(kobo) <= math.MaxInt32
}

// priceKobo turns a price typed in naira into whole kobo, rounding to the nearest kobo rather
// than truncating: 19.99 is 1998.999… as a float and must be stored as 1999. A fraction of a
// kobo (a receiving pre-fill of 70% of ₦123.45 is ₦86.415) is rounded too, since nobody can see
// it on screen. A negative, non-finite or too-big amount is a 400 naming the price (`what`, e.g.
// "product 2: cost price"). A price that must be above zero uses positivePriceKobo.
func priceKobo(what string, naira float64) (int, error) {
	kobo := math.Round(naira * 100)
	switch {
	case math.IsNaN(kobo):
		return 0, badRequest(what + " is not a number")
	case kobo < 0:
		return 0, badRequest(what + " cannot be negative")
	case !koboFits(kobo):
		return 0, badRequest(fmt.Sprintf("%s is too large (the most is ₦21,474,836.47)", what))
	}
	return int(kobo), nil
}

// positivePriceKobo is priceKobo for a price that must be above zero. The kobo is checked, not
// the naira, so a price that rounds to 0 (₦0.004) is refused too.
func positivePriceKobo(what string, naira float64) (int, error) {
	kobo, err := priceKobo(what, naira)
	if err == nil && kobo <= 0 {
		return 0, badRequest(what + " must be at least ₦0.01")
	}
	return kobo, err
}

// priceOptionLabel names price option n (counting from 1) in a refusal: by its name when it has
// one, by its place otherwise, so an unnamed option doesn't read as `price of ""`.
func priceOptionLabel(n int, name string) string {
	if name = strings.TrimSpace(name); name != "" {
		return fmt.Sprintf("price of %q", name)
	}
	return fmt.Sprintf("price option %d", n)
}
