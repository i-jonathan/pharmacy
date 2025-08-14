const searchInput = document.getElementById("item-search");
const searchResults = document.getElementById("search-results");
const receivingRows = document.getElementById("receiving-rows");
const subtotalDisplay = document.getElementById("subtotal");
const addItemModal = document.getElementById("item-modal");
const feedbackModal = document.getElementById("feedback-modal");
const feedbackTitle = document.getElementById("feedback-title");
const feedbackMessage = document.getElementById("feedback-message");
const feedbackClose = document.getElementById("feedback-close");
const form = document.getElementById("add-item-form");

let items = ["Paracetamol 500mg", "Amoxicillin 250mg", "Ibuprofen 200mg"];

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
      <td class="px-4 py-2">${product.name}</td>
      <td class="px-4 py-2">
        <input type="text" value="${product.barcode || ""}" placeholder="e.g. 0123456789012"
          class="w-full px-2 py-1 border rounded dark:bg-gray-700 dark:border-gray-600" />
      </td>
      <td class="px-4 py-2">
        <input type="number" step="0.01" min="1.00" value="${product.cost_price != null ? product.cost_price.toFixed(2) : "0.00"}" class="cost w-full px-2 py-1 border rounded dark:bg-gray-700 dark:border-gray-600" required/>
      </td>
      <td class="px-4 py-2">
        <input type="number" step="0.01" min="1.00" value="${product.default_price?.selling_price != null ? product.default_price?.selling_price.toFixed(2) : "0.00"}" class="w-full px-2 py-1 border rounded dark:bg-gray-700 dark:border-gray-600" required/>
      </td>
      <td class="px-4 py-2">
        <input type="number" min="1" class="qty w-full px-2 py-1 border rounded dark:bg-gray-700 dark:border-gray-600" required/>
      </td>
      <td class="px-4 py-2">
        <input type="date" class="w-full px-2 py-1 border rounded dark:bg-gray-700 dark:border-gray-600" required/>
      </td>
      <td class="px-4 py-2">
        <input type="text" placeholder="optional" class="w-full px-2 py-1 border rounded dark:bg-gray-700 dark:border-gray-600" />
      </td>
      <td class="px-4 py-2 text-center">
        <button onclick="this.closest('tr').remove(); updateSubtotal();" class="text-red-500 hover:text-red-700">🗑️</button>
      </td>
    `;

  row.querySelectorAll(".cost, .qty").forEach((input) => {
    input.addEventListener("input", updateSubtotal);
  });

  receivingRows.appendChild(row);
  searchInput.value = "";
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
      li.textContent = item.name;
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

  form.querySelectorAll("[required]").forEach((input) => {
    const errorText = input.nextElementSibling;
    if (input.value.trim() === "") {
      input.classList.add("border-red-500");
      if (errorText) errorText.classList.remove("hidden");
      isValid = false;
    } else {
      input.classList.remove("border-red-500");
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
    form.reset();
    form.classList.remove("validate");
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

form.querySelectorAll("[required]").forEach((input) => {
  input.addEventListener("input", () => {
    const errorText = input.nextElementSibling;
    if (input.value.trim() !== "") {
      input.classList.remove("border-red-500");
      if (errorText) errorText.classList.add("hidden");
    }
  });

  if (input.tagName === "SELECT") {
    input.addEventListener("change", () => {
      const errorText = input.nextElementSibling;
      if (input.value.trim() !== "") {
        input.classList.remove("border-red-500");
        if (errorText) errorText.classList.add("hidden");
      }
    });
  }
});
