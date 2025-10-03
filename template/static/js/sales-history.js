// ===== Helpers =====
const fmtNaira = (n) => "₦" + Number(n || 0).toLocaleString("en-NG");
const byDateDesc = (a, b) => new Date(b.date) - new Date(a.date);

// ===== DOM =====
const salesBody = document.getElementById("sales-body");
const panel = document.getElementById("sale-panel");
const closeBtn = document.getElementById("close-panel");
const totalEl = document.getElementById("filter-total");
const saleMeta = document.getElementById("sale-meta");
const paymentsBody = document.getElementById("payments-body");
const totalPaidCell = document.getElementById("total-paid");
const itemsBody = document.getElementById("items-body");
const saleTotalCell = document.getElementById("sale-total");
const saleChangeCell = document.getElementById("sale-change");
const startDate = document.getElementById("start-date");
const endDate = document.getElementById("end-date");
const rangeSelect = document.getElementById("predefined-range");

let rows = [];
let selectedRowIndex = -1;
let panelOpen = false;

document.addEventListener("DOMContentLoaded", () => {
  // select all in drop down by default
  rangeSelect.value = "all";
});

function renderTable() {
  salesBody.innerHTML = "";
  sales.data.forEach((s, idx) => {
    const tr = document.createElement("tr");
    tr.className =
      "cursor-pointer hover:bg-emerald-50 dark:hover:bg-emerald-900/30 focus:outline-none";
    tr.setAttribute("role", "row");
    tr.setAttribute("tabindex", "0");
    tr.dataset.index = idx;

    const dt = new Date(s.created_at);
    const dateDisplay = dt.toLocaleString([], {
      year: "numeric",
      month: "short",
      day: "2-digit",
      hour: "2-digit",
      minute: "2-digit",
    });

    tr.innerHTML = `
      <td class="px-4 py-3"> ${s.receipt_number} </td>
      <td class="px-4 py-3 whitespace-nowrap"> ${dateDisplay} </td>
      <td class="px-4 py-3"> ${s.items.length} </td>
      <td class="px-4 py-3"> ${fmtNaira(s.total)} </td>
      <td class="px-4 py-3"> ${s.cashier} </td>
    `;

    tr.addEventListener("click", () => {
      selectRow(idx, true);
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

    salesBody.appendChild(tr);
  });
  rows = Array.from(salesBody.querySelectorAll("tr"));

  totalEl.textContent = `Total: ₦${sales.total.toLocaleString()}`;
}

function selectRow(idx, scrollIntoView = false) {
  if (idx < 0 || idx >= rows.length) return;

  // clear previous selection
  if (selectedRowIndex !== -1) {
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
    "bg-emerald-50", // light mode soft green
    "dark:bg-emerald-900/20", // dark mode subtle overlay green
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

function openPanelFor(idx) {
  const sale = sales[idx];
  updatePanel(sale);
  if (!panelOpen) {
    panel.classList.add("open");
    panelOpen = true;
  }
}

function closePanel() {
  panel.classList.remove("open");
  panelOpen = false;
}

// ===== Panel rendering =====
function updatePanel(sale) {
  const dt = new Date(sale.created_at);
  saleMeta.textContent = `${sale.receipt_number} • ${dt.toLocaleString()} • Cashier: ${sale.cashier}`;

  // Payments
  paymentsBody.innerHTML = sale.payments
    .map(
      (p) => `
    <tr>
      <td class="px-3 py-2 capitalize">${p.method_name}</td>
      <td class="px-3 py-2 text-right">${fmtNaira(p.amount)}</td>
    </tr>
  `,
    )
    .join("");

  const totalPaid = sale.payments.reduce((sum, p) => sum + (p.amount || 0), 0);
  totalPaidCell.textContent = fmtNaira(totalPaid);

  // Items
  // <td class="px-3 py-2">
  //   <span class="font-medium">${it.product_name}</span>
  //   <span class="ml-1 text-sm text-gray-500">– ${it.manufacturer}</span>
  // </td>
  itemsBody.innerHTML = sale.items
    .map(
      (it) => `
    <tr>
      <td class="px-3 py-2">
        <div>
          <div class="font-medium">${it.product_name}</div>
          <div class="text-xs text-gray-500">${it.manufacturer}</div>
        </div>
      </td>
      <td class="px-3 py-2 text-right">${it.quantity}</td>
      <td class="px-3 py-2 text-right">${fmtNaira(it.unit_price)}</td>
      <td class="px-3 py-2 text-right">${fmtNaira(it.discount)}</td>
      <td class="px-3 py-2 text-right">${fmtNaira(it.quantity * it.unit_price)}</td>
    </tr>
  `,
    )
    .join("");

  saleTotalCell.textContent = fmtNaira(sale.total);
  const change = Math.max(0, totalPaid - sale.total);
  saleChangeCell.textContent = fmtNaira(change);
}

// ===== Keyboard navigation (non-blocking) =====
document.addEventListener("keydown", (e) => {
  if (e.key === "ArrowDown") {
    e.preventDefault();
    const next = Math.min(
      selectedRowIndex === -1 ? 0 : selectedRowIndex + 1,
      rows.length - 1,
    );
    selectRow(next, true);
    if (panelOpen) openPanelFor(next);
  }
  if (e.key === "ArrowUp") {
    e.preventDefault();
    const prev = Math.max(
      selectedRowIndex === -1 ? 0 : selectedRowIndex - 1,
      0,
    );
    selectRow(prev, true);
    if (panelOpen) openPanelFor(prev);
  }
  if (e.key === "Escape" && panelOpen) {
    closePanel();
  }
});

closeBtn.addEventListener("click", closePanel);

// Initial render
renderTable();
// Preselect the first row for quick nav
if (rows.length) selectRow(0);

function formatDate(date) {
  return date.toISOString().split("T")[0]; // YYYY-MM-DD
}

rangeSelect.addEventListener("change", () => {
  const today = new Date();
  let start, end;

  switch (rangeSelect.value) {
    case "today":
      start = new Date(today);
      end = new Date(today);
      break;

    case "yesterday":
      start = new Date(today);
      start.setDate(start.getDate() - 1);
      end = new Date(start);
      break;

    case "this-week":
      start = new Date(today);
      start.setDate(today.getDate() - today.getDay()); // Sunday
      end = new Date(today);
      break;

    case "last-week":
      start = new Date(today);
      start.setDate(today.getDate() - today.getDay() - 7); // previous Sunday
      end = new Date(start);
      end.setDate(start.getDate() + 6); // Saturday
      break;

    case "this-month":
      start = new Date(today.getFullYear(), today.getMonth(), 1);
      end = new Date(today);
      break;

    case "last-month":
      start = new Date(today.getFullYear(), today.getMonth() - 1, 1);
      end = new Date(today.getFullYear(), today.getMonth(), 0); // last day of prev month
      break;

    case "all":
      startDate.value = "";
      endDate.value = "";
      start = end = null;
      applyFilter();
      return;
    default:
      return;
  }

  if (start && end) {
    startDate.value = formatDate(start);
    endDate.value = formatDate(end);
  }

  applyFilter(start, end);
});

async function applyFilter(start, end) {
  const params = new URLSearchParams();
  if (start) params.set("start", formatDate(start));
  if (end) params.set("end", formatDate(end));

  try {
    const res = await fetch(`/sales/filter?${params.toString()}`);
    if (!res.ok) throw new Error("Failed to fetch sales");

    const data = await res.json();
    sales = data;

    renderTable();
    selectedRowIndex = -1;
  } catch (err) {
    console.error("Error fetching sales:", err);
  }
}
