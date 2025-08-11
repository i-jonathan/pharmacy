const searchInput = document.getElementById("item-search");
const searchResults = document.getElementById("search-results");
const receivingRows = document.getElementById("receiving-rows");
const subtotalDisplay = document.getElementById("subtotal");
const addItemModal = document.getElementById("item-modal");
const feedbackModal = document.getElementById("feedback-modal");
const feedbackTitle = document.getElementById("feedback-title");
const feedbackMessage = document.getElementById("feedback-message");
const feedbackClose = document.getElementById("feedback-close");

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

function addItemToTable(name) {
  const row = document.createElement("tr");
  row.classList.add("receiving-row");
  row.innerHTML = `
      <td class="px-4 py-2">${name}</td>
      <td class="px-4 py-2"><input type="text" placeholder="e.g. 0123456789012" class="w-full px-2 py-1 border rounded dark:bg-gray-700 dark:border-gray-600" /></td>
      <td class="px-4 py-2"><input type="number" class="cost w-full px-2 py-1 border rounded dark:bg-gray-700 dark:border-gray-600" /></td>
      <td class="px-4 py-2"><input type="number" class="w-full px-2 py-1 border rounded dark:bg-gray-700 dark:border-gray-600" /></td>
      <td class="px-4 py-2"><input type="number" class="qty w-full px-2 py-1 border rounded dark:bg-gray-700 dark:border-gray-600" /></td>
      <td class="px-4 py-2"><input type="date" class="w-full px-2 py-1 border rounded dark:bg-gray-700 dark:border-gray-600" /></td>
      <td class="px-4 py-2"><input type="text" placeholder="optional" class="w-full px-2 py-1 border rounded dark:bg-gray-700 dark:border-gray-600" /></td>
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

searchInput.addEventListener("input", () => {
  const value = searchInput.value.toLowerCase();
  searchResults.innerHTML = "";
  if (!value) return searchResults.classList.add("hidden");

  const matches = items.filter((item) => item.toLowerCase().includes(value));
  if (matches.length === 0) {
    searchResults.innerHTML =
      '<li class="px-4 py-2 text-sm text-gray-500">No items found</li>';
  } else {
    matches.forEach((item) => {
      const li = document.createElement("li");
      li.textContent = item;
      li.className =
        "px-4 py-2 cursor-pointer hover:bg-gray-200 dark:hover:bg-gray-700";
      li.addEventListener("click", () => addItemToTable(item));
      searchResults.appendChild(li);
    });
  }
  searchResults.classList.remove("hidden");
});

// Modal functionality
document.getElementById("add-item-btn").onclick = () =>
  addItemModal.classList.remove("hidden");

document.getElementById("cancel-modal").onclick = () =>
  addItemModal.classList.add("hidden");

document.getElementById("save-item").onclick = async () => {
  const payload = {
    name: document.getElementById("name").value.trim(),
    manufacturer: document.getElementById("manufacturer").value.trim(),
    category_id: parseInt(document.getElementById("category-id").value) || 0,
    reorder_level:
      parseInt(document.getElementById("reorder-level").value) || 0,
  };

  try {
    const res = await fetch("/add-item", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(payload),
    });

    if (res.status === 204) {
      showFeedbackModal("Success", "Product saved successfully.", true);
    } else {
      throw new Error(`Server error: ${res.status}`);
    }

    addItemModal.classList.add("hidden");
    addItemModal
      .querySelectorAll("input, select")
      .forEach((el) => (el.value = ""));
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
