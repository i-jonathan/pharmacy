const addModal = document.getElementById("add-modal");
const editModal = document.getElementById("edit-modal");

document.getElementById("add-btn").addEventListener("click", () => {
  addModal.classList.remove("hidden");
});

function closeAddModal() {
  addModal.classList.add("hidden");
}

let currentEditingId = null;

async function openEditModal(id) {
  currentEditingId = id;
  try {
    const response = await fetch(`/inventory/product/${id}`);
    if (!response.ok) throw new Error("Failed to fetch product details");
    const product = await response.json();

    // Set core fields
    document.getElementById("edit-name").value = product.name;
    document.getElementById("edit-barcode").value = product.barcode || "";
    document.getElementById("edit-manufacturer").value = product.manufacturer;
    document.getElementById("edit-category").value = product.category_id;
    document.getElementById("edit-stock").value = product.stock || 0;
    document.getElementById("edit-reorder-level").value = product.reorder_level;
    document.getElementById("edit-cost-price").value = product.cost_price;

    // Render price rows
    const container = document.getElementById("edit-prices-container");
    container.innerHTML = "";

    // Add default price first
    if (product.default_price) {
      addPriceRow({
        id: product.default_price.id,
        name: product.default_price.name,
        selling_price: product.default_price.selling_price,
        quantity_per_unit: product.default_price.quantity,
        is_default: true,
      });
    }

    // Add other prices
    if (product.price_options) {
      product.price_options.forEach((p) => {
        if (p.id !== product.default_price.id) {
          addPriceRow({
            id: p.id,
            name: p.name,
            selling_price: p.selling_price,
            quantity_per_unit: p.quantity,
            is_default: false,
          });
        }
      });
    }

    if (!product.default_price && (!product.price_options || product.price_options.length === 0)) {
      addPriceRow({ name: "Base", selling_price: 0, quantity_per_unit: 1, is_default: true });
    }

    editModal.classList.remove("hidden");
  } catch (err) {
    console.error(err);
    alert("Error loading product details: " + err.message);
  }
}

function addPriceRow(data = {}) {
  const container = document.getElementById("edit-prices-container");
  const row = document.createElement("div");
  row.className = "price-row grid grid-cols-12 gap-2 items-center bg-white dark:bg-gray-800 p-2 rounded border border-gray-200 dark:border-gray-700";
  
  const idInput = data.id ? `<input type="hidden" class="price-id" value="${data.id}">` : "";
  
  row.innerHTML = `
    ${idInput}
    <div class="col-span-4">
      <input type="text" placeholder="Name (e.g. Wholesale)" class="price-name w-full text-sm px-2 py-1 border rounded dark:bg-gray-700 dark:border-gray-600" value="${data.name || ""}" required>
    </div>
    <div class="col-span-3">
      <input type="number" step="0.01" placeholder="Price" class="price-selling w-full text-sm px-2 py-1 border rounded dark:bg-gray-700 dark:border-gray-600" value="${data.selling_price || ""}" required>
    </div>
    <div class="col-span-2">
      <input type="number" placeholder="Qty" class="price-qty w-full text-sm px-2 py-1 border rounded dark:bg-gray-700 dark:border-gray-600" value="${data.quantity_per_unit || 1}" required title="Quantity per unit">
    </div>
    <div class="col-span-2 flex items-center justify-center gap-1">
      <input type="radio" name="default-price" class="price-default" ${data.is_default ? "checked" : ""} title="Set as default">
      <span class="text-[10px] text-gray-500 uppercase font-bold">Def</span>
    </div>
    <div class="col-span-1 text-right">
      <button type="button" onclick="this.closest('.price-row').remove()" class="text-red-500 hover:text-red-700">✕</button>
    </div>
  `;
  container.appendChild(row);
}

function closeEditModal() {
  editModal.classList.add("hidden");
  currentEditingId = null;
}

document.getElementById("edit-item-form").addEventListener("submit", async (e) => {
  e.preventDefault();
  if (!currentEditingId) return;

  const btn = e.target.querySelector('button[type="submit"]');
  btn.disabled = true;
  btn.innerText = "Updating...";

  try {
    const priceRows = document.querySelectorAll(".price-row");
    const prices = Array.from(priceRows).map(row => {
      const id = row.querySelector(".price-id")?.value;
      return {
        id: id ? parseInt(id) : null,
        name: row.querySelector(".price-name").value,
        selling_price: parseFloat(row.querySelector(".price-selling").value),
        quantity_per_unit: parseInt(row.querySelector(".price-qty").value),
        is_default: row.querySelector(".price-default").checked
      };
    });

    const payload = {
      id: parseInt(currentEditingId),
      name: document.getElementById("edit-name").value,
      barcode: document.getElementById("edit-barcode").value,
      manufacturer: document.getElementById("edit-manufacturer").value,
      category_id: parseInt(document.getElementById("edit-category").value),
      reorder_level: parseInt(document.getElementById("edit-reorder-level").value),
      stock: parseInt(document.getElementById("edit-stock").value),
      cost_price: parseFloat(document.getElementById("edit-cost-price").value),
      prices: prices
    };

    const response = await fetch(`/inventory/product/${currentEditingId}`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload)
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.message || "Failed to update product");
    }

    // Success
    showToast("Product updated successfully", { type: "success" });
    
    closeEditModal();
    // Wait a bit before reload so they see the toast if it's on the same page
    setTimeout(() => window.location.reload(), 1000);
  } catch (err) {
    console.error(err);
    alert("Error updating product: " + err.message);
  } finally {
    btn.disabled = false;
    btn.innerText = "Update";
  }
});

let sortDirection = {}; // track asc/desc per column

function sortTable(column) {
  const tableBody = document.getElementById("inventory-body");
  const rows = Array.from(tableBody.querySelectorAll("tr"));

  // map column names to indexes
  const colIndex = {
    name: 0,
    category: 1,
    manufacturer: 2,
    stock: 3,
    price: 4,
    expiry: 5,
  }[column];

  // toggle sort direction (default ascending if not sorted yet)
  sortDirection[column] = !sortDirection[column];
  const ascending = sortDirection[column];

  // sort rows
  rows.sort((a, b) => {
    const aText = a.children[colIndex].innerText.trim();
    const bText = b.children[colIndex].innerText.trim();

    if (column === "stock" || column === "price") {
      const aNum = parseFloat(aText.replace(/[₦,]/g, "")) || 0;
      const bNum = parseFloat(bText.replace(/[₦,]/g, "")) || 0;
      return ascending ? aNum - bNum : bNum - aNum;
    }

    if (column === "expiry") {
      const aDate = new Date(aText);
      const bDate = new Date(bText);
      return ascending ? aDate - bDate : bDate - aDate;
    }

    return ascending ? aText.localeCompare(bText) : bText.localeCompare(aText);
  });

  // re-append sorted rows
  rows.forEach((row) => tableBody.appendChild(row));

  // reset all icons to gray ▲
  document.querySelectorAll("[id^='sort-icon-']").forEach((icon) => {
    icon.innerText = "▲";
    icon.classList.add("text-gray-400");
    icon.classList.remove("text-emerald-500");
  });

  // update active column icon
  const activeIcon = document.getElementById(`sort-icon-${column}`);
  activeIcon.innerText = ascending ? "▲" : "▼";
  activeIcon.classList.remove("text-gray-400");
  activeIcon.classList.add("text-emerald-500");
}

function searchTable() {
  const query = document.getElementById("searchBox").value.toLowerCase();
  const rows = document.querySelectorAll("#inventory-body tr");

  rows.forEach((row) => {
    const rowText = row.innerText.toLowerCase();
    row.style.display = rowText.includes(query) ? "" : "none";
  });
}
