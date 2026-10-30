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
      class="rounded-md border px-3 py-2 text-sm"
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

      <p
        v-if="unconfirmed"
        class="rounded-md border border-amber-300 bg-amber-50 p-3 text-sm text-amber-900 dark:border-amber-700 dark:bg-amber-950/40 dark:text-amber-200"
        role="alert"
      >
        We couldn't confirm this adjustment was saved. Press Save again to send the same adjustment: it is applied once, even if the first one got through. The form is locked until then so the two can't differ.
      </p>

      <fieldset :disabled="unconfirmed" class="min-w-0 space-y-4 border-0 p-0">
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
                    · counted {{ line.flags[0].counted_quantity }}<template v-if="sinceFlag(line.flags[0]) !== 0">, {{ signed(sinceFlag(line.flags[0])) }} moved since, so about {{ flagEstimate(line.flags[0]) }} now</template>
                  </span>
                  <span v-if="line.flags.length > 1"> (and {{ line.flags.length - 1 }} more)</span>
                  <label class="mt-1 flex items-center gap-1.5 text-foreground">
                    <input v-model="line.resolvesFlags" type="checkbox" class="h-3.5 w-3.5" />
                    This adjustment answers {{ line.flags.length > 1 ? "these flags" : "the flag" }}
                  </label>
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
                {{ line.stock }} → {{ after(line) }}
                <span class="text-xs text-muted-foreground">({{ signed(change(line)) }})</span>
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
      <p v-else class="text-sm text-muted-foreground">No products yet. Search above to add one.</p>

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
      </fieldset>

      <div class="flex justify-end">
        <Button :disabled="saving || !lines.length" @click="save">
          {{ saving ? "Saving..." : unconfirmed ? "Save again" : "Save adjustment" }}
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
import { ref, watch, onMounted, onBeforeUnmount } from "vue";
import { useRoute, useRouter, onBeforeRouteLeave } from "vue-router";
import { Plus, X } from "lucide-vue-next";
import { Button } from "@/components/ui/button";
import PagerBar from "../components/PagerBar.vue";
import { useConfirm } from "../composables/useConfirm.js";
import { usePermissions } from "../composables/usePermissions.js";
import { useStockFlags } from "../composables/useStockFlags.js";
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
let noticeTimer = null;

const editing = ref(false);
const lines = ref([]);
const reason = ref("");
const note = ref("");
const saving = ref(false);
// A save that failed without a clear refusal (timeout, lost reply) may have gone through. The form
// is locked until it is sent again with the same retry key, so an edit can't turn the retry into
// a second, different adjustment.
const unconfirmed = ref(false);
const query = ref("");
const results = ref([]);
let searchTimer = null;
let idempotencyKey = "";

const dismissing = ref(null);
const dismissNote = ref("");
const acting = ref(false);

function showNotice(text, error = false) {
  notice.value = { text, error };
  clearTimeout(noticeTimer);
  noticeTimer = setTimeout(() => { notice.value = null; }, error ? 8000 : 4000);
}

async function apiError(r, fallback) {
  const body = await r.json().catch(() => ({}));
  return body.error || fallback;
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
  if (!r.ok) throw new Error(await apiError(r, "Could not load flagged products."));
  flags.value = (await r.json()) || [];
}

async function loadHistory() {
  const r = await fetch(`/inventory/stock-adjustments?page=${historyPage.value}`);
  if (!r.ok) throw new Error(await apiError(r, "Could not load stock adjustments."));
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
    if (!r.ok) throw new Error(await apiError(r, "Could not dismiss the flag."));
    dismissing.value = null;
    showNotice(`Flag on ${f.product_name} dismissed.`);
    refreshFlagCount();
    await load([loadFlags]);
  } catch (e) {
    showNotice(e.message, true);
  } finally {
    acting.value = false;
  }
}

// ===== editor =====
function newKey() {
  return globalThis.crypto?.randomUUID?.() || `adj-${Date.now()}-${Math.random().toString(36).slice(2)}`;
}

function startNew() {
  lines.value = [];
  reason.value = "";
  note.value = "";
  query.value = "";
  results.value = [];
  unconfirmed.value = false;
  editing.value = true;
}

function isDirty() {
  return editing.value && (lines.value.length > 0 || note.value.trim() !== "");
}

// After an unconfirmed save the adjustment may already be stored, so leaving must not say
// otherwise.
function confirmDiscard() {
  if (unconfirmed.value) {
    return confirm({
      title: "Leave without confirming the save?",
      message: "This adjustment may already have been saved. Check History before entering it again.",
      confirmLabel: "Leave",
      cancelLabel: "Stay",
    });
  }
  return confirm({
    title: "Discard this adjustment?",
    message: "Nothing has been saved yet.",
    confirmLabel: "Discard",
    cancelLabel: "Keep editing",
  });
}

async function cancelEdit() {
  const wasUnconfirmed = unconfirmed.value;
  if (isDirty() && !(await confirmDiscard())) return;
  editing.value = false;
  unconfirmed.value = false;
  if (wasUnconfirmed) {
    tab.value = "history";
    await load();
  }
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
      if (r.ok) results.value = (await r.json()) || [];
    } catch {
      results.value = [];
    }
  }, 200);
}

// Search results don't carry stock, so a product picked from search is loaded first.
async function addFromSearch(p) {
  query.value = "";
  results.value = [];
  try {
    const r = await fetch(`/inventory/product/${p.id}`);
    if (!r.ok) throw new Error();
    const full = await r.json();
    addProduct({ id: full.id, name: full.name, manufacturer: full.manufacturer, stock: full.stock });
  } catch {
    showNotice(`Could not load ${p.name}.`, true);
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
    // Open flags on the product are shown; they are resolved only if the manager ticks that this
    // adjustment answers them (adjusting from a flag ticks it).
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

// From a flag: the cashier's count was taken when the flag was raised, so carry it forward by the
// sales, receipts and returns since (moved_since leaves out corrections) to estimate the shelf.
function flagEstimate(f) {
  if (f.counted_quantity === null || f.counted_quantity === undefined) return f.current_stock;
  return Math.max(0, f.counted_quantity + f.moved_since);
}

function sinceFlag(f) {
  return f.moved_since;
}

function adjustFlag(f) {
  startNew();
  reason.value = "count_correction";
  addProduct(
    { id: f.product_id, name: f.product_name, manufacturer: f.manufacturer, stock: f.current_stock },
    {
      quantity: flagEstimate(f),
      flags: [f, ...flags.value.filter((o) => o.product_id === f.product_id && o.id !== f.id)],
      resolvesFlags: true,
    },
  );
}

function validate() {
  if (!reason.value) return "Choose a reason.";
  if (reason.value === "other" && !note.value.trim()) return "Add a note saying what this adjustment is for.";
  for (const l of lines.value) {
    const q = Number(l.quantity);
    // An emptied box reads as "" (Number gives 0), which must not become "set to 0".
    if (l.quantity === "" || l.quantity === null || !Number.isInteger(q) || q < (l.mode === "set" ? 0 : 1)) {
      return `${l.name}: enter a whole quantity of at least ${l.mode === "set" ? 0 : 1}.`;
    }
    if (after(l) < 0) return `${l.name}: only ${l.stock} in stock.`;
    if (change(l) === 0 && !l.expiry) {
      return l.resolvesFlags && l.flags.length
        ? `${l.name}: the stock already matches, so there is nothing to adjust. Dismiss the flag instead.`
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
      `${l.name}: ${l.stock} → ${after(l)} (${signed(change(l))})` +
      (l.resolvesFlags && l.flags.length ? `, resolves ${l.flags.length > 1 ? `${l.flags.length} flags` : "the flag"}` : ""),
  );
  // One retry key per save attempt. It is kept only while a save is unconfirmed, so Save again
  // resends that same adjustment; any new attempt is a different save with a new key.
  if (!unconfirmed.value) idempotencyKey = newKey();
  const ok = await confirm({
    title: "Save this stock adjustment?",
    message: `${reasonLabel(reason.value)}. Stock changes straight away and the adjustment is recorded with your name.`,
    lines: summary.slice(0, 8).concat(summary.length > 8 ? [`and ${summary.length - 8} more`] : []),
    confirmLabel: unconfirmed.value ? "Save again" : "Save adjustment",
    busyLabel: "Saving…",
    cancelLabel: "Go back",
    run: postAdjustment,
  });
  if (!ok) return;
  // Refused: the dialog closed and the lines were brought up to date, so the next save is
  // validated and confirmed afresh.
  if (refused) return;

  editing.value = false;
  unconfirmed.value = false;
  showNotice("Stock adjusted.");
  refreshFlagCount();
  tab.value = "history";
  if (historyPage.value === 1) await load();
  else {
    historyPage.value = 1; // the watcher reloads history
    await load([loadFlags]);
  }
}

let refused = false;

const UNCONFIRMED_SAVE =
  "We couldn't confirm this was saved. Press Save again to send the same adjustment: it is applied once, even if the first one got through.";

// Sends the adjustment from the confirmation dialog. A 4xx means the server refused it and nothing
// was stored: the dialog closes (its summary would be out of date), the lines are brought up to
// date (stock moved by a sale, a flag handled by someone else) and the form unlocks. Any other
// failure (timeout, lost reply, a failed commit) may have saved it: the dialog stays open so Save
// again resends the same adjustment with the same retry key, and the form locks.
async function postAdjustment() {
  refused = false;
  const body = {
    reason: reason.value,
    note: note.value.trim(),
    idempotency_key: idempotencyKey,
    items: lines.value.map((l) => ({
      product_id: l.product_id,
      mode: l.mode,
      quantity: Number(l.quantity),
      expiry: l.expiry || null,
    })),
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
      unconfirmed.value = true;
      throw new Error(UNCONFIRMED_SAVE);
    }
    if (r.ok) return;
    const message = await apiError(r, "Could not save the adjustment.");
    if (r.status >= 400 && r.status < 500) {
      unconfirmed.value = false;
      refused = true;
      showNotice(message, true);
      await refreshLines();
      return;
    }
    unconfirmed.value = true;
    throw new Error(UNCONFIRMED_SAVE);
  } finally {
    saving.value = false;
  }
}

async function refreshLines() {
  await loadFlags().catch(() => {});
  const open = new Set(flags.value.map((f) => f.id));
  await Promise.all(
    lines.value.map(async (l) => {
      const kept = l.flags.filter((f) => open.has(f.id));
      if (kept.length !== l.flags.length) l.flags = kept;
      try {
        const r = await fetch(`/inventory/product/${l.product_id}`);
        if (!r.ok) return;
        const { stock } = await r.json();
        if (stock !== l.stock) l.stock = stock;
      } catch {
        // Keep the stock already shown; the server checks it again on save.
      }
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
      const r = await fetch(`/inventory/product/${productID}`);
      if (!r.ok) throw new Error();
      const p = await r.json();
      startNew();
      addProduct({ id: p.id, name: p.name, manufacturer: p.manufacturer, stock: p.stock });
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
