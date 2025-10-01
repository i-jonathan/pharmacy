const addModal = document.getElementById("add-modal");
const editModal = document.getElementById("edit-modal");

document.getElementById("add-btn").addEventListener("click", () => {
  addModal.classList.remove("hidden");
});

function closeAddModal() {
  addModal.classList.add("hidden");
}

function openEditModal(product) {
  // Example: product = { id: 1, name: "Paracetamol", category_id: 2 }

  // set the category dropdown
  const categorySelect = document.getElementById("edit-category");
  categorySelect.value = product.category_id;

  // set other fields
  document.getElementById("edit-name").value = product.name;
  document.getElementById("edit-stock").value = product.stock;
  document.getElementById("edit-manufacturer").value = product.manufacturer;
  document.getElementById("edit-reorder-level").value = product.reorder_level;

  // show the modal
  document.getElementById("edit-modal").classList.remove("hidden");
}

function closeEditModal() {
  editModal.classList.add("hidden");
}

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
