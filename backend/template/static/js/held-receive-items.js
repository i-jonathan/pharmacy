const tableBody = document.getElementById("held-receive-body");
const receivePanel = document.getElementById("receive-panel");
const closePanelBtn = document.getElementById("close-panel");
const receiveMeta = document.getElementById("receive-meta");
const itemsBody = document.getElementById("items-body");
const receiveTotalEl = document.getElementById("receive-total");
const restoreBtn = document.getElementById("restore-btn");

let heldReceiveItems = window.heldReceiveItems || [];
let rows = [];
let selectedRowIndex = -1;
let panelOpen = false;
let currentHeld = null;

function getProducts(held) {
  return held.payload?.products || [];
}

function getTotalCost(products) {
  return products.reduce((sum, item) => {
    const cost = parseFloat(item.cost_price || 0);
    const qty = parseInt(item.quantity || 0, 10);
    return sum + cost * qty;
  }, 0);
}

function renderHeldTable() {
  tableBody.innerHTML = "";

  if (!heldReceiveItems || heldReceiveItems.length === 0) {
    tableBody.innerHTML = `
      <tr>
        <td colspan="5" class="text-center py-6 text-gray-500">
          No held receive items found.
        </td>
      </tr>`;
    return;
  }

  heldReceiveItems.forEach((held, idx) => {
    const products = getProducts(held);
    const total = getTotalCost(products);
    const tr = document.createElement("tr");
    tr.tabIndex = 0;
    tr.className =
      "cursor-pointer hover:bg-emerald-50 dark:hover:bg-emerald-900/30 focus:outline-none";
    tr.innerHTML = `
      <td class="px-4 py-3 font-medium">${held.reference}</td>
      <td class="px-4 py-3">${new Date(held.updated_at).toLocaleString()}</td>
      <td class="px-4 py-3">${held.payload?.supplier || ""}</td>
      <td class="px-4 py-3">${products.length}</td>
      <td class="px-4 py-3">₦${total.toLocaleString()}</td>
    `;

    tr.addEventListener("click", () => {
      selectRow(idx, false);
      openPanelFor(idx);
    });

    tr.addEventListener("keydown", (e) => {
      if (e.key === "Enter" || e.key === " ") {
        e.preventDefault();
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

function selectRow(idx, scrollIntoView = false) {
  if (idx < 0 || idx >= rows.length) return;

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

function openPanelFor(idx) {
  const held = heldReceiveItems[idx];
  if (!held) return;
  currentHeld = held;

  const products = getProducts(held);
  receiveMeta.textContent = `${held.reference} • ${new Date(
    held.updated_at,
  ).toLocaleString()} • Supplier: ${held.payload?.supplier || ""}`;

  itemsBody.innerHTML = "";
  let total = 0;
  products.forEach((item) => {
    const cost = parseFloat(item.cost_price || 0);
    const qty = parseInt(item.quantity || 0, 10);
    const lineTotal = cost * qty;
    total += lineTotal;

    const tr = document.createElement("tr");
    tr.innerHTML = `
      <td class="px-4 py-2">
        <div class="font-medium">${item.name}</div>
        <div class="text-xs text-gray-500">${item.manufacturer || ""}</div>
      </td>
      <td class="px-4 py-2 text-right">${qty}</td>
      <td class="px-4 py-2 text-right">₦${cost.toLocaleString()}</td>
      <td class="px-4 py-2 text-right">₦${lineTotal.toLocaleString()}</td>
    `;
    itemsBody.appendChild(tr);
  });

  receiveTotalEl.textContent = `₦${total.toLocaleString()}`;
  receivePanel.classList.add("open");
  panelOpen = true;
}

function closePanel() {
  receivePanel.classList.remove("open");
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
  receiveMeta.textContent = "Select held items to view details.";
}

function restoreHeldReceiveItems() {
  if (!currentHeld) return alert("No held items selected.");
  localStorage.setItem("heldReceiveItems", JSON.stringify(currentHeld));
  window.location.href = "/inventory/receive-items";
}

async function deleteHeldReceiveItems() {
  if (!currentHeld) return alert("No held items selected.");
  const reference = currentHeld.reference;
  if (!confirm("Are you sure you want to delete these held items?")) return;

  try {
    const res = await fetch(
      `/inventory/receive-items/held/${encodeURIComponent(reference)}`,
      {
        method: "DELETE",
        headers: { "Content-Type": "application/json" },
      },
    );

    if (!res.ok) {
      const msg = await res.text();
      throw new Error(msg || "Failed to delete held items");
    }

    heldReceiveItems = heldReceiveItems.filter(
      (item) => item.reference !== reference,
    );
    renderHeldTable();
    closePanel();
    showToast("Held receiving items deleted successfully");
  } catch (err) {
    console.error("Delete error:", err);
    alert(`Failed to delete held items. ${err.message}`);
  }
}

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

closePanelBtn.addEventListener("click", closePanel);
restoreBtn.addEventListener("click", restoreHeldReceiveItems);

renderHeldTable();
