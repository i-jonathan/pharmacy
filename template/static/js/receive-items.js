const searchInput = document.getElementById("item-search");
const searchResults = document.getElementById("search-results");
const receivingRows = document.getElementById("receiving-rows");
const subtotalDisplay = document.getElementById("subtotal");
const addItemModal = document.getElementById("item-modal");
const feedbackModal = document.getElementById("feedback-modal");
const feedbackTitle = document.getElementById("feedback-title");
const feedbackMessage = document.getElementById("feedback-message");
const feedbackClose = document.getElementById("feedback-close");
const addItemForm = document.getElementById("add-item-form");
const receiveButton = document.getElementById("receive-button");
const supplierInput = document.getElementById("supplier-input");
const supplierResults = document.getElementById("supplier-results");
const clearSupplierBtn = document.getElementById("clear-supplier");
let supplierTimeout = null;

const greaterThanZero = (val) => val.trim() !== "" && parseFloat(val) > 0;
const nonEmpty = (val) => val.trim() !== "";

function updateSubtotal() {
  let total = 0;
  document.querySelectorAll(".receiving-row").forEach((row) => {
    const cost = parseFloat(row.querySelector(".cost")?.value || 0);
    const qty = parseInt(row.querySelector(".qty")?.value || 0);
    total += cost * qty;
  });
  subtotalDisplay.textContent = total.toLocaleString();
}

function addItemToTable(product) {
  const row = document.createElement("tr");
  row.classList.add("receiving-row");
  row.dataset.itemId = product.id;

  row.innerHTML = `
      <td class="px-4 py-2">${product.name}
        <br>
        <span class="block text-sm font-medium text-gray-500 dark:text-gray-400">
            ${product.manufacturer}
        </span>
      </td>
      <td class="px-4 py-2">
        <input type="text" value="${product.barcode || ""}" placeholder="e.g. 0123456789012"
          class="w-full px-2 py-1 border rounded dark:bg-gray-700 dark:border-gray-600" />
      </td>
      <td class="px-4 py-2">
        <input type="number" step="0.01" min="1.00" value="${product.cost_price != null ? product.cost_price.toFixed(2) : "0.00"}" class="cost w-full px-2 py-1 border rounded dark:bg-gray-700 dark:border-gray-600" required/>
      </td>
      <td class="px-4 py-2">
        <div class="flex flex-col">
          <input
            type="number"
            step="0.01"
            min="1.00"
            value="${product.default_price?.selling_price != null ? product.default_price?.selling_price.toFixed(2) : "0.00"}"
            class="selling-price w-full px-2 py-1 border rounded dark:bg-gray-700 dark:border-gray-600"
            required
          />
          <span class="suggested-price text-xs text-blue-600 dark:text-blue-400 cursor-pointer hidden"></span>
        </div>
      </td>
      <td class="px-4 py-2">
        <input type="number" min="1" class="qty w-full px-2 py-1 border rounded dark:bg-gray-700 dark:border-gray-600" required/>
      </td>
      <td class="px-4 py-2">
        <input type="date" class="w-full px-2 py-1 border rounded dark:bg-gray-700 dark:border-gray-600" required/>
      </td>
      <td class="px-4 py-2 text-center">
        <button onclick="this.closest('tr').remove(); updateSubtotal();" class="text-red-500 hover:text-red-700">🗑️</button>
      </td>
    `;

  row.querySelectorAll(".cost, .qty").forEach((input) => {
    input.addEventListener("input", updateSubtotal);
  });

  receivingRows.appendChild(row);
  setupPriceLogic(row);
  searchResults.classList.add("hidden");
}

let debounceTimer;

searchInput.addEventListener("input", () => {
  clearTimeout(debounceTimer);
  debounceTimer = setTimeout(runProductSearch, 300);
});

async function runProductSearch() {
  const value = searchInput.value.toLowerCase().trim();
  searchResults.innerHTML = "";
  if (!value) return searchResults.classList.add("hidden");

  try {
    const res = await fetch(
      `/inventory/search?query=${encodeURIComponent(value)}`,
    );
    if (!res.ok) {
      throw new Error(
        `Search error: ${res.status}, ${JSON.stringify(res.json())}`,
      );
    }

    const data = await res.json();
    if (data.length === 0) {
      searchResults.innerHTML =
        '<li class="px-4 py-2 text-sm text-gray-500">No items found</li>';
      searchResults.classList.remove("hidden");
      return;
    }
    data.forEach((item) => {
      const li = document.createElement("li");
      li.innerHTML = `${item.name}<span class="ml-1 text-sm text-gray-500 dark:text-gray-400">&bull; ${item.manufacturer || ""}</span>`;

      li.className =
        "px-4 py-2 cursor-pointer hover:bg-gray-200 dark:hover:bg-gray-700";
      li.addEventListener("click", () => {
        const alreadyExists = document.querySelector(
          `.receiving-row[data-item-id="${item.id}"]`,
        );

        if (alreadyExists) {
          searchInput.value = "";
          searchResults.classList.add("hidden");
          showToast(`${item.name} is already in the table.`, {
            type: "info",
            duration: 3000,
          });
          return;
        }

        addItemToTable(item);
        searchInput.value = "";
        searchInput.focus();
      });
      searchResults.appendChild(li);
    });
  } catch (err) {
    console.error("search for product failed: ", err);
    showFeedbackModal("Error", "Search Failed", false);
  }
  searchResults.classList.remove("hidden");
}

// Modal functionality
document.getElementById("add-item-btn").onclick = () =>
  addItemModal.classList.remove("hidden");

document.getElementById("cancel-modal").onclick = () =>
  addItemModal.classList.add("hidden");

document.getElementById("save-item").onclick = async () => {
  let isValid = true;

  addItemForm.querySelectorAll("[required]").forEach((input) => {
    const errorText = input.nextElementSibling;
    if (input.value.trim() === "") {
      input.classList.add("border-red-500", "dark:border-red-500");
      if (errorText) errorText.classList.remove("hidden");
      isValid = false;
    } else {
      input.classList.remove("border-red-500", "dark:border-red-500");
      if (errorText) errorText.classList.add("hidden");
    }
  });

  if (!isValid) return; // stop execution here

  const payload = {
    name: document.getElementById("name").value.trim(),
    manufacturer: document.getElementById("manufacturer").value.trim(),
    category_id: parseInt(document.getElementById("category-id").value) || 0,
    barcode: document.getElementById("barcode").value.trim(),
    reorder_level:
      parseInt(document.getElementById("reorder-level").value) || 0,
  };

  try {
    const res = await fetch("/inventory/add-item", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(payload),
    });

    if (!res.ok) {
      throw new Error(`Server error: ${res.status}`);
    }

    // showFeedbackModal("Success", "Product saved successfully.", true);
    const product = await res.json();
    addItemToTable(product);

    addItemModal.classList.add("hidden");
    addItemForm.reset();
    addItemForm.classList.remove("validate");
  } catch (err) {
    console.error("failed to save product: ", err);
    showFeedbackModal(
      "Error",
      "Failed to save product. Please try again.",
      false,
    );
  }
};

function showFeedbackModal(title, message, isSuccess = true) {
  feedbackTitle.textContent = title;
  feedbackMessage.textContent = message;

  // color title based on success/failure
  feedbackTitle.classList.remove("text-green-600", "text-red-600");
  feedbackTitle.classList.add(isSuccess ? "text-green-600" : "text-red-600");

  feedbackModal.classList.remove("hidden");
}

feedbackClose.addEventListener("click", () => {
  feedbackModal.classList.add("hidden");
});

addItemForm.querySelectorAll("[required]").forEach((input) => {
  input.addEventListener("input", () => {
    const errorText = input.nextElementSibling;
    if (input.value.trim() !== "") {
      input.classList.remove("border-red-500", "dark:border-red-500");
      if (errorText) errorText.classList.add("hidden");
    }
  });

  if (input.tagName === "SELECT") {
    input.addEventListener("change", () => {
      const errorText = input.nextElementSibling;
      if (input.value.trim() !== "") {
        input.classList.remove("border-red-500", "dark:border-red-500");
        if (errorText) errorText.classList.add("hidden");
      }
    });
  }
});

function collectTableData() {
  const rows = document.querySelectorAll("#receiving-table tbody tr");
  const data = [];

  rows.forEach((row) => {
    const cells = row.querySelectorAll("td input");

    data.push({
      id: row.dataset.itemId ? parseInt(row.dataset.itemId) : null,
      barcode: cells[0].value.trim(),
      cost_price: parseFloat(cells[1].value) || 0,
      selling_price: parseFloat(cells[2].value) || 0,
      quantity: parseInt(cells[3].value) || 0,
      expiry: cells[4].value + "T00:00:00Z" || null,
    });
  });

  return data;
}

function attachRealtimeValidation(input, checkFn) {
  input.addEventListener(
    "input",
    () => {
      if (checkFn(input.value)) {
        input.classList.remove("border-red-500", "dark:border-red-500");
      }
    },
    { once: true },
  );
}

function validateReceiveItems() {
  const supplierInput = document.getElementById("supplier-input");
  let allValid = true;

  // Supplier validation
  supplierInput.classList.remove("border-red-500", "dark:border-red-500");
  if (!nonEmpty(supplierInput.value)) {
    supplierInput.classList.add("border-red-500", "dark:border-red-500");
    attachRealtimeValidation(supplierInput, nonEmpty);
    allValid = false;
  }
  const rows = document.querySelectorAll("#receiving-table tbody tr");

  rows.forEach((row) => {
    const inputs = row.querySelectorAll("td input");
    const costInput = inputs[1];
    const sellingInput = inputs[2];
    const qtyInput = inputs[3];
    const expiryInput = inputs[4];

    // Reset borders first
    [costInput, sellingInput, qtyInput, expiryInput].forEach((input) => {
      input.classList.remove("border-red-500", "dark:border-red-500");
    });

    // Validation checks
    if (!greaterThanZero(costInput.value)) {
      costInput.classList.add("border-red-500", "dark:border-red-500");
      attachRealtimeValidation(costInput, greaterThanZero);
      allValid = false;
    }
    if (!greaterThanZero(sellingInput.value)) {
      sellingInput.classList.add("border-red-500", "dark:border-red-500");
      attachRealtimeValidation(sellingInput, greaterThanZero);
      allValid = false;
    }
    if (!greaterThanZero(qtyInput.value)) {
      qtyInput.classList.add("border-red-500", "dark:border-red-500");
      attachRealtimeValidation(qtyInput, greaterThanZero);
      allValid = false;
    }
    if (!nonEmpty(expiryInput.value)) {
      expiryInput.classList.add("border-red-500", "dark:border-red-500");
      attachRealtimeValidation(expiryInput, nonEmpty);
      allValid = false;
    }
  });
  return allValid;
}

async function processReceiveItems() {
  const allValid = validateReceiveItems();

  if (!allValid) {
    return;
  }

  const payload = {
    supplier: supplierInput.value.trim(),
    products: collectTableData(),
  };

  try {
    const res = await postReceiveSupply(payload);
    addItemForm.reset();
    clearSupplierInput();
    receivingRows.innerHTML = "";
    updateSubtotal();

    showFeedbackModal("Saved!", `${res.message}`, true);
  } catch (err) {
    console.error(err);
    showFeedbackModal(
      "Failed to Saved",
      "There was an error saving the supply.",
      false,
    );
  }
}

receiveButton.addEventListener("click", processReceiveItems);

async function postReceiveSupply(payload) {
  try {
    const res = await fetch("/inventory/receive-items", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(payload),
    });

    if (!res.ok) {
      throw new Error(`HTTP error! Status: ${res.status}`);
    }

    return await res.json();
  } catch (err) {
    console.error("Error posting supply:", err);
    throw err;
  }
}

supplierInput.addEventListener("input", () => {
  clearTimeout(supplierTimeout);
  const value = supplierInput.value.trim();
  supplierResults.innerHTML = "";

  if (!value) {
    supplierResults.classList.add("hidden");
    return;
  }

  supplierTimeout = setTimeout(supplierSearch(value), 300);
});

async function supplierSearch(value) {
  try {
    const res = await fetch(
      `/inventory/suppliers/search?query=${encodeURIComponent(value)}`,
    );
    if (!res.ok) throw new Error("Failed to search suppliers");

    const { data } = await res.json();

    if (!data || data.length === 0) {
      supplierResults.innerHTML = `<li class="px-4 py-2 text-gray-500 text-sm">No suppliers found</li>`;
    } else {
      data.forEach((supplier) => {
        const li = document.createElement("li");
        li.textContent = supplier;
        li.className =
          "px-4 py-2 cursor-pointer hover:bg-gray-200 dark:hover:bg-gray-700";
        li.addEventListener("click", () => {
          supplierInput.value = supplier;
          supplierResults.classList.add("hidden");

          // Disable and show clear button
          supplierInput.disabled = true;
          clearSupplierBtn.classList.remove("hidden");
        });
        supplierResults.appendChild(li);
      });
    }
    supplierResults.classList.remove("hidden");
  } catch (err) {
    console.error("Supplier search failed:", err);
  }
}

// Clear supplier selection
clearSupplierBtn.addEventListener("click", () => {
  clearSupplierInput();
  supplierInput.focus();
});

function clearSupplierInput() {
  supplierInput.value = "";
  supplierInput.disabled = false;
  clearSupplierBtn.classList.add("hidden");
}

// Hide dropdown on outside click
document.addEventListener("click", (e) => {
  if (!supplierResults.contains(e.target) && e.target !== supplierInput) {
    supplierResults.classList.add("hidden");
  }
});

function setupPriceLogic(row) {
  const costInput = row.querySelector(".cost");
  const sellingInput = row.querySelector(".selling-price");
  const suggestionSpan = row.querySelector(".suggested-price");

  function updateSuggestion() {
    const cost = parseFloat(costInput.value) || 0;
    if (!cost) {
      suggestionSpan.classList.add("hidden");
      return;
    }

    // calculate suggested price: cost * 1.3, round up to nearest 50
    let suggested = Math.ceil((cost * 1.3) / 50) * 50;

    suggestionSpan.textContent = `Suggested: ₦${suggested}`;
    suggestionSpan.classList.remove("hidden");

    // when clicked, populate selling input
    suggestionSpan.onclick = () => {
      sellingInput.value = suggested.toFixed(2);
      suggestionSpan.classList.add("hidden");
    };
  }

  // update whenever selling price changes or cost changes
  sellingInput.addEventListener("input", updateSuggestion);
  costInput.addEventListener("input", updateSuggestion);

  updateSuggestion(); // run once initially
}

window.addEventListener("beforeunload", (e) => {
  const hasRows = receivingRows && receivingRows.children.length > 0;

  if (hasRows) {
    e.preventDefault();
    return "";
  }
});
