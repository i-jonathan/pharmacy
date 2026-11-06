// One receipt for the whole app, laid out for an 80mm thermal roll. The POS prints it when a sale
// completes and Sales History prints it again as a marked copy.
//
// buildReceiptHtml(receipt) returns a complete HTML document; printReceiptHtml(html) prints it
// through a hidden frame, so there is no popup for the browser to block.
//
// receipt = {
//   receiptNumber, date (Date or ISO text), customer?,
//   items: [{ name, qty, price, discount? }]          money in naira
//   discount?, total, payments: [{ method, amount }], change,
//   returns?: [{ name, qty, refund }],
//   reprint?: { at } | true                            marks the slip as a copy
// }

const STORE_NAME = "Primocrest Pharmacy";

export function escapeHtml(value) {
  return String(value ?? "")
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#39;");
}

function money(n) {
  return "&#8358;" + Number(n || 0).toLocaleString("en-NG", { maximumFractionDigits: 2 });
}

function when(value) {
  const d = value instanceof Date ? value : new Date(value);
  return Number.isNaN(d.getTime())
    ? ""
    : d.toLocaleString([], { dateStyle: "short", timeStyle: "short" });
}

function row(left, right, extra = "") {
  return `<div class="row ${extra}"><span>${left}</span><span>${right}</span></div>`;
}

export function buildReceiptHtml(r) {
  const items = (r.items || [])
    .map((item) => {
      const gross = item.price * item.qty;
      const lineTotal = Math.max(0, gross - (item.discount || 0));
      const discount = item.discount ? row("Discount", "-" + money(item.discount), "small") : "";
      return `<div class="item">
        <div class="name">${escapeHtml(item.name)}</div>
        ${row(`${escapeHtml(item.qty)} x ${money(item.price)}`, money(lineTotal))}
        ${discount}
      </div>`;
    })
    .join("");

  const payments = (r.payments || [])
    .map((p) => row(escapeHtml(p.method), money(p.amount)))
    .join("");

  const returns = (r.returns || []).length
    ? `<hr/><div class="center bold">RETURNED</div>` +
      r.returns
        .map((x) => row(`${escapeHtml(x.name)} x ${escapeHtml(x.qty)}`, "-" + money(x.refund)))
        .join("")
    : "";

  const reprint = r.reprint
    ? `<div class="center bold">*** REPRINT${
        r.reprint.at ? " " + escapeHtml(when(r.reprint.at)) : ""
      } ***</div>`
    : "";

  return `<!doctype html>
<html>
<head>
<meta charset="utf-8"/>
<title>Receipt ${escapeHtml(r.receiptNumber)}</title>
<style>
  @page { size: 80mm auto; margin: 0; }
  * { box-sizing: border-box; }
  html, body { margin: 0; padding: 0; background: #fff; }
  body { width: 72mm; margin: 0 auto; padding: 4mm 2mm; font-family: monospace; font-size: 12px; line-height: 1.35; color: #000; }
  .center { text-align: center; }
  .bold { font-weight: bold; }
  .big { font-size: 14px; }
  .small { font-size: 11px; }
  .row { display: flex; justify-content: space-between; gap: 8px; }
  .row span:first-child { min-width: 0; overflow-wrap: anywhere; }
  .row span:last-child { white-space: nowrap; }
  .item { margin: 4px 0; }
  .name { overflow-wrap: anywhere; }
  hr { border: 0; border-top: 1px dashed #000; margin: 6px 0; }
</style>
</head>
<body>
  <div class="center bold big">${STORE_NAME}</div>
  <div class="center small">${escapeHtml(when(r.date))}${
    r.receiptNumber ? " &middot; #" + escapeHtml(r.receiptNumber) : ""
  }</div>
  ${reprint}
  <hr/>
  ${items}
  <hr/>
  ${r.discount > 0 ? row("You saved", money(r.discount), "small") : ""}
  ${row("Total", money(r.total), "bold")}
  ${payments}
  ${r.change > 0 ? row("Change", money(r.change)) : ""}
  ${returns}
  <hr/>
  ${r.customer ? `<div class="center small">${escapeHtml(r.customer)}</div>` : ""}
  <div class="center small" style="margin-top:6px">Thank you</div>
</body>
</html>`;
}

let frame = null;

// printReceiptHtml prints the document in a hidden frame and settles once the print dialog has
// been handed over (or immediately, on a kiosk-printing till). It rejects if the browser refuses
// or the frame does not load within five seconds.
export function printReceiptHtml(html) {
  return new Promise((resolve, reject) => {
    frame?.remove();
    const mine = document.createElement("iframe");
    frame = mine;
    mine.setAttribute("aria-hidden", "true");
    mine.style.cssText = "position:fixed;right:0;bottom:0;width:0;height:0;border:0;";
    // A frame that never loads must not leave the caller waiting; a frame replaced by a newer
    // print is simply done.
    const giveUp = setTimeout(() => {
      if (frame === mine) reject(new Error("The receipt did not load for printing."));
      else resolve();
    }, 5000);
    mine.onload = () => {
      clearTimeout(giveUp);
      try {
        mine.contentWindow.focus();
        mine.contentWindow.print();
        resolve();
      } catch (e) {
        reject(e);
      }
    };
    mine.onerror = () => {
      clearTimeout(giveUp);
      reject(new Error("The receipt could not be loaded for printing."));
    };
    mine.srcdoc = html;
    document.body.appendChild(mine);
  });
}
