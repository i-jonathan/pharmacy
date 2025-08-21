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
    });
  }

  renderCart();
}

function renderCart() {
  receiptItems.innerHTML = "";

  cart.forEach((item, i) => {
    const isDiscounted =
      item.default_price?.selling_price &&
      item.price < item.default_price?.selling_price;

    const row = document.createElement("tr");
    row.className = isDiscounted ? "bg-green-50 dark:bg-green-900/20" : "";

    row.innerHTML = `
      <td class="px-4 py-2 w-[40%]">
        <span class="font-medium">${item.name}</span>
        <span class="ml-1 text-sm text-gray-500 dark:text-gray-400">- ${item.manufacturer}</span>
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

      <td class="px-4 py-2 text-center">
        <div>
          <span class="price-display cursor-pointer" data-index="${i}">₦${item.price.toFixed(2)}</span>
          <input type="number" value="${item.price.toFixed(2)}" min="0"
            class="price-input hidden w-20 text-center px-2 py-1 border rounded dark:bg-gray-700 dark:border-gray-600"
            data-index="${i}">
        </div>
      </td>

      <td class="px-4 py-2 text-center">
        <span class="total-display cursor-pointer" data-index="${i}">₦${(item.qty * item.price).toFixed(2)}</span>
        <input type="number" value="${(item.qty * item.price).toFixed(2)}" min="0"
          class="total-input hidden w-24 text-center px-2 py-1 border rounded dark:bg-gray-700 dark:border-gray-600"
          data-index="${i}">
      </td>

      <td class="px-4 py-2 text-center w-[1%]">
        <button class="text-red-500" onclick="removeItem(${i})">🗑️</button>
      </td>
    `;

    receiptItems.appendChild(row);
  });

  // Quantity input listeners
  document.querySelectorAll(".qty-input").forEach((input) => {
    input.addEventListener("input", (e) => {
      const i = +e.target.dataset.index;
      cart[i].qty = +e.target.value;
      renderCart();
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
  const span = document.querySelector(
    `.price-display[data-index="${index}"]`,
  );

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
      commitUnitPriceChange(e.target)
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
  const span = document.querySelector(
    `.total-display[data-index="${index}"]`,
  );

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
      commitTotalPriceChange(e.target)
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
  document
    .querySelectorAll("button")
    .forEach((btn) => btn.classList.remove("disabled"));
  document.getElementById("payment-modal").classList.add("hidden");
  updateTotals();
}

function saveSale() {
  alert("Saving sale to backend...");
}

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
