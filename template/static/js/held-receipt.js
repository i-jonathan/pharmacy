// ==============================
// DOM ELEMENTS
// ==============================
const tableBody = document.getElementById("held-receipts-body");
const salePanel = document.getElementById("sale-panel");
const closePanelBtn = document.getElementById("close-panel");
const saleMeta = document.getElementById("sale-meta");
const itemsBody = document.getElementById("items-body");
const saleTotalEl = document.getElementById("sale-total");
const paymentsBody = document.getElementById("payments-body");
const totalPaidEl = document.getElementById("total-paid");
const restoreBtn = document.getElementById("restore-btn");

// ==============================
// GLOBALS
// ==============================
let heldSales = window.heldSales || [];
let rows = [];
let selectedRowIndex = -1;
let panelOpen = false;
let currentHeld = null;

// ==============================
// MAIN RENDER FUNCTION
// ==============================
function renderHeldTable() {
  tableBody.innerHTML = "";

  if (!heldSales || heldSales.length === 0) {
    tableBody.innerHTML = `
      <tr>
        <td colspan="5" class="text-center py-6 text-gray-500">
          No held receipts found.
        </td>
      </tr>`;
    return;
  }

  heldSales.forEach((sale, idx) => {
    const { reference, updated_at, payload } = sale;
    const cart = payload.cart || [];
    const total = cart.reduce(
      (sum, item) =>
        sum + parseFloat(item.price || 0) * parseFloat(item.qty || 1),
      0,
    );

    const tr = document.createElement("tr");
    tr.tabIndex = 0;
    tr.className =
      "cursor-pointer hover:bg-emerald-50 dark:hover:bg-emerald-900/30 focus:outline-none";
    tr.innerHTML = `
      <td class="px-4 py-3 font-medium">${reference}</td>
      <td class="px-4 py-3">${new Date(updated_at).toLocaleString()}</td>
      <td class="px-4 py-3">${cart.length}</td>
      <td class="px-4 py-3">₦${total.toLocaleString()}</td>
    `;

    tr.addEventListener("click", () => {
      selectRow(idx, false);
      openPanelFor(idx);
    });

    tr.addEventListener("keydown", (e) => {
      if (e.key === "Enter" || e.key === " ") {
        e.preventDefault();
        // toggle panel for this row
        if (panelOpen && selectedRowIndex === idx) {
          closePanel();
        } else {
          selectRow(idx, true);
          openPanelFor(idx);
        }
      }
    });

    tableBody.appendChild(tr);
  });

  rows = Array.from(tableBody.querySelectorAll("tr"));
}

// ==============================
// ROW SELECTION
// ==============================
function selectRow(idx, scrollIntoView = false) {
  if (idx < 0 || idx >= rows.length) return;

  // clear previous selection
  if (selectedRowIndex !== -1 && rows[selectedRowIndex]) {
    rows[selectedRowIndex].classList.remove(
      "bg-emerald-50",
      "dark:bg-emerald-900/20",
      "font-semibold",
      "ring-2",
      "ring-emerald-400",
    );
    rows[selectedRowIndex].setAttribute("aria-selected", "false");
  }

  selectedRowIndex = idx;
  const row = rows[idx];

  // add selection styles
  row.classList.add(
    "bg-emerald-50",
    "dark:bg-emerald-900/20",
    "font-semibold",
    "ring-2",
    "ring-emerald-400",
  );
  row.setAttribute("aria-selected", "true");
  row.focus({ preventScroll: !scrollIntoView });

  if (scrollIntoView) {
    row.scrollIntoView({ block: "nearest" });
  }
}

// ==============================
// OPEN DETAILS PANEL
// ==============================
function openPanelFor(idx) {
  const sale = heldSales[idx];
  if (!sale) return;
  currentHeld = sale;

  const { reference, payload, updated_at } = sale;
  const { cart, payments } = payload;

  saleMeta.textContent = `${reference} • ${new Date(
    updated_at,
  ).toLocaleString()}`;

  // Items
  itemsBody.innerHTML = "";
  let total = 0;
  cart.forEach((item) => {
    const price = parseFloat(item.price || 0);
    const qty = parseFloat(item.qty || 1);
    const lineTotal = price * qty;
    total += lineTotal;
    const tr = document.createElement("tr");
    tr.innerHTML = `
      <td class="px-4 py-2">
        <div class="font-medium">${item.name}</div>
        <div class="text-xs text-gray-500">${item.manufacturer || ""}</div>
      </td>
      <td class="px-4 py-2 text-right">${qty}</td>
      <td class="px-4 py-2 text-right">₦${price.toLocaleString()}</td>
      <td class="px-4 py-2 text-right">₦${lineTotal.toLocaleString()}</td>
    `;
    itemsBody.appendChild(tr);
  });
  saleTotalEl.textContent = `₦${total.toLocaleString()}`;

  // Payments
  paymentsBody.innerHTML = "";
  const paymentKeys = Object.keys(payments || {});
  let totalPaid = 0;
  if (paymentKeys.length === 0) {
    paymentsBody.innerHTML = `
      <tr><td colspan="2" class="px-3 py-2 text-center text-gray-500">No payments recorded</td></tr>`;
  } else {
    paymentKeys.forEach((method) => {
      const amount = parseFloat(payments[method]) || 0;
      totalPaid += amount;
      const tr = document.createElement("tr");
      tr.innerHTML = `
        <td class="px-3 py-2">${method}</td>
        <td class="px-3 py-2 text-right">₦${amount.toLocaleString()}</td>
      `;
      paymentsBody.appendChild(tr);
    });
  }
  totalPaidEl.textContent = `₦${totalPaid.toLocaleString()}`;

  // open panel
  salePanel.classList.add("open");
  panelOpen = true;
}

// ==============================
// CLOSE PANEL
// ==============================
function closePanel() {
  salePanel.classList.remove("open");
  panelOpen = false;
  if (selectedRowIndex !== -1 && rows[selectedRowIndex]) {
    rows[selectedRowIndex].classList.remove(
      "bg-emerald-50",
      "dark:bg-emerald-900/20",
      "font-semibold",
      "ring-2",
      "ring-emerald-400",
    );
  }
  currentHeld = null;
  saleMeta.textContent = "Select a sale to view details.";
}

// ==============================
// RESTORE SALE
// ==============================
function restoreHeldSale() {
  if (!currentHeld) return alert("No sale selected.");
  localStorage.setItem("heldSale", JSON.stringify(currentHeld));
  window.location.href = "/sales/receipt";
}

// ==============================
// DELETE SALE
// ==============================
function deleteHeldSale(ref) {
  if (!confirm("Are you sure you want to delete this held sale?")) return;
  heldSales = heldSales.filter((s) => s.reference !== ref);
  renderHeldTable();
  closePanel();
}

// ==============================
// KEYBOARD NAVIGATION
// ==============================
document.addEventListener("keydown", (e) => {
  if (!rows.length) return;

  if (e.key === "ArrowDown") {
    e.preventDefault();
    const next =
      selectedRowIndex === -1
        ? 0
        : Math.min(selectedRowIndex + 1, rows.length - 1);
    selectRow(next, true);
    if (panelOpen) openPanelFor(next);
  }

  if (e.key === "ArrowUp") {
    e.preventDefault();
    const prev =
      selectedRowIndex === -1 ? 0 : Math.max(selectedRowIndex - 1, 0);
    selectRow(prev, true);
    if (panelOpen) openPanelFor(prev);
  }

  if (e.key === "Escape" && panelOpen) {
    closePanel();
  }
});

// ==============================
// EVENTS
// ==============================
closePanelBtn.addEventListener("click", closePanel);
restoreBtn.addEventListener("click", restoreHeldSale);

// ==============================
// INIT
// ==============================
renderHeldTable();
