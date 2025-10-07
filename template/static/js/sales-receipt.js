let cart = [];
let payments = {};
let holdReference = "";
let selectedPaymentMethod = "";
let searchTimeout;

const subtotalDisplay = document.getElementById("subtotal");
const paidDisplay = document.getElementById("paid");
const changeDisplay = document.getElementById("change");
const receiptItems = document.getElementById("receipt-items");
const paymentInput = document.getElementById("payment-amount");
const searchResults = document.getElementById("search-results");
const itemSearch = document.getElementById("item-search");
const paymentModal = document.getElementById("payment-modal");

window.addEventListener("DOMContentLoaded", () => {
  if (itemSearch) {
    itemSearch.focus();
  }
});

function updateTotals() {
  let subtotal = cart.reduce((sum, item) => {
    return sum.plus(new Decimal(item.price).times(item.qty));
  }, new Decimal(0));

  let paid = Object.values(payments).reduce((sum, amount) => {
    return sum.plus(new Decimal(amount));
  }, new Decimal(0));

  // Round both subtotal & paid to whole numbers
  subtotal = subtotal.toDecimalPlaces(0, Decimal.ROUND_HALF_EVEN);
  paid = paid.toDecimalPlaces(0, Decimal.ROUND_HALF_EVEN);

  // Change = (paid - subtotal), but never below 0
  let change = Decimal.max(paid.minus(subtotal), 0).toDecimalPlaces(
    0,
    Decimal.ROUND_HALF_EVEN,
  );

  subtotalDisplay.textContent = subtotal.toNumber().toLocaleString();
  paidDisplay.textContent = paid.toNumber().toLocaleString();
  changeDisplay.textContent = change.toNumber().toLocaleString();

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
      price: new Decimal(item.default_price?.selling_price || 0),
      qty: new Decimal(1),
      default_price: item.default_price,
      price_options: item.price_options,
      selected_price_id: item.default_price?.id,
    });
  }

  renderCart();
}

function roundTo50or100(value) {
  // value is a Decimal
  const n = value.toNumber();
  const remainder = n % 50;

  if (remainder === 0) return value; // already on 50/100 boundary
  return new Decimal(Math.ceil(n / 50) * 50); // always round UP
}

function renderCart() {
  receiptItems.innerHTML = "";

  cart.forEach((item, i) => {
    const selectedOption = item.price_options?.find(
      (opt) => opt.id === item.selected_price_id,
    );

    const isDiscounted =
      selectedOption?.selling_price &&
      item.price.lt(new Decimal(selectedOption.selling_price));

    // Calculate row total with Decimal
    let rowTotal = item.price.times(item.qty);
    let roundedTotal = rowTotal
      .div(50)
      .toDecimalPlaces(0, Decimal.ROUND_HALF_UP)
      .times(50);

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
          ₦${item.price.toDecimalPlaces(2).toString()}
        </span>

        <input type="number" value="${item.price.toDecimalPlaces(2).toString()}" min="0"
          class="no-spinners price-input block hidden w-20 text-center px-2 py-1 border rounded
                  dark:bg-gray-700 dark:border-gray-600"
          data-index="${i}">
      </td>

      <td class="px-4 py-2 text-center">
        <div>
          <span class="total-display cursor-pointer" data-index="${i}">
            ₦${roundedTotal.toFixed(0)}
          </span>
          <input type="number" value="${roundedTotal.toFixed(0)}" min="0"
            class="no-spinners total-input hidden w-24 text-center px-2 py-1 border rounded dark:bg-gray-700 dark:border-gray-600"
            data-index="${i}">
        </div>

        ${
          isDiscounted
            ? `<div class="text-xs text-red-500 mt-1">
                -₦${new Decimal(item.default_price.selling_price).minus(item.price).times(item.qty).toDecimalPlaces(0, Decimal.ROUND_HALF_EVEN).toString()} discount
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

  // keep qty listeners...
  document.querySelectorAll(".qty-input").forEach((input) => {
    input.addEventListener("blur", (e) => {
      const i = +e.target.dataset.index;
      cart[i].qty = new Decimal(Math.max(1, +e.target.value || 1));
      renderCart();
      itemSearch.focus();
    });
    input.addEventListener("keydown", (e) => {
      if (e.key === "Enter") {
        e.preventDefault();
        const i = +e.target.dataset.index;
        cart[i].qty = new Decimal(Math.max(1, +e.target.value || 1));
        renderCart();
      }
    });
  });

  document.querySelectorAll(".qty-btn").forEach((btn) => {
    btn.addEventListener("click", (e) => {
      const i = +e.target.dataset.index;
      const action = e.target.dataset.action;
      if (action === "inc") cart[i].qty = cart[i].qty.plus(1);
      if (action === "dec" && cart[i].qty.gt(1))
        cart[i].qty = cart[i].qty.minus(1);
      renderCart();
    });
  });

  updateTotals();
  validateSale();
  itemSearch.focus();
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

  const newPrice = new Decimal(input.value || 0);
  cart[index].price = newPrice;

  span.textContent = `₦${newPrice.toDecimalPlaces(2).toNumber().toLocaleString()}`;
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

  const newTotal = new Decimal(input.value || 0);
  const qty = new Decimal(cart[index].qty);

  cart[index].price = qty.gt(0)
    ? newTotal.div(qty).toDecimalPlaces(2)
    : new Decimal(0);

  span.textContent = `₦${newTotal.toDecimalPlaces(2).toNumber().toLocaleString()}`;
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
    .map((opt) => {
      const sellingPrice = new Decimal(opt.selling_price);

      return `
        <button
          class="price-chip px-3 py-1 rounded-full border text-sm
            ${
              item.selected_price_id == opt.id
                ? "bg-emerald-500 text-white border-emerald-600"
                : "border-gray-300 dark:border-gray-600 hover:bg-gray-100 dark:hover:bg-gray-700"
            }"
          data-index="${index}"
          data-price="${sellingPrice.toDecimalPlaces(2).toString()}"
          data-price-id="${opt.id}">
          ${opt.name} - ₦${sellingPrice.toDecimalPlaces(2).toNumber().toLocaleString()}
        </button>
      `;
    })
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
        ${
          options ||
          `<span class="text-sm text-gray-500 dark:text-gray-400">
             No price options available.
           </span>`
        }

        <div class="ml-auto flex items-center gap-2">
          <span class="text-sm text-gray-600 dark:text-gray-300">Custom:</span>
          <input type="number" min="0" step="0.01"
            class="no-spinners price-custom w-28 px-2 py-1 border rounded
                   dark:bg-gray-700 dark:border-gray-600"
            data-index="${index}"
            value="${item.price.toDecimalPlaces(2).toNumber()}" />
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
    const price = new Decimal(chip.dataset.price || 0);
    const priceId = chip.dataset.priceId || null;

    cart[idx].price = price;
    cart[idx].selected_price_id = priceId;

    renderCart();
  }

  const apply = e.target.closest(".price-apply");
  if (apply) {
    const idx = +apply.dataset.index;
    const input = document.querySelector(`.price-custom[data-index="${idx}"]`);
    const v = new Decimal(input.value || 0);
    if (!v.isNaN()) {
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
    <div class="flex items-center justify-between">
      <!-- Left side -->
      <div class="flex flex-col">
        <span class="font-medium">${item.name}</span>
        <span class="text-sm text-gray-500 dark:text-gray-400">${item.manufacturer || ""}</span>
      </div>

      <!-- Right side -->
      <div class="ml-4 text-right">
        <span class="text-green-600 dark:text-green-400 font-semibold text-lg">
          ₦${item.default_price?.selling_price ?? ""}
        </span>
      </div>
    </div>
  `;
  li.className =
    "px-4 py-2 cursor-pointer hover:bg-gray-200 dark:hover:bg-gray-700";
  li.onclick = () => {
    addItem(item);
    hideResults();
    itemSearch.value = "";
    itemSearch.focus();
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

    activeIndex = 0;
    requestAnimationFrame(() => {
      const items = searchResults.querySelectorAll("li");
      if (!items.length) return;

      // Ensure dropdown scroll starts at top
      searchResults.scrollTop = 0;

      // update visuals - align to the top so old scroll position isn't reused
      updateActive(items, /*alignStart=*/ true);
    });
  } catch (err) {
    console.error("Search error", err);
    searchResults.classList.add("hidden");
  }
}

function hideResults() {
  searchResults.classList.add("hidden");
  activeIndex = -1;
}

function updateActive(items, alignStart = false) {
  items.forEach((el, i) => {
    if (i === activeIndex) {
      el.classList.add("bg-emerald-100", "dark:bg-emerald-700");
      // alignStart=true used when we want the first item to be at the top
      el.scrollIntoView({ block: alignStart ? "start" : "nearest" });
    } else {
      el.classList.remove("bg-emerald-100", "dark:bg-emerald-700");
    }
  });
}

itemSearch.addEventListener("input", (e) => {
  const value = e.target.value.toLowerCase();
  if (!value) {
    searchResults.innerHTML = "";
    searchResults.classList.add("hidden");
    return;
  }

  debounce(() => fetchAndRenderResults(value), 300);
  resetActiveIndex();
  searchResults.scrollTop = 0;
});

function openPayment(method) {
  selectedPaymentMethod = method;
  document.getElementById("payment-method-name").textContent =
    `Pay with ${method}`;

  const unpaid = getUnpaidAmount();

  const value = payments[method] ? payments[method] : unpaid;
  paymentInput.value = value;

  const modal = document.getElementById("payment-modal");
  modal.classList.remove("hidden");

  setTimeout(() => {
    paymentInput.focus();
    paymentInput.select();
  }, 50);
}

paymentModal.addEventListener("keydown", (e) => {
  if (e.key === "Enter") {
    e.preventDefault();
    savePayment();
  } else if (e.key === "Escape") {
    e.preventDefault();
    closePaymentModal();
  }
});

function closePaymentModal() {
  document.getElementById("payment-modal").classList.add("hidden");
  paymentInput.value = "";
}

function savePayment() {
  const amount = parseFloat(paymentInput.value || 0);
  if (amount > 0) {
    payments[selectedPaymentMethod] = amount;
  }
  paymentModal.classList.add("hidden");
  updateTotals();
}

function cancelPayment() {
  delete payments[selectedPaymentMethod];
  paymentModal.classList.add("hidden");
  updateTotals();
}

async function saveSale(printAfterSave = false) {
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
      showToast(`Sale saved successfully`, {
        type: "info",
        duration: 3000,
      });

      if (printAfterSave) {
        printReceipt(payload);
      }

      const saveBtn = document.getElementById("saveSaleBtn");
      const savePrintBtn = document.getElementById("savePrintBtn");
      saveBtn.disabled = true;
      savePrintBtn.disabled = true;

      cart = [];
      payments = {};
      holdReference = "";
      renderCart();
      validateSale();
      // window.location.reload();
    } else {
      const errorData = await response.json().catch(() => ({}));
      console.error("Save sale failed:", errorData);
      showToast(`Failed to save sale. Please try again.`, {
        type: "error",
        duration: 3000,
      });
    }
  } catch (error) {
    console.error("Error saving sale:", error);
    showToast(`An error occurred while saving the sale.`, {
      type: "error",
      duration: 3000,
    });
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

  if (cart.length === 0) {
    isValid = false;
  }

  // toggle Save button
  const saveBtn = document.getElementById("saveSaleBtn");
  const savePrintBtn = document.getElementById("savePrintBtn");
  if (saveBtn) saveBtn.disabled = !isValid;
  if (savePrintBtn) savePrintBtn.disabled = !isValid;

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

async function holdSale() {
  try {
    const payload = {
      reference: holdReference,
      payload: {
        cart,
        payments,
      },
    };

    const response = await fetch("/sales/hold", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(payload),
    });

    if (!response.ok) {
      throw new Error(`Failed to hold sale: ${response.statusText}`);
    }

    const data = await response.json();
    showToast(`Sale held successfully`, {
      type: "info",
      duration: 3000,
    });

    cart = [];
    payments = {};
    holdReference = "";

    renderCart();

    console.log("Hold response:", data);
    return data;
  } catch (error) {
    console.error("Error holding sale:", error);
    showToast(`Error holding sale`, {
      type: "error",
      duration: 3000,
    });
  }
}

// Helpers
const D = (v) => (v instanceof Decimal ? v : new Decimal(v ?? 0));
const toNaira = (d) => D(d).toDecimalPlaces(0, Decimal.ROUND_HALF_UP); // whole ₦
const nearest50 = (d) =>
  D(d).div(50).toDecimalPlaces(0, Decimal.ROUND_HALF_UP).times(50);

function buildSalePayload(cart, payments) {
  const items = cart.map((item) => {
    const qty = D(item.qty);
    const unitPrice = D(item.price);

    const selected = item.price_options?.find(
      (p) => p.id === item.selected_price_id,
    )?.selling_price;
    const listUnit = D(selected ?? unitPrice);

    const lineSubtotal = listUnit.times(qty);
    const lineTotalRaw = unitPrice.times(qty);
    const lineTotal = nearest50(lineTotalRaw);
    let lineDiscount = lineSubtotal.minus(lineTotal);
    if (lineDiscount.lt(0)) lineDiscount = new Decimal(0);

    return {
      name: item.name,
      product_id: item.id,
      quantity: toNaira(qty).toNumber(),
      price_id: parseInt(item.selected_price_id, 10),
      unit_price: unitPrice.toDecimalPlaces(2).toNumber(2),
      discount: toNaira(lineDiscount).toNumber(),
      total: toNaira(lineTotal).toNumber(),
    };
  });

  const total = items.reduce((s, it) => s.plus(D(it.total)), new Decimal(0));
  const discount = items.reduce(
    (s, it) => s.plus(D(it.discount)),
    new Decimal(0),
  );
  const subtotal = total.plus(discount);

  const paymentList = Object.entries(payments)
    .filter(([, amount]) => D(amount).gt(0))
    .map(([method, amount]) => ({
      amount: toNaira(amount).toNumber(),
      payment_method: method,
    }));

  return {
    held_sale_reference: holdReference,
    subtotal: toNaira(subtotal).toNumber(),
    discount: toNaira(discount).toNumber(),
    total: toNaira(total).toNumber(),
    items,
    payments: paymentList,
  };
}

function closePaymentModal() {
  document.getElementById("payment-modal").classList.add("hidden");
}

let activeIndex = -1; // track highlighted index
function resetActiveIndex() {
  activeIndex = -1;
}

// Handle key presses on search input
itemSearch.addEventListener("keydown", (e) => {
  const items = searchResults.querySelectorAll("li");
  if (!items.length) return;

  if (e.key === "ArrowDown") {
    e.preventDefault();
    activeIndex = (activeIndex + 1) % items.length;
    updateActive(items);
  } else if (e.key === "ArrowUp") {
    e.preventDefault();
    activeIndex = (activeIndex - 1 + items.length) % items.length;
    updateActive(items);
  } else if (e.key === "Enter" && activeIndex >= 0) {
    e.preventDefault();
    items[activeIndex].click();
    resetActiveIndex();
  } else if (e.key === "Escape") {
    hideResults();
  }
});

function getUnpaidAmount() {
  let subtotal = cart.reduce((sum, item) => {
    return sum.plus(new Decimal(item.price).times(item.qty));
  }, new Decimal(0));

  let paid = Object.values(payments).reduce((sum, amount) => {
    return sum.plus(new Decimal(amount));
  }, new Decimal(0));
  return subtotal - paid;
}

document.addEventListener("keydown", (e) => {
  switch (e.code) {
    case "F1":
      if (e.shiftKey) {
        e.preventDefault();
        openPayment("Cash");
      }
      break;
    case "F2":
      if (e.shiftKey) {
        e.preventDefault();
        openPayment("Card");
      }
      break;
    case "F3":
      if (e.shiftKey) {
        e.preventDefault();
        openPayment("Transfer");
      }
      break;
    case "F11":
      e.preventDefault();
      const unpaid = getUnpaidAmount();
      if (unpaid > 0) {
        openPayment("Cash");
      } else {
        saveSale();
      }
      break;
  }
});

document.addEventListener("DOMContentLoaded", () => {
  const heldData = localStorage.getItem("heldSale");
  if (!heldData) return;

  try {
    const data = JSON.parse(heldData);
    holdReference = data.reference;

    if (data.payload.cart && Array.isArray(data.payload.cart)) {
      cart = (data.payload.cart || []).map((i) => ({
        ...i,
        price: D(i.price),
        qty: D(i.qty),
      }));
    }

    if (
      data.payload.payments &&
      Object.keys(data.payload.payments).length > 0
    ) {
      payments = data.payload.payments;
    }

    renderCart();
    localStorage.removeItem("heldSale");

    showToast("Held sale restored successfully", "success");
  } catch (err) {
    console.error("Error restoring held sale:", err);
    showToast("Could not restore held sale", "error");
  }
});
