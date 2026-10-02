import { reactive, computed, ref, onMounted, onUnmounted } from "vue";
import { csrfFetch } from "../lib/csrf.js";

const API_BASE = "";
export function discardPersistedPosState() {
  try {
    localStorage.removeItem("posState");
    localStorage.removeItem("resumeHeldSale");
  } catch {
    // The in-memory cart is also cleared before logout navigation.
  }
}

function createIdempotencyKey() {
  try {
    return crypto.randomUUID();
  } catch {
    return `${Date.now()}-${Math.random().toString(36).slice(2)}`;
  }
}

function cloneForStorage(obj) {
  return JSON.parse(JSON.stringify(obj));
}

export function usePos() {
  const cart = reactive([]);
  const payments = reactive({ Cash: 0, Card: 0, Transfer: 0 });
  const customer = ref("Walk-in Customer");
  const holdReference = ref(null);
  const saleIdempotencyKey = ref(createIdempotencyKey());
  const selectedPaymentMethod = ref("Cash");
  const amountTendered = ref(0);

  // --- Computed ---
  const subtotal = computed(() => {
    return cart.reduce((sum, item) => sum + item.price * item.qty, 0);
  });

  const totalDiscount = computed(() => {
    return cart.reduce((sum, item) => sum + (item.discount || 0), 0);
  });

  const total = computed(() => {
    const t = subtotal.value - totalDiscount.value;
    return Math.max(0, t);
  });

  const amountPaid = computed(() => {
    return Object.values(payments).reduce((sum, v) => sum + (Number(v) || 0), 0);
  });

  const change = computed(() => {
    return Math.max(0, amountPaid.value - total.value);
  });

  const amountOwed = computed(() => {
    return Math.max(0, total.value - amountPaid.value);
  });

  const cartCount = computed(() => {
    return cart.reduce((sum, item) => sum + item.qty, 0);
  });

  function buildPriceOptions(product, price, priceId) {
    const options = [];
    const seen = new Set();
    // Look up the actual name for this price option
    let name = "Base";
    if (product.price_options && Array.isArray(product.price_options)) {
      const matching = product.price_options.find((o) => o.id === priceId);
      if (matching && matching.name) name = matching.name;
    }
    // Always include the selected/default price as an option
    options.push({ id: priceId, name, price: price, selling_price: price * 100 });
    seen.add(priceId);
    // Add API-provided price options if any
    if (product.price_options && Array.isArray(product.price_options)) {
      for (const o of product.price_options) {
        const optPrice = o.selling_price / 100;
        if (!seen.has(o.id)) {
          options.push({ id: o.id, name: o.name || "Variant", price: optPrice, selling_price: o.selling_price });
          seen.add(o.id);
        }
      }
    }
    return options;
  }

  // --- Cart Methods ---
  function addItem(product, priceId, price) {
    const productOptions = product.priceOptions || product.price_options || [];
    const defaultOption = productOptions.find((option) => Number(option.id) === Number(product.default_price_id))
      || productOptions.find((option) => Number.isInteger(Number(option.id)) && Number(option.id) > 0);
    const pid = Number(priceId || product.priceId || product.default_price_id || defaultOption?.id) || 0;
    const pprice = price ?? product.price ?? (Number(defaultOption?.selling_price) / 100 || 0);
    const existing = cart.find(
      (item) => item.id === product.id && item.priceId === pid
    );
    if (existing) {
      existing.qty += 1;
    } else {
      cart.push({
        id: product.id,
        name: product.name,
        manufacturer: product.manufacturer || "",
        price: pprice,
        priceId: pid,
        priceOptions: buildPriceOptions(product, pprice, pid),
        qty: 1,
        discount: 0,
      });
    }
  }

  function removeItem(index) {
    cart.splice(index, 1);
  }

  function updateQty(index, qty) {
    qty = Math.trunc(Number(qty) || 0);
    if (qty <= 0) {
      cart.splice(index, 1);
    } else {
      cart[index].qty = qty;
    }
  }

  function updateDiscount(index, discount) {
    cart[index].discount = Number(discount) || 0;
  }

  function updatePrice(index, priceId, price) {
    cart[index].priceId = priceId;
    cart[index].price = price;
  }

  function updatePayment(method, amount) {
    payments[method] = Number(amount) || 0;
  }

  function clearCart() {
    cart.splice(0);
    try {
      localStorage.removeItem("posState");
    } catch {
      // Clearing the active in-memory cart still succeeds without storage.
    }
    Object.keys(payments).forEach((k) => (payments[k] = 0));
    customer.value = "Walk-in Customer";
    holdReference.value = null;
    saleIdempotencyKey.value = createIdempotencyKey();
    amountTendered.value = 0;
  }

  // --- API Methods ---
  async function holdCart() {
    if (cart.length === 0) throw new Error("Cannot hold an empty sale");

    const payload = {
      reference: holdReference.value || "",
      payload: {
        cart: cloneForStorage(cart),
        payments: { ...payments },
        customer: customer.value,
        saleIdempotencyKey: saleIdempotencyKey.value,
      },
    };

    const resp = await csrfFetch(`${API_BASE}/sales/hold`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });
    if (!resp.ok) throw new Error("Failed to hold sale");
    clearCart();
    return resp.json();
  }

  async function completeSale() {
    const hasMissingPriceID = cart.some((item) => {
      const options = item.priceOptions || item.price_options || [];
      const itemPriceID = Number(item.priceId ?? item.price_id);
      return !(Number.isInteger(itemPriceID) && itemPriceID > 0)
        && !options.some((option) => Number.isInteger(Number(option.id)) && Number(option.id) > 0);
    });
    let defaultPriceIDs = new Map();
    if (hasMissingPriceID) {
      const inventoryResponse = await fetch(`${API_BASE}/inventory/item-list`);
      if (!inventoryResponse.ok) throw new Error("Could not verify product prices. Refresh the POS and try again.");
      const inventory = await inventoryResponse.json();
      defaultPriceIDs = new Map((inventory.items || []).map((product) => [
        Number(product.id),
        Number(product.default_price_id),
      ]));
    }

    const items = cart.map((item) => {
      const options = item.priceOptions || item.price_options || [];
      const validOptions = options.filter((option) => Number.isInteger(Number(option.id)) && Number(option.id) > 0);
      const selectedOption = validOptions.find((option) => Number(option.id) === Number(item.priceId ?? item.price_id))
        || validOptions.find((option) => Number(option.price ?? Number(option.selling_price) / 100) === Number(item.price ?? item.unit_price))
        || validOptions[0];
      const currentPriceId = Number(item.priceId ?? item.price_id);
      const priceId = Number.isInteger(currentPriceId) && currentPriceId > 0
        ? currentPriceId
        : Number(selectedOption?.id) || defaultPriceIDs.get(Number(item.id ?? item.product_id));
      if (!Number.isInteger(priceId) || priceId <= 0) {
        throw new Error(`Select a valid price for ${item.name} before completing the sale.`);
      }

      const quantity = Math.trunc(Number(item.qty ?? item.quantity) || 0);
      const unitPrice = Number(item.price ?? item.unit_price) || 0;
      const discount = Number(item.discount) || 0;
      return {
        product_id: Number(item.id ?? item.product_id),
        quantity,
        price_id: priceId,
        unit_price: unitPrice,
        discount,
        total: unitPrice * quantity - discount,
      };
    });
    const payload = {
      idempotency_key: saleIdempotencyKey.value,
      subtotal: Number(subtotal.value) || 0,
      discount: Number(totalDiscount.value) || 0,
      total: Number(total.value) || 0,
      items,
      payments: Object.entries(payments)
        .filter(([, amount]) => Number(amount) > 0)
        .map(([method, amount]) => ({
          payment_method: method,
          amount: Number(amount),
        })),
    };

    if (holdReference.value) {
      payload.held_sale_reference = holdReference.value;
    }

    const resp = await csrfFetch(`${API_BASE}/sales/`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });
    if (!resp.ok) throw new Error("Failed to complete sale");
    return resp.json();
  }

  async function fetchHeldTransactions() {
    const resp = await fetch(`${API_BASE}/sales/api/held`);
    if (!resp.ok) throw new Error("Failed to fetch held transactions");
    return resp.json();
  }

  function printReceipt() {
    const now = new Date();
    const receiptNumber = "POS-" + now.getTime().toString(36).toUpperCase();
    const itemsHtml = cart
      .map(
        (item) => {
          const lineTotal = (item.price * item.qty) - (item.discount || 0);
          const disc = item.discount ? `<td style="text-align:right">${item.discount.toLocaleString()}</td>` : "<td></td>";
          return `
          <tr>
            <td>${item.name}</td>
            <td style="text-align:center">${item.qty}</td>
            <td style="text-align:right">${item.price.toLocaleString()}</td>
            ${disc}
            <td style="text-align:right">${lineTotal.toLocaleString()}</td>
          </tr>`;
        }
      )
      .join("");

    const hasDiscount = totalDiscount.value > 0;
    const html = `
      <html>
      <head><title>Receipt ${receiptNumber}</title></head>
      <body style="font-family:monospace;max-width:300px;margin:0 auto;padding:10px;">
        <h3 style="text-align:center">Primocrest Pharmacy</h3>
        <p style="text-align:center;font-size:12px;">${now.toLocaleString()}</p>
        <p style="text-align:center">Receipt: ${receiptNumber}</p>
        <hr/>
        <table style="width:100%;font-size:13px;">
          <tr><th style="text-align:left">Item</th><th>Qty</th><th style="text-align:right">Price</th><th style="text-align:right">Disc</th><th style="text-align:right">Total</th></tr>
          ${itemsHtml}
        </table>
        <hr/>
        ${hasDiscount ? `<p style="text-align:right;font-size:11px;">Discount: &#8358;${totalDiscount.value.toLocaleString()}</p>` : ""}
        <p style="text-align:right;font-weight:bold">Total: &#8358;${total.value.toLocaleString()}</p>
        <p style="text-align:right">Paid: &#8358;${amountPaid.value.toLocaleString()}</p>
        <p style="text-align:right">Change: &#8358;${change.value.toLocaleString()}</p>
        <p style="text-align:center;font-size:11px;">${customer.value}</p>
        <script>window.onload=function(){window.print();window.close();}</` + `script>
      </body>
      </html>`;

    const w = window.open("", "_blank");
    w.document.write(html);
    w.document.close();
  }

  async function deleteHeldTransaction(reference) {
    const resp = await csrfFetch(`${API_BASE}/sales/held/${reference}`, {
      method: "DELETE",
    });
    if (!resp.ok) throw new Error("Failed to delete held transaction");
  }

  function restoreHeld(transaction) {
    let payload;
    try {
      payload = typeof transaction.payload === "string"
        ? JSON.parse(transaction.payload)
        : transaction.payload;
    } catch {
      return;
    }

    clearCart();

    if (payload.cart && Array.isArray(payload.cart)) {
      payload.cart.forEach((item) => {
        cart.push({ ...item });
      });
    }
    if (payload.payments) {
      Object.keys(payments).forEach((k) => {
        payments[k] = payload.payments[k] || 0;
      });
    }
    if (payload.customer) customer.value = payload.customer;
    if (payload.saleIdempotencyKey) saleIdempotencyKey.value = payload.saleIdempotencyKey;

    holdReference.value = transaction.reference;
  }

  // --- Keyboard Shortcuts ---
  let searchInputRef = null;

  function setSearchRef(el) {
    searchInputRef = el;
  }

  function focusSearch() {
    requestAnimationFrame(() => searchInputRef?.focus());
  }

  function onKeyDown(e) {
    if (e.key === "F3") {
      e.preventDefault();
      focusSearch();
    }
    if (e.key === "F5") {
      e.preventDefault();
      if (cart.length > 0) completeSale().then(() => {
        clearCart();
        focusSearch();
      });
    }
    if (e.key === "F6") {
      e.preventDefault();
      if (cart.length > 0) holdCart();
    }
  }

  function warnBeforeUnload(event) {
    if (cart.length === 0) return;
    event.preventDefault();
    event.returnValue = "";
  }

  function clearUnheldCartOnPageHide() {
    if (cart.length > 0) clearCart();
  }

  onMounted(() => {
    window.addEventListener("keydown", onKeyDown);
    window.addEventListener("beforeunload", warnBeforeUnload);
    window.addEventListener("pagehide", clearUnheldCartOnPageHide);
    // Check for a held sale being resumed from the Held Sales page
    try {
      const raw = localStorage.getItem("resumeHeldSale");
      if (raw) {
        const tx = JSON.parse(raw);
        restoreHeld(tx);
        localStorage.removeItem("resumeHeldSale");
      }
    } catch {
      localStorage.removeItem("resumeHeldSale");
    }
    // Discard drafts persisted by older builds; unfinished sales must be held explicitly.
    try {
      localStorage.removeItem("posState");
    } catch {
      // Storage may be unavailable; no draft is restored by this build.
    }
  });

  onUnmounted(() => {
    window.removeEventListener("keydown", onKeyDown);
    window.removeEventListener("beforeunload", warnBeforeUnload);
    window.removeEventListener("pagehide", clearUnheldCartOnPageHide);
  });

  return {
    cart,
    payments,
    customer,
    holdReference,
    saleIdempotencyKey,
    selectedPaymentMethod,
    amountTendered,
    subtotal,
    totalDiscount,
    total,
    amountPaid,
    change,
    amountOwed,
    cartCount,
    addItem,
    removeItem,
    updateQty,
    updateDiscount,
    updatePrice,
    updatePayment,
    clearCart,
    holdCart,
    completeSale,
    printReceipt,
    fetchHeldTransactions,
    deleteHeldTransaction,
    restoreHeld,
    setSearchRef,
    focusSearch,
  };
}
