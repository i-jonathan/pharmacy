// A purchase order line can receive more than it ordered, so what is left on a line and what came
// in over it are worked out here once: an excess on one line never shows as a negative shortfall
// and never offsets a shortfall on another.

// Units still to arrive on a line; 0 once the line has received what was ordered or more.
export function outstandingOf(line) {
  return Math.max(0, line.quantity - line.received_quantity);
}

// Units received beyond what the line ordered.
export function extraOf(line) {
  return Math.max(0, line.received_quantity - line.quantity);
}
