const cart = [];
const payments = {};
let selectedPaymentMethod = "";
let searchTimeout;

const subtotalDisplay = document.getElementById("subtotal");
const paidDisplay = document.getElementById("paid");
const changeDisplay = document.getElementById("change");
const receiptItems = document.getElementById("receipt-items");
const paymentInput = document.getElementById("payment-amount");
const searchResults = document.getElementById("search-results");
const itemSearch = document.getElementById("item-search");

function updateTotals() {
  let subtotal = cart.reduce((sum, item) => sum + item.price * item.qty, 0);
  let paid = Object.values(payments).reduce((a, b) => a + b, 0);
  subtotalDisplay.textContent = subtotal.toLocaleString();
  paidDisplay.textContent = paid.toLocaleString();
  changeDisplay.textContent = Math.max(paid - subtotal, 0).toLocaleString();
  validateSale();
}

function addItem(item) {
  const existing = cart.find((i) => i.id === item.id);

  if (existing) {
    existing.qty++;
  } else {
    cart.push({
      id: item.id,
      name: item.name,
      manufacturer: item.manufacturer || "",
      price: item.default_price?.selling_price || 0,
      qty: 1,
      default_price: item.default_price,
      price_options: item.price_options,
      selected_price_id: item.default_price?.id,
    });
  }

  renderCart();
}

function renderCart() {
  receiptItems.innerHTML = "";

  cart.forEach((item, i) => {
    const selectedOption = item.price_options?.find(
      (opt) => opt.id === item.selected_price_id,
    );

    const isDiscounted =
      selectedOption?.selling_price &&
      item.price < selectedOption.selling_price;

    const row = document.createElement("tr");
    row.classList.remove("bg-green-50", "dark:bg-green-900/20");

    if (isDiscounted) {
      row.classList.add("bg-green-50", "dark:bg-green-900/20");
    }

    row.dataset.index = i;

    row.innerHTML = `
      <td class="px-4 py-2 w-[40%]">
        <span class="font-medium">${item.name}</span><span class="ml-1 text-sm text-gray-500 dark:text-gray-400">- ${item.manufacturer}</span>
      </td>
      <td class="px-4 py-2 w-[20%]">
        <div class="flex items-center justify-center gap-3">
          <button class="qty-btn text-red-500 border rounded w-6 h-6 flex items-center justify-center"
            data-index="${i}" data-action="dec">-</button>
          <input type="number" value="${item.qty}" min="1"
            class="no-spinners w-12 text-center px-2 py-1 border rounded dark:bg-gray-700 dark:border-gray-600 qty-input"
            data-index="${i}">
          <button class="qty-btn text-green-500 border rounded w-6 h-6 flex items-center justify-center"
            data-index="${i}" data-action="inc">+</button>
        </div>
      </td>

      <!-- Prices button column -->
      <td class="px-4 py-2 text-center w-[1%]">
        <button type="button"
          class="prices-toggle px-2 py-1 text-xs rounded bg-gray-100 dark:bg-gray-700
                  hover:bg-gray-200 dark:hover:bg-gray-600 text-gray-700 dark:text-gray-200"
          data-index="${i}">
          Prices
        </button>
      </td>

      <td class="px-4 py-2 text-center">
        <span class="price-display block cursor-pointer font-medium text-gray-800 dark:text-gray-200"
              data-index="${i}">
          ₦${item.price.toFixed(2)}
        </span>

        <input type="number" value="${item.price.toFixed(2)}" min="0"
          class="no-spinners price-input block hidden w-20 text-center px-2 py-1 border rounded
                  dark:bg-gray-700 dark:border-gray-600"
          data-index="${i}">
      </td>

      <td class="px-4 py-2 text-center">
        <div>
          <span class="total-display cursor-pointer" data-index="${i}">
            ₦${(item.qty * item.price).toFixed(2)}
          </span>
          <input type="number" value="${(item.qty * item.price).toFixed(2)}" min="0"
            class="no-spinners total-input hidden w-24 text-center px-2 py-1 border rounded dark:bg-gray-700 dark:border-gray-600"
            data-index="${i}">
        </div>

        ${
          isDiscounted
            ? `<div class="text-xs text-red-500 mt-1">
                -₦${((item.default_price.selling_price - item.price) * item.qty).toFixed(2)} discount
              </div>`
            : ""
        }
      </td>

      <td class="px-4 py-2 text-center w-[1%]">
        <button class="text-red-500" onclick="removeItem(${i})">🗑️</button>
      </td>
    `;

    receiptItems.appendChild(row);
  });

  // Quantity input listeners
  document.querySelectorAll(".qty-input").forEach((input) => {
    // Commit on blur
    input.addEventListener("blur", (e) => {
      const i = +e.target.dataset.index;
      cart[i].qty = Math.max(1, +e.target.value || 1); // avoid 0 or NaN
      renderCart();
    });

    // Commit on Enter
    input.addEventListener("keydown", (e) => {
      if (e.key === "Enter") {
        e.preventDefault(); // stop form submit
        const i = +e.target.dataset.index;
        cart[i].qty = Math.max(1, +e.target.value || 1);
        renderCart();
      }
    });
  });

  // Quantity buttons
  document.querySelectorAll(".qty-btn").forEach((btn) => {
    btn.addEventListener("click", (e) => {
      const i = +e.target.dataset.index;
      const action = e.target.dataset.action;
      if (action === "inc") cart[i].qty++;
      if (action === "dec" && cart[i].qty > 1) cart[i].qty--;
      renderCart();
    });
  });

  updateTotals();
  validateSale();
}

// handle clicking and editing unit price
document.addEventListener("click", (e) => {
  if (e.target.classList.contains("price-display")) {
    const span = e.target;
    const index = span.dataset.index;
    const input = document.querySelector(`.price-input[data-index="${index}"]`);

    span.classList.add("hidden");
    input.classList.remove("hidden");
    input.focus();
  }
});

function commitUnitPriceChange(input) {
  const index = input.dataset.index;
  const span = document.querySelector(`.price-display[data-index="${index}"]`);

  const newPrice = parseFloat(input.value) || 0;
  cart[index].price = newPrice;

  span.textContent = `₦${newPrice}`;
  span.classList.remove("hidden");
  input.classList.add("hidden");

  renderCart();
}

receiptItems.addEventListener(
  "blur",
  (e) => {
    if (e.target.classList.contains("price-input")) {
      commitUnitPriceChange(e.target);
    }
  },
  true,
);

receiptItems.addEventListener("keydown", (e) => {
  if (e.target.matches(".price-input") && e.key === "Enter") {
    e.preventDefault();
    commitUnitPriceChange(e.target);
  }
});

// handle clicking and manually updating total price
document.addEventListener("click", (e) => {
  if (e.target.classList.contains("total-display")) {
    const span = e.target;
    const index = span.dataset.index;
    const input = document.querySelector(`.total-input[data-index="${index}"]`);

    span.classList.add("hidden");
    input.classList.remove("hidden");
    input.focus();
  }
});

function commitTotalPriceChange(input) {
  const index = input.dataset.index;
  const span = document.querySelector(`.total-display[data-index="${index}"]`);

  const newTotal = parseFloat(input.value) || 0;
  const qty = cart[index].qty;

  // recalc unit price from total
  cart[index].price = qty > 0 ? newTotal / qty : 0;

  span.textContent = `₦${newTotal}`;
  span.classList.remove("hidden");
  input.classList.add("hidden");

  renderCart();
}

// Handle blur for total edit
receiptItems.addEventListener(
  "blur",
  (e) => {
    if (e.target.classList.contains("total-input")) {
      commitTotalPriceChange(e.target);
    }
  },
  true,
);

receiptItems.addEventListener("keydown", (e) => {
  if (e.target.matches(".total-input") && e.key === "Enter") {
    e.preventDefault();
    commitTotalPriceChange(e.target);
  }
});

function buildPriceRow(index) {
  const item = cart[index];
  const options = (item.price_options || [])
    .map(
      (opt) => `
      <button
        class="price-chip px-3 py-1 rounded-full border text-sm
          ${
            item.selected_price_id == opt.id
              ? "bg-emerald-500 text-white border-emerald-600"
              : "border-gray-300 dark:border-gray-600 hover:bg-gray-100 dark:hover:bg-gray-700"
          }"
        data-index="${index}"
        data-price="${opt.selling_price}"
        data-price-id="${opt.id}">
        ${opt.name} - ₦${opt.selling_price.toFixed(2)}
      </button>`,
    )
    .join("");

  const colCount = document
    .querySelector("#receipt-items")
    .closest("table")
    .querySelectorAll("thead th").length;

  const tr = document.createElement("tr");
  tr.className = "price-row bg-gray-50 dark:bg-gray-900/30";
  tr.dataset.index = index;
  tr.innerHTML = `
    <td colspan="${colCount}" class="px-4 py-3">
      <div class="flex flex-wrap items-center gap-2">
        ${options || `<span class="text-sm text-gray-500 dark:text-gray-400">No price options available.</span>`}

        <div class="ml-auto flex items-center gap-2">
          <span class="text-sm text-gray-600 dark:text-gray-300">Custom:</span>
          <input type="number" min="0" step="0.01"
            class="no-spinners price-custom w-28 px-2 py-1 border rounded
                   dark:bg-gray-700 dark:border-gray-600"
            data-index="${index}" value="${item.price.toFixed(2)}" />
          <button
            class="price-apply px-3 py-1 text-xs rounded bg-primary text-white hover:bg-emerald-700"
            data-index="${index}">
            Apply
          </button>
        </div>
      </div>
    </td>
  `;
  return tr;
}

function togglePriceRow(index) {
  // close any existing price row
  const open = document.querySelector(".price-row");
  if (open) {
    if (+open.dataset.index === index) {
      open.remove();
      return;
    }
    open.remove();
  }
  // insert after the item row
  const mainRow = document.querySelector(
    `#receipt-items tr[data-index="${index}"]`,
  );
  if (mainRow) {
    mainRow.insertAdjacentElement("afterend", buildPriceRow(index));
  }
}

// Open/close the inline price row
receiptItems.addEventListener("click", (e) => {
  const btn = e.target.closest(".prices-toggle");
  if (btn) {
    togglePriceRow(+btn.dataset.index);
  }

  const chip = e.target.closest(".price-chip");
  if (chip) {
    const idx = +chip.dataset.index;
    const price = parseFloat(chip.dataset.price) || 0;
    const priceId = chip.dataset.priceId || null;

    cart[idx].price = price;
    cart[idx].selected_price_id = priceId;
    renderCart();
  }

  const apply = e.target.closest(".price-apply");
  if (apply) {
    const idx = +apply.dataset.index;
    const input = document.querySelector(`.price-custom[data-index="${idx}"]`);
    const v = parseFloat(input.value);
    if (!isNaN(v)) {
      cart[idx].price = v;
      renderCart();
    }
  }
});

// Close the inline row when clicking outside it or the Prices button
document.addEventListener("click", (e) => {
  if (!e.target.closest(".price-row") && !e.target.closest(".prices-toggle")) {
    const open = document.querySelector(".price-row");
    if (open) open.remove();
  }
});

function removeItem(index) {
  cart.splice(index, 1);
  renderCart();
}

function debounce(func, delay = 300) {
  clearTimeout(searchTimeout);
  searchTimeout = setTimeout(func, delay);
}

function renderSearchResult(item) {
  const li = document.createElement("li");
  li.innerHTML = `
    <div class="flex flex-col">
      <span class="font-medium">${item.name}</span>
      <span class="text-sm text-gray-500 dark:text-gray-400">${item.manufacturer || ""}</span>
    </div>
  `;
  li.className =
    "px-4 py-2 cursor-pointer hover:bg-gray-200 dark:hover:bg-gray-700";
  li.onclick = () => {
    addItem(item);
    searchResults.classList.add("hidden");
    itemSearch.value = "";
  };
  return li;
}

async function fetchAndRenderResults(query) {
  try {
    const res = await fetch(
      `/inventory/search?query=${encodeURIComponent(query)}`,
    );
    const matches = await res.json();

    searchResults.innerHTML = "";
    if (!matches.length) {
      searchResults.classList.add("hidden");
      return;
    }

    matches.forEach((item) =>
      searchResults.appendChild(renderSearchResult(item)),
    );
    searchResults.classList.remove("hidden");
  } catch (err) {
    console.error("Search error", err);
    searchResults.classList.add("hidden");
  }
}

itemSearch.addEventListener("input", (e) => {
  const value = e.target.value.toLowerCase();
  if (!value) {
    searchResults.innerHTML = "";
    searchResults.classList.add("hidden");
    return;
  }

  debounce(() => fetchAndRenderResults(value), 300);
});

function openPayment(method) {
  selectedPaymentMethod = method;
  document.getElementById("payment-method-name").textContent =
    `Pay with ${method}`;
  paymentInput.value = payments[method] || "";
  document.getElementById("payment-modal").classList.remove("hidden");
}

function savePayment() {
  const amount = parseFloat(paymentInput.value || 0);
  if (amount > 0) {
    payments[selectedPaymentMethod] = amount;
  }
  document.getElementById("payment-modal").classList.add("hidden");
  updateTotals();
}

function cancelPayment() {
  delete payments[selectedPaymentMethod];
  document.getElementById("payment-modal").classList.add("hidden");
  updateTotals();
}

async function saveSale() {
  payload = buildSalePayload(cart, payments);
  try {
    const response = await fetch("/sales/", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(payload),
    });

    if (response.ok) {
      console.log("Sale saved successfully!");
      window.location.reload();
    } else {
      const errorData = await response.json().catch(() => ({}));
      console.error("Save sale failed:", errorData);
      alert("Failed to save sale. Please try again.");
    }
  } catch (error) {
    console.error("Error saving sale:", error);
    alert("An error occurred while saving the sale.");
  }
}

function validateSale() {
  let isValid = true;

  document.querySelectorAll("#receipt-items tr").forEach((row) => {
    const qtyInput = row.querySelector(".qty-input");
    const priceInput = row.querySelector(".price-input");

    const qty = parseInt(qtyInput?.value) || 0;
    const price = parseFloat(priceInput?.value) || 0;

    // Highlight invalid cases
    if (qty < 1 || price <= 0) {
      row.classList.remove("bg-green-50", "dark:bg-green-900/20");
      row.classList.add("bg-red-100", "dark:bg-red-900/30");
    } else {
      row.classList.remove("bg-red-100", "dark:bg-red-900/30");
    }

    // Disable only if quantity < 1
    if (qty < 1) {
      isValid = false;
    }
  });

  // validate amount paid vs total
  const subtotal = cart.reduce((sum, item) => sum + item.price * item.qty, 0);
  const paid = Object.values(payments).reduce((a, b) => a + b, 0);

  if (paid < subtotal) {
    isValid = false;
    document.getElementById("amount-paid").classList.add("text-red-500");
  } else {
    document.getElementById("amount-paid").classList.remove("text-red-500");
  }

  // toggle Save button
  const saveBtn = document.getElementById("saveSaleBtn");
  if (saveBtn) saveBtn.disabled = !isValid;

  const cartNotEmpty = cart.length > 0;
  const holdBtn = document.getElementById("hold-sale-btn");
  if (holdBtn) holdBtn.disabled = !cartNotEmpty;
}

// validate when inputs change or blur
document.addEventListener("input", (e) => {
  if (
    e.target.classList.contains("qty-input") ||
    e.target.classList.contains("price-input") ||
    e.target.id === "amount-paid"
  ) {
    validateSale();
  }
});

function holdSale() {
  alert("Holding current sale...");
  cart.length = 0;
  for (let key in payments) delete payments[key];
  renderCart();
}

function viewHeld() {
  document.getElementById("held-modal").classList.remove("hidden");
}

function closeHeld() {
  document.getElementById("held-modal").classList.add("hidden");
}

function buildSalePayload(cart, payments) {
  const items = cart.map((item) => {
    const subtotal = item.price * item.qty;
    const unit_price = item.price;
    const selectedPrice =
      item.price_options.find((p) => p.id === item.selected_price_id)
        ?.selling_price || unit_price;

    const discount = (selectedPrice - unit_price) * item.qty;
    const total = subtotal - discount;

    return {
      product_id: item.id,
      quantity: item.qty,
      price_id: item.selected_price_id,
      unit_price: item.price,
      discount: discount,
      total: total,
    };
  });

  const subtotal = items.reduce((sum, i) => sum + i.unit_price * i.quantity, 0);
  const discount = items.reduce((sum, i) => sum + i.discount, 0);
  const total = subtotal - discount;

  const paymentList = Object.entries(payments)
    .filter(([_, amount]) => amount > 0)
    .map(([method, amount]) => ({
      amount,
      payment_method: method,
    }));

  return {
    subtotal,
    discount,
    total,
    items,
    payments: paymentList,
  };
}

function closePaymentModal() {
    document.getElementById("payment-modal").classList.add("hidden");
}
