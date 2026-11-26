import { reactive, computed, ref, onMounted, onUnmounted } from "vue";
import { csrfFetch } from "../lib/csrf.js";
import { readError } from "../lib/api.js";
import { usePermissions } from "./usePermissions.js";
import { useConfirm } from "./useConfirm.js";
import { buildReceiptHtml, printReceiptHtml } from "../lib/receipt.js";

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
  const { hasPermission } = usePermissions();
  const cart = reactive([]);
  const payments = reactive({ Cash: 0, Card: 0, Transfer: 0 });
  // The customer picked for this sale, {id, name, phone}, or null for a walk-in.
  const customer = ref(null);
  // The customer on a held sale that this user may not see: not shown and not sold to, but kept
  // so holding the sale again doesn't erase it.
  const hiddenCustomer = ref(null);
  const holdReference = ref(null);
  const saleIdempotencyKey = ref(createIdempotencyKey());
  // sale:oversell lets a user complete a sale that takes stock below zero, after confirming.
  const canOversell = computed(() => hasPermission("sale:oversell"));
  const amountTendered = ref(0);

  // --- Computed ---
  // Sums of prices like 19.99 carry float error (59.97000000000001); the till works in whole kobo
  // so an exact payment is seen as exact and the figures sent agree with the server's.
  const wholeKobo = (n) => Math.round(n * 100) / 100;

  const subtotal = computed(() => {
    return wholeKobo(cart.reduce((sum, item) => sum + item.price * item.qty, 0));
  });

  const totalDiscount = computed(() => {
    return wholeKobo(cart.reduce((sum, item) => sum + (item.discount || 0), 0));
  });

  const total = computed(() => {
    const t = wholeKobo(subtotal.value - totalDiscount.value);
    return Math.max(0, t);
  });

  const amountPaid = computed(() => {
    return wholeKobo(Object.values(payments).reduce((sum, v) => sum + (Number(v) || 0), 0));
  });

  const change = computed(() => {
    return Math.max(0, wholeKobo(amountPaid.value - total.value));
  });

  const amountOwed = computed(() => {
    return Math.max(0, wholeKobo(total.value - amountPaid.value));
  });

  const cartCount = computed(() => {
    return cart.reduce((sum, item) => sum + item.qty, 0);
  });

  function buildPriceOptions(product, price, priceId) {
    const options = [];
    const seen = new Set();
    // Look up the actual name for this price option
    let name = "Base";
    let quantityPerUnit = 1;
    if (product.price_options && Array.isArray(product.price_options)) {
      const matching = product.price_options.find((o) => o.id === priceId);
      if (matching && matching.name) name = matching.name;
      if (matching?.quantity_per_unit > 0) quantityPerUnit = matching.quantity_per_unit;
    }
    // Always include the selected/default price as an option
    options.push({ id: priceId, name, price: price, selling_price: price * 100, quantityPerUnit });
    seen.add(priceId);
    // Add API-provided price options if any
    if (product.price_options && Array.isArray(product.price_options)) {
      for (const o of product.price_options) {
        const optPrice = o.selling_price / 100;
        if (!seen.has(o.id)) {
          options.push({
            id: o.id,
            name: o.name || "Variant",
            price: optPrice,
            selling_price: o.selling_price,
            quantityPerUnit: o.quantity_per_unit > 0 ? o.quantity_per_unit : 1,
          });
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
        stock: stockOf(product),
      });
    }
  }

  // The stock the product list showed, in stock units, or null when the till doesn't know it (a
  // held sale resumed before the list loaded).
  function stockOf(product) {
    const stock = Number(product?.stock);
    return product?.stock === undefined || product?.stock === null || !Number.isFinite(stock) ? null : stock;
  }

  // Copies the stock and pack sizes of a freshly loaded product list onto the cart's lines (a sale
  // held before pack sizes were kept has none).
  let lastProducts = [];
  function syncStock(products = lastProducts) {
    lastProducts = products;
    const byId = new Map(products.map((p) => [Number(p.id), p]));
    for (const item of cart) {
      const product = byId.get(Number(item.id ?? item.product_id));
      if (!product) continue;
      item.stock = stockOf(product);
      for (const option of item.priceOptions || []) {
        const listed = (product.price_options || []).find((o) => Number(o.id) === Number(option.id));
        if (listed?.quantity_per_unit > 0) option.quantityPerUnit = listed.quantity_per_unit;
      }
    }
  }

  // For each cart line, the stock shown for its product when the cart takes more than that (its
  // lines summed: a pack and singles of one product), else null. Only a hint: the server decides.
  // What the cart takes from each product, in stock units (product id → units). Lines use the same
  // fallbacks as completeSale (a held sale may carry price_id and quantity).
  function stockTaken() {
    const taking = new Map();
    for (const item of cart) {
      const id = Number(item.id ?? item.product_id);
      const priceId = Number(item.priceId ?? item.price_id);
      const option = (item.priceOptions || item.price_options || []).find((o) => Number(o.id) === priceId);
      const perUnit = Number(option?.quantityPerUnit ?? option?.quantity_per_unit) || 1;
      const qty = Math.trunc(Number(item.qty ?? item.quantity) || 0);
      taking.set(id, (taking.get(id) || 0) + qty * perUnit);
    }
    return taking;
  }

  const stockShortfalls = computed(() => {
    const taking = stockTaken();
    return cart.map((item) => {
      const stock = item.stock ?? null;
      if (stock === null) return null;
      return taking.get(Number(item.id ?? item.product_id)) > stock ? Math.max(0, stock) : null;
    });
  });

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

  // Amounts typed at the till (or kept in a held sale) are whole kobo, because the server refuses a
  // fraction of a kobo; rounding each line's discount also keeps the line totals and the sale's
  // discount adding up to the same figure. A negative amount is nothing.
  function toWholeKobo(amount) {
    return Math.max(0, wholeKobo(Number(amount) || 0));
  }

  function updateDiscount(index, discount) {
    cart[index].discount = toWholeKobo(discount);
  }

  function updatePrice(index, priceId, price) {
    cart[index].priceId = priceId;
    cart[index].price = price;
  }

  function updatePayment(method, amount) {
    payments[method] = toWholeKobo(amount);
  }

  function clearCart() {
    cart.splice(0);
    try {
      localStorage.removeItem("posState");
    } catch {
      // Clearing the active in-memory cart still succeeds without storage.
    }
    Object.keys(payments).forEach((k) => (payments[k] = 0));
    customer.value = null;
    hiddenCustomer.value = null;
    holdReference.value = null;
    saleIdempotencyKey.value = createIdempotencyKey();
    amountTendered.value = 0;
  }

  // The customer a held sale remembers: the one on screen, or one this user may not see.
  function heldCustomerId() {
    return (customer.value || hiddenCustomer.value)?.id || 0;
  }

  // --- API Methods ---
  async function holdCart() {
    if (cart.length === 0) throw new Error("Cannot hold an empty sale");

    const payload = {
      reference: holdReference.value || "",
      payload: {
        cart: cloneForStorage(cart),
        payments: { ...payments },
        // Only the id: the held-sale list is readable by every signed-in user, so no name or phone.
        customer: heldCustomerId() ? { id: heldCustomerId() } : null,
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

  // allowOversell is sent only after a refusal for stock, by a user with sale:oversell who
  // confirmed selling more than is in stock.
  async function completeSale({ allowOversell = false } = {}) {
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

    if (customer.value?.id) payload.customer_id = customer.value.id;
    if (allowOversell) payload.allow_oversell = true;
    if (holdReference.value) {
      payload.held_sale_reference = holdReference.value;
    }

    const resp = await csrfFetch(`${API_BASE}/sales/`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });
    if (!resp.ok) {
      // A refusal the cashier can act on (400, 403, 409: for example a deactivated customer, or no
      // permission to attach one) says why; a server error keeps the plain message.
      if ([400, 403, 409].includes(resp.status)) {
        const { message, code } = await readError(resp, "Failed to complete sale");
        throw Object.assign(new Error(message), { code });
      }
      throw new Error("Failed to complete sale");
    }
    return resp.json();
  }

  async function fetchHeldTransactions() {
    const resp = await fetch(`${API_BASE}/sales/api/held`);
    if (!resp.ok) throw new Error("Failed to fetch held transactions");
    return resp.json();
  }

  // printReceipt prints the cart as a receipt. Call it before the cart is cleared, with the number
  // the server gave the saved sale (the same one Sales History shows).
  async function printReceipt(receiptNumber) {
    // A resumed held sale knows only its customer's id until the name is read. Printing the
    // placeholder would put the wrong name on the receipt, so read it now or refuse to print.
    if (customer.value?.named === false) {
      await refreshCustomer(customer.value.id);
      if (customer.value?.named === false) {
        throw new Error("The customer's name could not be read, so the receipt was not printed.");
      }
    }
    return printReceiptHtml(
      buildReceiptHtml({
        receiptNumber,
        date: new Date(),
        customer: customer.value?.name || "",
        items: cart.map((item) => ({
          name: item.name,
          qty: item.qty,
          price: item.price,
          discount: item.discount || 0,
        })),
        discount: totalDiscount.value,
        total: total.value,
        payments: Object.entries(payments)
          .filter(([, amount]) => Number(amount) > 0)
          .map(([method, amount]) => ({ method, amount: Number(amount) })),
        change: change.value,
      }),
    );
  }

  async function deleteHeldTransaction(reference) {
    const resp = await csrfFetch(`${API_BASE}/sales/held/${reference}`, {
      method: "DELETE",
    });
    if (!resp.ok) throw new Error("Failed to delete held transaction");
  }

  // A held sale remembers only who the customer is; read the name and phone that are on file now.
  // A deactivated customer stays on the sale: completing it says so.
  async function refreshCustomer(id) {
    try {
      const resp = await fetch(`${API_BASE}/customers/${id}`);
      if (!resp.ok || customer.value?.id !== id) return;
      const latest = await resp.json();
      if (customer.value?.id !== id) return;
      customer.value = { id, name: latest.name, phone: latest.phone || "", active: latest.active };
    } catch {
      // The placeholder stays, and the receipt is not printed until the name is read.
    }
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
        // A sale held before amounts were rounded may carry a fraction of a kobo.
        // The stock saved with a held sale may be days old: unknown until the list is read again.
        cart.push({ ...item, discount: toWholeKobo(item.discount), stock: null });
      });
      // The product list may already be loaded.
      syncStock();
    }
    if (payload.payments) {
      Object.keys(payments).forEach((k) => {
        payments[k] = toWholeKobo(payload.payments[k]);
      });
    }
    // Older held sales kept a typed name; only a customer picked from the list is restored.
    // A user who may not see customers gets none: the sale would be refused with the customer on it.
    const held = payload.customer;
    if (held && typeof held === "object" && held.id) {
      if (hasPermission("customer:view")) {
        // An older held sale kept the name and phone; a newer one keeps only the id, and shows a
        // placeholder (named: false) until the name is read.
        customer.value = held.name
          ? { id: held.id, name: held.name, phone: held.phone || "" }
          : { id: held.id, name: "Customer", phone: "", named: false };
        refreshCustomer(held.id);
      } else {
        hiddenCustomer.value = { id: held.id };
      }
    }
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

  // F5 and F6 do what the buttons do: the page registers its handlers, which run one at a time and
  // show the server's reason when a sale is refused.
  const shortcutActions = { complete: null, hold: null };
  function setShortcutActions(actions) {
    Object.assign(shortcutActions, actions);
  }

  const { state: confirmState } = useConfirm();

  function onKeyDown(e) {
    // A dialog is waiting for an answer (for example Reactivate): the sale must not complete or be
    // held behind it. The key is still taken, so the browser does not reload on F5.
    if (confirmState.open && (e.key === "F5" || e.key === "F6")) {
      e.preventDefault();
      return;
    }
    if (e.key === "F3") {
      e.preventDefault();
      focusSearch();
    }
    if (e.key === "F5") {
      e.preventDefault();
      shortcutActions.complete?.();
    }
    if (e.key === "F6") {
      e.preventDefault();
      shortcutActions.hold?.();
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
    hiddenCustomer,
    holdReference,
    saleIdempotencyKey,
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
    canOversell,
    stockShortfalls,
    stockTaken,
    syncStock,
    printReceipt,
    fetchHeldTransactions,
    deleteHeldTransaction,
    restoreHeld,
    setSearchRef,
    setShortcutActions,
    focusSearch,
  };
}
