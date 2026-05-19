import { reactive, computed, ref, onMounted, onUnmounted } from "vue";

const API_BASE = "";

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
  const orderNote = ref("");
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

  // --- Cart Methods ---
  function addItem(product, priceId, price) {
    const pid = priceId || product.priceId || 0;
    const pprice = price ?? product.price ?? 0;
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
        qty: 1,
        discount: 0,
      });
    }
  }

  function removeItem(index) {
    cart.splice(index, 1);
  }

  function updateQty(index, qty) {
    if (qty <= 0) {
      cart.splice(index, 1);
    } else {
      cart[index].qty = qty;
    }
  }

  function updateDiscount(index, discount) {
    cart[index].discount = Number(discount) || 0;
  }

  function updatePayment(method, amount) {
    payments[method] = Number(amount) || 0;
  }

  function clearCart() {
    cart.splice(0);
    Object.keys(payments).forEach((k) => (payments[k] = 0));
    customer.value = "Walk-in Customer";
    orderNote.value = "";
    holdReference.value = null;
    saleIdempotencyKey.value = createIdempotencyKey();
    amountTendered.value = 0;
  }

  // --- API Methods ---
  async function holdCart() {
    const payload = {
      reference: holdReference.value || "",
      payload: JSON.stringify({
        cart: cloneForStorage(cart),
        payments: { ...payments },
        customer: customer.value,
        orderNote: orderNote.value,
        saleIdempotencyKey: saleIdempotencyKey.value,
      }),
    };

    const resp = await fetch(`${API_BASE}/sales/hold`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });
    if (!resp.ok) throw new Error("Failed to hold sale");
    clearCart();
    return resp.json();
  }

  async function completeSale() {
    const payload = {
      idempotency_key: saleIdempotencyKey.value,
      subtotal: subtotal.value,
      discount: totalDiscount.value,
      total: total.value,
      items: cloneForStorage(cart).map((item) => ({
        product_id: item.id,
        quantity: item.qty,
        price_id: item.priceId,
        unit_price: item.price,
        discount: item.discount || 0,
        total: item.price * item.qty - (item.discount || 0),
      })),
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

    const resp = await fetch(`${API_BASE}/sales/`, {
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
        ${orderNote.value ? `<p style="text-align:center;font-size:11px;">Note: ${orderNote.value}</p>` : ""}
        <script>window.onload=function(){window.print();window.close();}</` + `script>
      </body>
      </html>`;

    const w = window.open("", "_blank");
    w.document.write(html);
    w.document.close();
  }

  async function deleteHeldTransaction(reference) {
    const resp = await fetch(`${API_BASE}/sales/held/${reference}`, {
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
    if (payload.orderNote) orderNote.value = payload.orderNote;
    if (payload.saleIdempotencyKey) saleIdempotencyKey.value = payload.saleIdempotencyKey;

    holdReference.value = transaction.reference;
  }

  function saveToLocalStorage() {
    const state = {
      cart: cloneForStorage(cart),
      payments: { ...payments },
      customer: customer.value,
      orderNote: orderNote.value,
      holdReference: holdReference.value,
      saleIdempotencyKey: saleIdempotencyKey.value,
    };
    localStorage.setItem("posState", JSON.stringify(state));
  }

  function restoreFromLocalStorage() {
    try {
      const raw = localStorage.getItem("posState");
      if (!raw) return false;
      const state = JSON.parse(raw);
      if (state.cart && Array.isArray(state.cart)) {
        state.cart.forEach((item) => cart.push(item));
      }
      if (state.payments) {
        Object.keys(payments).forEach((k) => {
          payments[k] = state.payments[k] || 0;
        });
      }
      if (state.customer) customer.value = state.customer;
      if (state.orderNote) orderNote.value = state.orderNote;
      if (state.holdReference) holdReference.value = state.holdReference;
      if (state.saleIdempotencyKey) saleIdempotencyKey.value = state.saleIdempotencyKey;
      localStorage.removeItem("posState");
      return true;
    } catch {
      return false;
    }
  }

  // --- Keyboard Shortcuts ---
  let searchInputRef = null;

  function setSearchRef(el) {
    searchInputRef = el;
  }

  function onKeyDown(e) {
    if (e.key === "F3") {
      e.preventDefault();
      searchInputRef?.focus();
    }
    if (e.key === "F5") {
      e.preventDefault();
      if (cart.length > 0) completeSale().then(() => clearCart());
    }
    if (e.key === "F6") {
      e.preventDefault();
      if (cart.length > 0) holdCart();
    }
  }

  onMounted(() => {
    window.addEventListener("keydown", onKeyDown);
    restoreFromLocalStorage();
    window.addEventListener("beforeunload", () => {
      if (cart.length > 0) saveToLocalStorage();
    });
  });

  onUnmounted(() => {
    window.removeEventListener("keydown", onKeyDown);
  });

  return {
    cart,
    payments,
    customer,
    orderNote,
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
    updatePayment,
    clearCart,
    holdCart,
    completeSale,
    printReceipt,
    fetchHeldTransactions,
    deleteHeldTransaction,
    restoreHeld,
    setSearchRef,
  };
}
