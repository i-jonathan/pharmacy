<template>
  <div v-if="!canAdjust" class="p-4 sm:p-6 lg:p-8">
    <h1 class="text-2xl font-semibold tracking-tight">Stock Adjustments</h1>
    <p class="mt-2 text-sm text-muted-foreground">
      Only managers can adjust stock. To report a problem, flag the product from Products.
    </p>
  </div>
  <div v-else class="space-y-5 p-4 sm:p-6 lg:p-8">
    <header class="flex flex-wrap items-start justify-between gap-3">
      <div>
        <h1 class="text-2xl font-semibold tracking-tight">Stock Adjustments</h1>
        <p class="mt-1 text-sm text-muted-foreground">
          Correct stock outside sales and receiving: write-offs, losses, opening stock and miscounts.
        </p>
      </div>
      <Button v-if="!editing" @click="startNew()">
        <Plus :size="16" class="mr-2" />
        New adjustment
      </Button>
    </header>

    <p
      v-if="notice"
      ref="noticeEl"
      class="scroll-mt-20 rounded-md border px-3 py-2 text-sm"
      :class="notice.error ? 'border-destructive/30 bg-destructive/5 text-destructive' : 'border-border bg-muted/40'"
      role="status"
    >
      {{ notice.text }}
    </p>

    <!-- ===== EDITOR ===== -->
    <section v-if="editing" class="space-y-4 rounded-lg border border-border bg-card p-4">
      <div class="flex flex-wrap items-center justify-between gap-2">
        <h2 class="text-lg font-semibold">New adjustment</h2>
        <Button variant="outline" size="sm" :disabled="saving" @click="cancelEdit">
          <X :size="14" class="mr-1" />
          Cancel
        </Button>
      </div>

      <div class="relative">
        <input
          v-model="query"
          class="w-full rounded-md border border-border bg-background px-3 py-2 text-sm"
          placeholder="Search products to adjust"
          aria-label="Search products"
          @input="searchProducts"
        />
        <div
          v-if="results.length"
          class="absolute z-10 mt-1 max-h-64 w-full overflow-y-auto rounded-md border border-border bg-card shadow-lg"
        >
          <button
            v-for="p in results"
            :key="p.id"
            type="button"
            class="flex w-full items-center justify-between gap-3 border-b border-border p-3 text-left text-sm last:border-0 hover:bg-muted/40"
            @click="addFromSearch(p)"
          >
            <span>
              {{ p.name }}
              <small class="text-muted-foreground">{{ p.manufacturer }}</small>
            </span>
          </button>
        </div>
      </div>

      <div v-if="lines.length" class="overflow-x-auto rounded-md border border-border">
        <table class="w-full text-sm">
          <thead class="bg-muted/30 text-xs uppercase tracking-wider text-muted-foreground">
            <tr>
              <th class="p-2 text-left">Product</th>
              <th class="p-2 text-left">Change</th>
              <th class="w-28 p-2 text-left">Quantity</th>
              <th class="w-40 p-2 text-left">New expiry</th>
              <th class="w-36 p-2 text-right">Stock</th>
              <th class="w-16 p-2"></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="line in lines" :key="line.product_id" class="border-t border-border align-top">
              <td class="p-2">
                {{ line.name }}
                <span v-if="line.manufacturer" class="ml-1 text-xs text-muted-foreground">{{ line.manufacturer }}</span>
                <div v-if="line.flags.length" class="mt-0.5 text-xs text-amber-700 dark:text-amber-300">
                  Flagged: {{ line.flags[0].note }}
                  <span v-if="line.flags[0].counted_quantity !== null">
                    · counted {{ line.flags[0].counted_quantity }} when the system said {{ line.flags[0].stock_at_flag }}
                  </span>
                  <span v-if="line.flags.length > 1"> (and {{ line.flags.length - 1 }} more)</span>
                  <label class="mt-1 flex items-center gap-1.5 text-foreground">
                    <input v-model="line.resolvesFlags" type="checkbox" class="h-3.5 w-3.5" />
                    This adjustment answers {{ line.flags.length > 1 ? "these flags" : "the flag" }}
                  </label>
                </div>
                <div v-if="line.stockChanged" class="mt-0.5 text-xs text-destructive">
                  Stock changed from {{ line.stockChanged.from }} to {{ line.stockChanged.to }}; enter the count again.
                </div>
              </td>
              <td class="p-2">
                <div class="inline-flex overflow-hidden rounded-md border border-border" role="group" :aria-label="`How to change ${line.name}`">
                  <button
                    v-for="m in MODES"
                    :key="m.key"
                    type="button"
                    class="whitespace-nowrap px-2.5 py-1 text-xs"
                    :class="line.mode === m.key ? 'bg-primary text-primary-foreground' : 'hover:bg-accent'"
                    :aria-pressed="line.mode === m.key"
                    @click="setMode(line, m.key)"
                  >
                    {{ m.label }}
                  </button>
                </div>
              </td>
              <td class="p-2">
                <input
                  v-model.number="line.quantity"
                  type="number"
                  @input="line.stockChanged = null"
                  :min="line.mode === 'set' ? 0 : 1"
                  class="w-24 rounded-md border border-border bg-background px-2 py-1 text-sm"
                  :aria-label="`Quantity for ${line.name}`"
                />
              </td>
              <td class="p-2">
                <input
                  v-model="line.expiry"
                  type="date"
                  class="w-36 rounded-md border border-border bg-background px-2 py-1 text-sm"
                  :aria-label="`New expiry for ${line.name}`"
                />
              </td>
              <td class="p-2 text-right tabular-nums" :class="after(line) < 0 ? 'text-destructive' : ''">
                <template v-if="line.quantity === ''">{{ line.stock }} → ?</template>
                <template v-else>
                  {{ line.stock }} → {{ after(line) }}
                  <span class="text-xs text-muted-foreground">({{ signed(change(line)) }})</span>
                </template>
              </td>
              <td class="p-2 text-right">
                <button type="button" class="text-xs text-red-600 hover:underline dark:text-red-400" @click="removeLine(line)">
                  Remove
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <p v-else-if="!loadingProducts.length" class="text-sm text-muted-foreground">No products yet. Search above to add one.</p>
      <p v-if="loadingProducts.length" class="text-sm text-muted-foreground" role="status">
        Loading {{ loadingProducts.join(", ") }}…
      </p>

      <div class="grid gap-3 md:grid-cols-2">
        <label class="block text-sm">
          Reason
          <select v-model="reason" class="mt-1 w-full rounded-md border border-border bg-background px-3 py-2 text-sm">
            <option disabled value="">Choose a reason</option>
            <option v-for="r in REASONS" :key="r.key" :value="r.key">{{ r.label }}</option>
          </select>
        </label>
        <label class="block text-sm">
          Note <span class="text-muted-foreground">{{ reason === "other" ? "(required)" : "(optional)" }}</span>
          <input
            v-model="note"
            maxlength="500"
            class="mt-1 w-full rounded-md border border-border bg-background px-3 py-2 text-sm"
            placeholder="What happened, where it was found, who counted"
          />
        </label>
      </div>

      <div class="flex justify-end">
        <Button :disabled="saving || loadingProducts.length > 0 || !lines.length" @click="save">
          {{ saving ? "Saving..." : "Save adjustment" }}
        </Button>
      </div>
    </section>

    <!-- ===== LISTS ===== -->
    <section v-if="!editing" class="space-y-3">
      <div class="flex gap-1 border-b border-border">
        <button
          v-for="t in TABS"
          :key="t.key"
          type="button"
          class="-mb-px border-b-2 px-4 py-2 text-sm font-medium"
          :class="tab === t.key ? 'border-primary text-foreground' : 'border-transparent text-muted-foreground hover:text-foreground'"
          @click="tab = t.key"
        >
          {{ t.label }}<span v-if="t.key === 'flags' && flags.length"> ({{ flags.length }})</span>
        </button>
      </div>

      <div v-if="loading" class="py-12 text-center text-sm text-muted-foreground">Loading...</div>
      <div v-else-if="loadError" class="py-12 text-center text-sm">
        <p class="mb-3 text-muted-foreground">{{ loadError }}</p>
        <Button variant="outline" size="sm" @click="load">Retry</Button>
      </div>

      <!-- Flagged -->
      <template v-else-if="tab === 'flags'">
        <div v-if="!flags.length" class="rounded-lg border border-border bg-card p-8 text-center text-sm text-muted-foreground">
          Nothing is flagged for checking.
        </div>
        <article v-for="f in flags" :key="f.id" class="rounded-lg border border-l-4 border-border border-l-amber-500 bg-card p-4">
          <div class="flex flex-wrap items-start justify-between gap-3">
            <div class="min-w-0 flex-1 basis-72">
              <div class="font-medium">
                {{ f.product_name }}
                <span class="ml-1 text-sm font-normal text-muted-foreground">{{ f.manufacturer }}</span>
              </div>
              <p class="mt-1 text-sm">{{ f.note }}</p>
              <p class="mt-1 text-xs text-muted-foreground">
                In stock {{ f.current_stock }}<span v-if="f.counted_quantity !== null"> · counted {{ f.counted_quantity }} when it was {{ f.stock_at_flag }}</span>
                · flagged by {{ f.flagged_by || "unknown" }} · {{ formatDate(f.created_at) }}
              </p>
            </div>
            <div v-if="dismissing === f.id" class="w-full space-y-2 md:w-auto md:min-w-80">
              <input
                v-model="dismissNote"
                maxlength="500"
                class="w-full rounded-md border border-border bg-background px-3 py-2 text-sm"
                placeholder="Why no adjustment is needed"
                :aria-label="`Why ${f.product_name} needs no adjustment`"
              />
              <div class="flex gap-2">
                <Button size="sm" :disabled="acting || !dismissNote.trim()" @click="dismiss(f)">
                  {{ acting ? "Saving..." : "Dismiss flag" }}
                </Button>
                <Button size="sm" variant="outline" :disabled="acting" @click="dismissing = null">Keep it</Button>
              </div>
            </div>
            <div v-else class="flex gap-2">
              <Button size="sm" variant="outline" @click="startDismiss(f)">Dismiss</Button>
              <Button size="sm" @click="adjustFlag(f)">Adjust</Button>
            </div>
          </div>
        </article>
      </template>

      <!-- History -->
      <template v-else>
        <div v-if="!history.length" class="rounded-lg border border-border bg-card p-8 text-center text-sm text-muted-foreground">
          No stock adjustments yet.
        </div>
        <article v-for="a in history" :key="a.id" class="rounded-lg border border-border bg-card p-4">
          <div class="flex flex-wrap items-baseline justify-between gap-2">
            <div class="font-medium">
              {{ reasonLabel(a.reason) }}
              <span class="ml-1 text-sm font-normal text-muted-foreground">by {{ a.created_by || "unknown" }}</span>
            </div>
            <div class="text-xs text-muted-foreground">{{ formatDateTime(a.created_at) }}</div>
          </div>
          <p v-if="a.note" class="mt-1 text-sm text-muted-foreground">{{ a.note }}</p>
          <ul class="mt-2 space-y-0.5 text-sm">
            <li v-for="i in a.items" :key="i.product_id" class="flex flex-wrap justify-between gap-2">
              <span>{{ i.product_name }} <small class="text-muted-foreground">{{ i.manufacturer }}</small></span>
              <span class="tabular-nums">
                {{ i.quantity_before }} → {{ i.quantity_before + i.quantity_change }}
                <span :class="i.quantity_change < 0 ? 'text-destructive' : 'text-emerald-700 dark:text-emerald-400'">({{ signed(i.quantity_change) }})</span>
                <span v-if="i.expiry_after" class="ml-2 text-xs text-muted-foreground">expiry {{ formatDay(i.expiry_after) }}</span>
              </span>
            </li>
          </ul>
        </article>
        <PagerBar v-if="historyTotal > PER_PAGE" v-model:page="historyPage" :total="historyTotal" :size="PER_PAGE" />
      </template>
    </section>
  </div>
</template>

<script setup>
import { ref, watch, nextTick, onMounted, onBeforeUnmount } from "vue";
import { useRoute, useRouter, onBeforeRouteLeave } from "vue-router";
import { Plus, X } from "lucide-vue-next";
import { Button } from "@/components/ui/button";
import PagerBar from "../components/PagerBar.vue";
import { useConfirm } from "../composables/useConfirm.js";
import { usePermissions } from "../composables/usePermissions.js";
import { useStockFlags } from "../composables/useStockFlags.js";
import { readError, newRetryKey } from "../lib/api.js";
import { csrfFetch } from "../lib/csrf.js";

const PER_PAGE = 20;
const MODES = [
  { key: "set", label: "Set to" },
  { key: "add", label: "Add" },
  { key: "remove", label: "Remove" },
];
const REASONS = [
  { key: "expired", label: "Expired" },
  { key: "damaged", label: "Damaged" },
  { key: "theft_loss", label: "Theft or loss" },
  { key: "count_correction", label: "Count correction" },
  { key: "opening_stock", label: "Opening stock" },
  { key: "internal_use", label: "Internal use" },
  { key: "other", label: "Other" },
];
const TABS = [
  { key: "flags", label: "Flagged" },
  { key: "history", label: "History" },
];

const { hasPermission } = usePermissions();
const canAdjust = hasPermission("stock:adjust");
const { confirm } = useConfirm();
const { refresh: refreshFlagCount } = useStockFlags();
const route = useRoute();
const router = useRouter();

const tab = ref("flags");
const loading = ref(false);
const loadError = ref("");
const flags = ref([]);
const history = ref([]);
const historyPage = ref(1);
const historyTotal = ref(0);
const notice = ref(null);
const noticeEl = ref(null);
let noticeTimer = null;

const editing = ref(false);
const lines = ref([]);
const reason = ref("");
const note = ref("");
const saving = ref(false);
// Names of products still loading for the open form. Save waits for them, so the confirmation
// always lists every line that will be sent; startNew clears it with the old form.
const loadingProducts = ref([]);
// Bumped for each new form, so a product that finishes loading after its form closed is dropped.
let formGeneration = 0;
const query = ref("");
const results = ref([]);
let searchTimer = null;
// One retry key per form. The server answers a resend of the same lines with the first save, and
// refuses the key for anything else, so a retry after an unclear failure can't apply twice.
let idempotencyKey = "";

const dismissing = ref(null);
const dismissNote = ref("");
const acting = ref(false);

function showNotice(text, error = false) {
  notice.value = { text, error };
  clearTimeout(noticeTimer);
  noticeTimer = setTimeout(() => { notice.value = null; }, error ? 8000 : 4000);
  // On a phone the notice can be far above the button that caused it.
  if (error) nextTick(() => noticeEl.value?.scrollIntoView({ block: "nearest", behavior: "smooth" }));
}

function reasonLabel(key) {
  return REASONS.find((r) => r.key === key)?.label || key;
}

function signed(n) {
  return n > 0 ? `+${n}` : String(n);
}

function formatDate(d) {
  return d ? new Date(d).toLocaleDateString() : "";
}

// Expiry dates are calendar days sent as midnight UTC; show that day in every timezone.
function formatDay(d) {
  return d ? new Date(d).toLocaleDateString(undefined, { timeZone: "UTC" }) : "";
}

function formatDateTime(d) {
  return d ? new Date(d).toLocaleString() : "";
}

// ===== lists =====
async function loadFlags() {
  const r = await fetch("/inventory/stock-flags");
  if (!r.ok) throw new Error((await readError(r, "Could not load flagged products.")).message);
  flags.value = (await r.json()) || [];
}

async function loadHistory() {
  const r = await fetch(`/inventory/stock-adjustments?page=${historyPage.value}`);
  if (!r.ok) throw new Error((await readError(r, "Could not load stock adjustments.")).message);
  const page = await r.json();
  history.value = page.data || [];
  historyTotal.value = page.total_count || 0;
}

async function load(parts = [loadFlags, loadHistory]) {
  loading.value = true;
  loadError.value = "";
  try {
    await Promise.all(parts.map((p) => p()));
  } catch (e) {
    loadError.value = e.message;
  } finally {
    loading.value = false;
  }
}

watch(historyPage, () => load([loadHistory]));

function startDismiss(f) {
  dismissing.value = f.id;
  dismissNote.value = "";
}

async function dismiss(f) {
  acting.value = true;
  try {
    const r = await csrfFetch(`/inventory/stock-flags/${f.id}/dismiss`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ note: dismissNote.value.trim() }),
    });
    if (!r.ok) throw new Error((await readError(r, "Could not dismiss the flag.")).message);
    dismissing.value = null;
    showNotice(`Flag on ${f.product_name} dismissed.`);
    refreshFlagCount();
    await load([loadFlags]);
  } catch (e) {
    showNotice(e.message, true);
    // Refused because someone else handled it: show the list as it is now.
    await loadFlags().catch(() => {});
    refreshFlagCount();
    if (!flags.value.some((x) => x.id === f.id)) dismissing.value = null;
  } finally {
    acting.value = false;
  }
}

// ===== editor =====

function startNew() {
  lines.value = [];
  reason.value = "";
  note.value = "";
  query.value = "";
  results.value = [];
  formGeneration += 1;
  loadingProducts.value = [];
  idempotencyKey = newRetryKey();
  editing.value = true;
}

function isDirty() {
  return editing.value && (lines.value.length > 0 || note.value.trim() !== "");
}

function confirmDiscard() {
  return confirm({
    title: "Discard this adjustment?",
    message: "Lines you haven't saved will be lost. If a save didn't confirm, check History first.",
    confirmLabel: "Discard",
    cancelLabel: "Keep editing",
  });
}

async function cancelEdit() {
  if (isDirty() && !(await confirmDiscard())) return;
  editing.value = false;
}

function searchProducts() {
  clearTimeout(searchTimer);
  const q = query.value.trim();
  if (q.length < 2) {
    results.value = [];
    return;
  }
  searchTimer = setTimeout(async () => {
    try {
      const r = await fetch(`/inventory/search?query=${encodeURIComponent(q)}`);
      const found = r.ok ? (await r.json()) || [] : [];
      // A slower, older search must not replace the results for what is typed now.
      if (query.value.trim() === q) results.value = found;
    } catch {
      if (query.value.trim() === q) results.value = [];
    }
  }, 200);
}

// Search results don't carry stock (they say 0), so the product is loaded for its real stock.
async function loadProduct(id) {
  // A request that hangs would otherwise keep Save disabled until the page is reloaded.
  const r = await fetch(`/inventory/product/${id}`, { signal: AbortSignal.timeout(15000) });
  if (!r.ok) throw new Error();
  const p = await r.json();
  return { id: p.id, name: p.name, manufacturer: p.manufacturer, stock: p.stock };
}

async function addFromSearch(p) {
  query.value = "";
  results.value = [];
  const generation = formGeneration;
  // Only a load for the form still open may add a line, report an error or hold up Save.
  const current = () => editing.value && generation === formGeneration;
  loadingProducts.value = [...loadingProducts.value, p.name];
  try {
    const product = await loadProduct(p.id);
    if (current()) addProduct(product);
  } catch {
    if (current()) showNotice(`Could not load ${p.name}.`, true);
  } finally {
    if (generation === formGeneration) {
      const i = loadingProducts.value.indexOf(p.name);
      if (i >= 0) loadingProducts.value = loadingProducts.value.filter((_, j) => j !== i);
    }
  }
}

function addProduct(p, extra = {}) {
  query.value = "";
  results.value = [];
  if (lines.value.some((l) => l.product_id === p.id)) {
    showNotice(`${p.name} is already in this adjustment.`, true);
    return;
  }
  lines.value.push({
    product_id: p.id,
    name: p.name,
    manufacturer: p.manufacturer || "",
    stock: p.stock ?? 0,
    mode: "set",
    quantity: p.stock ?? 0,
    expiry: "",
    // Open flags are shown; they resolve only when ticked (adjusting from a flag ticks them).
    flags: flags.value.filter((f) => f.product_id === p.id),
    resolvesFlags: false,
    ...extra,
  });
}

function removeLine(line) {
  lines.value = lines.value.filter((l) => l !== line);
}

function setMode(line, mode) {
  if (line.mode === mode) return;
  line.mode = mode;
  line.stockChanged = null;
  line.quantity = mode === "set" ? line.stock : 1;
}

function change(line) {
  const q = Number(line.quantity) || 0;
  if (line.mode === "set") return q - line.stock;
  return line.mode === "add" ? q : -q;
}

function after(line) {
  return line.stock + change(line);
}

function adjustFlag(f) {
  startNew();
  reason.value = "count_correction";
  addProduct(
    { id: f.product_id, name: f.product_name, manufacturer: f.manufacturer, stock: f.current_stock },
    {
      flags: [f, ...flags.value.filter((o) => o.product_id === f.product_id && o.id !== f.id)],
      resolvesFlags: true,
    },
  );
}

// An expiry fix with no stock change; it is sent as "add 0", so it has no stock target.
function isExpiryOnly(l) {
  return Boolean(l.expiry) && l.quantity !== "" && change(l) === 0;
}

function validate() {
  if (!reason.value) return "Choose a reason.";
  if (reason.value === "other" && !note.value.trim()) return "Add a note saying what this adjustment is for.";
  for (const l of lines.value) {
    if (isExpiryOnly(l)) continue; // fine even when stock is below zero
    const q = Number(l.quantity);
    // An emptied box reads as "" (Number gives 0), which must not become "set to 0".
    if (l.quantity === "" || l.quantity === null || !Number.isInteger(q) || q < (l.mode === "set" ? 0 : 1)) {
      return `${l.name}: enter a whole quantity of at least ${l.mode === "set" ? 0 : 1}.`;
    }
    if (change(l) < 0 && after(l) < 0) return `${l.name}: only ${l.stock} in stock.`;
    if (change(l) === 0 && !l.expiry) {
      return l.resolvesFlags && l.flags.length
        ? `${l.name}: enter the count you found, or dismiss the flag if the stock is right.`
        : `${l.name}: nothing to change.`;
    }
  }
  return "";
}

async function save() {
  const problem = validate();
  if (problem) {
    showNotice(problem, true);
    return;
  }
  const summary = lines.value.map(
    (l) =>
      `${l.name}${l.manufacturer ? ` (${l.manufacturer})` : ""}: ${l.stock} → ${after(l)} (${signed(change(l))})` +
      (l.resolvesFlags && l.flags.length ? `, resolves ${l.flags.length > 1 ? `${l.flags.length} flags` : "the flag"}` : ""),
  );
  const ok = await confirm({
    title: "Save this stock adjustment?",
    message: `${reasonLabel(reason.value)}. Stock changes straight away and the adjustment is recorded with your name.`,
    lines: summary.slice(0, 8).concat(summary.length > 8 ? [`and ${summary.length - 8} more`] : []),
    confirmLabel: "Save adjustment",
    busyLabel: "Saving…",
    cancelLabel: "Go back",
    run: postAdjustment,
  });
  if (!ok) return;
  // Refused: the lines were refreshed, so the next save is checked and confirmed again.
  if (refused) return;

  showNotice("Stock adjusted.");
  await closeOnHistory();
}

// The form is done (saved now, or found already saved): show the newest adjustments.
async function closeOnHistory() {
  editing.value = false;
  refreshFlagCount();
  tab.value = "history";
  if (historyPage.value === 1) await load();
  else {
    historyPage.value = 1; // the watcher reloads history
    await load([loadFlags]);
  }
}

let refused = false;

const UNSURE =
  "We couldn't confirm this was saved. Press Save adjustment again: if the first one got through, it won't be applied twice.";

// Runs inside the confirm dialog. A refusal (4xx) closes it and shows why; after a 400 or 409 the
// lines are reloaded. Any other failure may have saved it, so the dialog stays open to resend: the
// form's key makes a resend of the same lines safe.
async function postAdjustment() {
  refused = false;
  const body = {
    reason: reason.value,
    note: note.value.trim(),
    idempotency_key: idempotencyKey,
    items: lines.value.map((l) =>
      isExpiryOnly(l)
        ? { product_id: l.product_id, mode: "add", quantity: 0, expiry: l.expiry || null }
        : {
            product_id: l.product_id,
            mode: l.mode,
            quantity: Number(l.quantity),
            expiry: l.expiry || null,
            // A count is checked against the stock it was made from.
            expected_stock: l.mode === "set" ? l.stock : null,
          },
    ),
    flag_ids: lines.value.filter((l) => l.resolvesFlags).flatMap((l) => l.flags.map((f) => f.id)),
  };
  saving.value = true;
  try {
    let r;
    try {
      r = await csrfFetch("/inventory/stock-adjustments", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      });
    } catch {
      throw new Error(UNSURE);
    }
    if (r.ok) return;
    if (r.status >= 500) throw new Error(UNSURE);
    refused = true;
    const reply = await readError(r, "Could not save the adjustment.");
    showNotice(reply.message, true);
    if (reply.code === "already_saved") {
      // This form's one save already went through: it is done, and History shows what was saved.
      await closeOnHistory();
      return;
    }
    if (r.status === 400 || r.status === 409) await refreshLines();
  } finally {
    saving.value = false;
  }
}

// After a refusal, reload each line's stock. A count typed against the old stock can't be kept
// (it would be applied to the new one), so it is cleared with a note; a line meant as no change
// (an expiry fix) stays no change.
async function refreshLines() {
  await loadFlags().catch(() => {});
  const open = new Set(flags.value.map((f) => f.id));
  await Promise.all(
    lines.value.map(async (l) => {
      l.flags = l.flags.filter((f) => open.has(f.id));
      let stock;
      try {
        stock = (await loadProduct(l.product_id)).stock;
      } catch {
        // Keep the stock shown; the server checks it again on save.
      }
      if (stock === undefined || stock === l.stock) return;
      if (l.mode === "set") {
        const noChange = change(l) === 0;
        l.quantity = noChange ? stock : "";
        l.stockChanged = noChange ? null : { from: l.stock, to: stock };
      }
      l.stock = stock;
    }),
  );
}

// ===== lifecycle =====
function warnBeforeUnload(e) {
  if (isDirty()) {
    e.preventDefault();
    e.returnValue = "";
  }
}

onBeforeRouteLeave(async () => {
  // Mid-save the dialog belongs to the save; a discard prompt would take it over.
  if (saving.value) {
    showNotice("Wait for the save to finish.", true);
    return false;
  }
  if (!isDirty()) return true;
  return confirmDiscard();
});

onMounted(async () => {
  window.addEventListener("beforeunload", warnBeforeUnload);
  if (!canAdjust) return;
  await load();
  if (!flags.value.length && history.value.length) tab.value = "history";
  // Opened from a product's "Adjust stock": start with that product.
  const productID = Number(route.query.product);
  if (productID) {
    router.replace({ path: route.path });
    try {
      const p = await loadProduct(productID);
      if (editing.value) {
        // A form was started while the page loaded; adding to it could change a save already
        // being confirmed, so leave it to the manager.
        showNotice(`To adjust ${p.name}, add it from the search.`);
      } else {
        startNew();
        addProduct(p);
      }
    } catch {
      showNotice("Could not load that product.", true);
    }
  }
});

onBeforeUnmount(() => {
  window.removeEventListener("beforeunload", warnBeforeUnload);
  clearTimeout(noticeTimer);
  clearTimeout(searchTimer);
});
</script>
