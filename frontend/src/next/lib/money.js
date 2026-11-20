// Prices are typed in naira; the server stores whole kobo, rounding to the nearest one
// (service/money.go), and refuses a price that rounds to 0 kobo. These checks match it.

// amountOf reads a typed amount: a cleared number box ("" from v-model.number), or text that
// isn't a number, is 0.
export function amountOf(v) {
  const n = Number(v);
  return Number.isFinite(n) ? n : 0;
}

// underAKobo is true for a price the server would refuse as under ₦0.01.
export function underAKobo(v) {
  return Math.round(amountOf(v) * 100) < 1;
}

// priceOptionLabel names price option n (counting from 1) the way the server does: by its name
// when it has one, by its place otherwise.
export function priceOptionLabel(n, name) {
  const trimmed = (name || "").trim();
  return trimmed ? `price of "${trimmed}"` : `price option ${n}`;
}
