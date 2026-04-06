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
    const sellingPrice = parseFloat(
      row.querySelector(".selling-price")?.value || 0,
    );
    // Use cost price for subtotal calculation (what you're paying)
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
            placeholder="Enter selling price (updates default)"
            required
          />
          <span class="suggested-price text-xs text-blue-600 dark:text-blue-400 cursor-pointer">Suggested: ₦${(product.cost_price * 1.3).toFixed(2)}</span>
        </div>
      </td>
      <td class="px-4 py-2">
        <input type="number" min="1" class="qty w-full px-2 py-1 border rounded dark:bg-gray-700 dark:border-gray-600" required/>
      </td>
      <td class="px-4 py-2">
        <input type="date" class="w-full px-2 py-1 border rounded dark:bg-gray-700 dark:border-gray-600" required/>
      </td>
      <td class="px-4 py-2 text-center">
        <button onclick="removeProductRow(this)" class="text-red-500 hover:text-red-700">🗑️</button>
        <button onclick="togglePriceManagement(this)" class="ml-2 text-purple-600 dark:text-purple-400 hover:text-purple-800 dark:hover:text-purple-300" title="Manage Price Options">
            ⚙️
        </button>
      </td>
    `;

  // Add price management row (hidden by default)
  const priceManagementRow = document.createElement("tr");
  priceManagementRow.className = "price-management-row hidden";
  priceManagementRow.dataset.productId = product.id;
  priceManagementRow.innerHTML = `
      <td colspan="6" class="p-4 bg-gray-50 dark:bg-gray-800 border-t border-b border-gray-200 dark:border-gray-700">
        <div class="flex items-center justify-between mb-3">
          <h4 class="text-sm font-semibold text-gray-700 dark:text-gray-300">Price Options Management</h4>
        </div>
        <div class="price-options-list space-y-2">
          ${
            product.price_options && product.price_options.length > 0
              ? product.price_options
                  .map(
                    (price) => `
                    <div class="flex items-center gap-3 p-2 rounded border
                      ${
                        price.id === product.default_price?.id
                          ? "bg-gray-100 dark:bg-gray-800 border-gray-300 dark:border-gray-600 opacity-80"
                          : "bg-white dark:bg-gray-700 border-gray-200 dark:border-gray-600"
                      }">

                      <div class="flex flex-col">
                        <label class="text-xs text-gray-600 dark:text-gray-400 mb-1">
                          Name ${price.id === product.default_price?.id ? '<span class="text-purple-500">(Default)</span>' : ""}
                        </label>
                        <input
                          type="text"
                          value="${price.name}"
                          class="price-name flex-1 px-2 py-1 border rounded text-sm
                            ${
                              price.id === product.default_price?.id
                                ? "bg-gray-200 dark:bg-gray-600 text-gray-500 dark:text-gray-300 cursor-not-allowed border-gray-300"
                                : "bg-white dark:bg-gray-600 dark:border-gray-500"
                            }"
                          placeholder="Name"
                          data-price-id="${price.id}"
                          ${price.id === product.default_price?.id ? "disabled" : ""}
                        />
                      </div>

                      <div class="flex flex-col">
                        <label class="text-xs text-gray-600 dark:text-gray-400 mb-1">Price</label>
                        <input
                          type="number"
                          value="${price.selling_price.toFixed(2)}"
                          class="price-value w-24 px-2 py-1 border rounded text-sm
                            ${
                              price.id === product.default_price?.id
                                ? "bg-gray-200 dark:bg-gray-600 text-gray-500 dark:text-gray-300 cursor-not-allowed border-gray-300"
                                : "bg-white dark:bg-gray-600 dark:border-gray-500"
                            }"
                          placeholder="Price"
                          data-price-id="${price.id}"
                          ${price.id === product.default_price?.id ? "disabled" : ""}
                        />
                      </div>

                      <div class="flex flex-col">
                        <label class="text-xs text-gray-600 dark:text-gray-400 mb-1">Qty/Unit</label>
                        <input
                          type="number"
                          value="${price.quantity || 1}"
                          class="quantity-per-unit w-20 px-2 py-1 border rounded text-sm
                            ${
                              price.id === product.default_price?.id
                                ? "bg-gray-200 dark:bg-gray-600 text-gray-500 dark:text-gray-300 cursor-not-allowed border-gray-300"
                                : "bg-white dark:bg-gray-600 dark:border-gray-500"
                            }"
                          placeholder="Qty/Unit"
                          data-price-id="${price.id}"
                          ${price.id === product.default_price?.id ? "disabled" : ""}
                        />
                      </div>

                    </div>
            `,
                  )
                  .join("")
              : '<div class="text-sm text-gray-500 dark:text-gray-400">No price options yet. Add one below.</div>'
          }
        </div>
        <button type="button" class="add-price-option mt-3 text-sm text-green-600 dark:text-green-400 cursor-pointer hover:underline">
          + Add Price Option
        </button>
      </td>
    `;

  // Add both rows to the table
  receivingRows.appendChild(row);
  receivingRows.appendChild(priceManagementRow);

  // Handle suggested price click
  const suggestedPrice = row.querySelector(".suggested-price");
  suggestedPrice.addEventListener("click", function () {
    const suggestedPrice = parseFloat(
      this.textContent.replace("Suggested: ₦", ""),
    );
    const sellingPriceInput = row.querySelector(".selling-price");
    sellingPriceInput.value = suggestedPrice.toFixed(2);
    updateSubtotal();
  });

  // Sync main selling price with default price option
  const mainSellingPriceInput = row.querySelector(".selling-price");
  const defaultPriceInput = priceManagementRow.querySelector(
    `.price-value[data-price-id="${product.default_price?.id}"]`,
  );

  if (mainSellingPriceInput && defaultPriceInput) {
    // Sync main input to default price option
    mainSellingPriceInput.addEventListener("input", function () {
      defaultPriceInput.value = this.value;
    });

    // Sync default price option to main input (if it's not disabled)
    defaultPriceInput.addEventListener("input", function () {
      if (!this.disabled) {
        mainSellingPriceInput.value = this.value;
      }
    });
  }

  // Use setupPriceLogic for price suggestions
  setupPriceLogic(row);

  // Handle add new price option
  const addPriceBtn = priceManagementRow.querySelector(".add-price-option");
  addPriceBtn.addEventListener("click", function () {
    const priceOptionsList = priceManagementRow.querySelector(
      ".price-options-list",
    );
    const newPriceOption = document.createElement("div");
    newPriceOption.className =
      "flex items-center gap-3 p-2 bg-white dark:bg-gray-700 rounded border dark:border-gray-600";
    newPriceOption.innerHTML = `
      <div class="flex flex-col">
        <label class="text-xs text-gray-600 dark:text-gray-400 mb-1">Name</label>
        <input type="text" class="price-name flex-1 px-2 py-1 border rounded dark:bg-gray-600 dark:border-gray-500 text-sm" placeholder="Name" data-price-id="new"/>
      </div>
      <div class="flex flex-col">
        <label class="text-xs text-gray-600 dark:text-gray-400 mb-1">Price</label>
        <input type="number" class="price-value w-24 px-2 py-1 border rounded dark:bg-gray-600 dark:border-gray-500 text-sm" placeholder="Price" data-price-id="new"/>
      </div>
      <div class="flex flex-col">
        <label class="text-xs text-gray-600 dark:text-gray-400 mb-1">Qty/Unit</label>
        <input type="number" value="1" class="quantity-per-unit w-20 px-2 py-1 border rounded dark:bg-gray-600 dark:border-gray-500 text-sm" placeholder="Qty/Unit" data-price-id="new"/>
      </div>
    `;
    priceOptionsList.appendChild(newPriceOption);
  });

  row.querySelectorAll(".cost, .qty, .selling-price").forEach((input) => {
    input.addEventListener("input", updateSubtotal);
  });

  searchResults.classList.add("hidden");
}

// Global function to remove product row and associated price management row
function removeProductRow(button) {
  const currentRow = button.closest('tr');
  const priceManagementRow = currentRow.nextElementSibling;
  
  // Remove price management row if it exists and is the right one
  if (priceManagementRow && priceManagementRow.classList.contains('price-management-row')) {
    priceManagementRow.remove();
  }
  
  // Remove the main row
  currentRow.remove();
  updateSubtotal();
}

// Global function to toggle price management
function togglePriceManagement(button) {
  const currentRow = button.closest("tr");
  let priceManagementRow;

  if (button.textContent === "✕") {
    // Close button in price management row
    priceManagementRow = currentRow;
    currentRow.classList.add("hidden");
  } else {
    // Settings button in main row
    priceManagementRow = currentRow.nextElementSibling;
    if (
      priceManagementRow &&
      priceManagementRow.classList.contains("price-management-row")
    ) {
      priceManagementRow.classList.toggle("hidden");
    }
  }
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
  const rows = document.querySelectorAll(
    "#receiving-table tbody tr.receiving-row",
  );
  const data = [];

  rows.forEach((row) => {
    const cells = row.querySelectorAll("td input");
    const sellingPriceInput = row.querySelector(".selling-price");
    const productId = row.dataset.itemId;

    // Find the price management row for this product
    const priceManagementRow = document.querySelector(
      `.price-management-row[data-product-id="${productId}"]`,
    );

    // Collect price options changes
    const priceOptionsChanges = [];
    if (priceManagementRow) {
      const priceOptionDivs = priceManagementRow.querySelectorAll(
        ".price-options-list > div",
      );

      priceOptionDivs.forEach((div) => {
        const nameInput = div.querySelector(".price-name");
        const valueInput = div.querySelector(".price-value");
        const quantityInput = div.querySelector(".quantity-per-unit");
        const priceId = nameInput.dataset.priceId;

        if (nameInput && valueInput && quantityInput) {
          // Validate price option data
          const name = nameInput.value.trim();
          const price = parseFloat(valueInput.value) || 0;
          const quantity = parseInt(quantityInput.value) || 1;

          if (name && price > 0 && quantity > 0) {
            priceOptionsChanges.push({
              id: priceId === "new" ? null : parseInt(priceId),
              name: name,
              selling_price: price,
              quantity_per_unit: quantity,
            });
          }
        }
      });
    }

    data.push({
      id: productId ? parseInt(productId) : null,
      barcode: cells[0].value.trim(),
      cost_price: parseFloat(cells[1].value) || 0,
      selling_price: parseFloat(cells[2].value) || 0,
      quantity: parseInt(cells[3].value) || 0,
      expiry: cells[4].value + "T00:00:00Z" || null,
      price_options_changes: priceOptionsChanges,
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
  const rows = document.querySelectorAll(
    "#receiving-table tbody tr.receiving-row",
  );

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
