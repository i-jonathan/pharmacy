const items = { Panadol: 200, "Vitamin C": 150, Amoxicillin: 300 };
const cart = [];
const payments = {};
let selectedPaymentMethod = "";

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

function addItem(name) {
  const price = items[name] || 0;
  const existing = cart.find((i) => i.name === name);
  if (existing) {
    existing.qty++;
  } else {
    cart.push({ name, qty: 1, price });
  }
  renderCart();
}

function renderCart() {
  receiptItems.innerHTML = "";

  cart.forEach((item, i) => {
    const row = document.createElement("tr");
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

      <td class="px-4 py-2 text-center">₦${item.price}</td>
      <td class="px-4 py-2 text-center">₦${item.qty * item.price}</td>

      <td class="px-4 py-2 text-center w-[1%]">
        <button class="text-red-500" onclick="removeItem(${i})">🗑️</button>
      </td>

    `;
    receiptItems.appendChild(row); // ✅ this was missing
  });

  // ✅ attach listeners once after rendering
  document.querySelectorAll(".qty-input").forEach((input) => {
    input.addEventListener("input", (e) => {
      const i = +e.target.dataset.index;
      cart[i].qty = +e.target.value;
      renderCart();
    });
  });

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

function removeItem(index) {
  cart.splice(index, 1);
  renderCart();
}

itemSearch.addEventListener("input", (e) => {
  const value = e.target.value.toLowerCase();
  searchResults.innerHTML = "";
  if (!value) return searchResults.classList.add("hidden");
  const matches = Object.keys(items).filter((name) =>
    name.toLowerCase().includes(value),
  );
  matches.forEach((name) => {
    const li = document.createElement("li");
    li.textContent = name;
    li.className =
      "px-4 py-2 cursor-pointer hover:bg-gray-200 dark:hover:bg-gray-700";
    li.onclick = () => {
      addItem(name);
      searchResults.classList.add("hidden");
      itemSearch.value = "";
    };
    searchResults.appendChild(li);
  });
  searchResults.classList.remove("hidden");
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
