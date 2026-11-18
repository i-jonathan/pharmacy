<template>
  <div v-if="canLookup || modelValue" class="px-4 py-1.5" @keydown.esc.stop="close">
    <!-- Finding a customer -->
    <div v-if="mode === 'search'" class="space-y-1.5">
      <div class="flex items-center gap-2">
        <Search :size="14" class="shrink-0 text-muted-foreground" aria-hidden="true" />
        <input
          ref="searchInput"
          v-model="query"
          type="text"
          autocomplete="off"
          class="min-w-0 flex-1 border-none bg-transparent text-sm outline-none"
          maxlength="120"
          placeholder="Search by phone or name"
          aria-label="Search customers by phone or name"
          @input="error = ''"
          @keydown.enter.prevent="pickFirst"
        />
        <button type="button" class="rounded p-1 text-muted-foreground hover:text-foreground" aria-label="Cancel" @click="close">
          <X :size="14" />
        </button>
      </div>
      <p v-if="error" role="alert" class="rounded-md bg-destructive/10 px-2 py-1 text-xs text-destructive">{{ error }}</p>
      <ul v-if="results.length" class="divide-y divide-border overflow-hidden rounded-md border border-border" aria-label="Matching customers">
        <li v-for="c in results.slice(0, 6)" :key="c.id">
          <div
            v-if="c.active === false"
            class="flex w-full items-center justify-between gap-2 bg-muted/40 px-2 py-1.5 text-sm text-muted-foreground"
          >
            <span class="min-w-0 truncate">
              <span class="font-medium">{{ c.name }}</span>
              <span class="ml-2 text-xs tabular-nums">{{ showPhone(c.phone) }}</span>
            </span>
            <span class="shrink-0 rounded bg-muted px-1.5 py-0.5 text-[11px] font-medium uppercase tracking-wide">Deactivated</span>
            <Button v-if="canManage" variant="outline" size="sm" class="h-6 shrink-0 text-xs" :disabled="disabled" @click="reactivate(c)">Reactivate</Button>
            <span v-else class="shrink-0 text-xs">Ask a manager</span>
          </div>
          <button
            v-else
            type="button"
            class="flex w-full items-center justify-between gap-3 px-2 py-1.5 text-left text-sm hover:bg-muted focus-visible:bg-muted focus-visible:outline-none"
            @click="pick(c)"
          >
            <span class="truncate font-medium">{{ c.name }}</span>
            <span class="shrink-0 text-xs tabular-nums text-muted-foreground">{{ showPhone(c.phone) }}</span>
          </button>
        </li>
      </ul>
      <p v-if="results.length > 6" class="text-xs text-muted-foreground">Keep typing to narrow the list.</p>
      <p v-else-if="results.length > 1" class="text-xs text-muted-foreground">Choose a customer, or keep typing to narrow the list.</p>
      <p v-else-if="noMatch" class="text-xs text-muted-foreground" role="status">No customer found.</p>
      <Button v-if="noMatch && canCreate" variant="outline" size="sm" class="h-7 w-full min-w-0 text-xs" @click="startNew">
        <span class="truncate">Add "{{ query.trim() }}" as a new customer</span>
      </Button>
    </div>

    <!-- Adding a new customer without leaving the sale -->
    <form v-else-if="mode === 'new'" class="space-y-1.5" @submit.prevent="create">
      <div class="flex items-center justify-between">
        <span class="text-xs font-semibold uppercase tracking-wider text-muted-foreground">New customer</span>
        <button type="button" class="rounded p-1 text-muted-foreground hover:text-foreground" aria-label="Cancel" @click="close">
          <X :size="14" />
        </button>
      </div>
      <input
        ref="nameInput"
        v-model="newName"
        maxlength="120"
        required
        autocomplete="off"
        class="w-full rounded-md border border-border bg-background px-2 py-1.5 text-sm"
        placeholder="Name"
        aria-label="Customer name"
      />
      <input
        v-model="newPhone"
        type="tel"
        inputmode="tel"
        maxlength="30"
        autocomplete="off"
        class="w-full rounded-md border border-border bg-background px-2 py-1.5 text-sm"
        placeholder="Phone (optional), e.g. 0803 123 4567"
        aria-label="Customer phone"
      />
      <p v-if="error" role="alert" class="rounded-md bg-destructive/10 px-2 py-1 text-xs text-destructive">{{ error }}</p>
      <div class="flex gap-2">
        <Button type="submit" size="sm" class="h-7 flex-1 text-xs" :disabled="saving">
          {{ saving ? "Saving…" : "Save and add to sale" }}
        </Button>
        <Button type="button" variant="outline" size="sm" class="h-7 text-xs" :disabled="saving" @click="startSearch">Back</Button>
      </div>
    </form>

    <!-- The customer on this sale -->
    <div v-else-if="modelValue" class="flex flex-wrap items-center gap-x-2 gap-y-1">
      <User :size="14" class="shrink-0 text-muted-foreground" aria-hidden="true" />
      <span class="min-w-[8rem] flex-1 truncate text-sm">
        <span class="font-medium">{{ modelValue.name }}</span>
        <span v-if="modelValue.phone" class="ml-2 text-xs tabular-nums text-muted-foreground">{{ showPhone(modelValue.phone) }}</span>
      </span>
      <template v-if="modelValue.active === false">
        <span class="shrink-0 rounded bg-destructive/10 px-1.5 py-0.5 text-[11px] font-medium uppercase tracking-wide text-destructive">Deactivated</span>
        <Button v-if="canManage" variant="outline" size="sm" class="h-7 shrink-0 text-xs" :disabled="disabled" @click="reactivate(modelValue)">Reactivate</Button>
      </template>
      <Button v-if="canLookup" variant="ghost" size="sm" class="h-7 text-xs" :disabled="disabled" @click="startSearch">Change</Button>
      <button
        type="button"
        class="rounded p-1 text-muted-foreground hover:text-foreground disabled:opacity-50"
        aria-label="Remove customer from this sale"
        :disabled="disabled"
        @click="remove"
      >
        <X :size="14" />
      </button>
    </div>

    <!-- No customer: a walk-in -->
    <div v-else class="flex items-center gap-2">
      <User :size="14" class="shrink-0 text-muted-foreground" aria-hidden="true" />
      <span class="flex-1 text-sm text-muted-foreground">Walk-in customer</span>
      <Button variant="outline" size="sm" class="h-7 text-xs" :disabled="disabled" @click="startSearch">Add customer</Button>
    </div>
  </div>
</template>

<script setup>
import { computed, nextTick, onBeforeUnmount, ref, watch } from "vue";
import { Search, User, X } from "lucide-vue-next";
import { Button } from "@/components/ui/button";
import { csrfFetch } from "../../lib/csrf.js";
import { readError } from "../../lib/api.js";
import { showPhone } from "../../lib/phone.js";
import { usePermissions } from "../../composables/usePermissions.js";
import { useConfirm } from "../../composables/useConfirm.js";

// The customer on the sale. modelValue is {id, name, phone, active} or null (a walk-in). Seeing customers
// (customer:view) is needed to pick one; adding one at the till needs customer:manage. Where a new
// customer makes no sense (changing the customer on a saved sale), allowCreate is false.
const props = defineProps({
  modelValue: { type: Object, default: null },
  disabled: { type: Boolean, default: false },
  allowCreate: { type: Boolean, default: true },
  // True when picking for a sale that is already saved: the pick is only a choice until the
  // person saves it, but a reactivation happens at once.
  savedSale: { type: Boolean, default: false },
});
const emit = defineEmits(["update:modelValue"]);

const { hasPermission } = usePermissions();
const canLookup = hasPermission("customer:view");
const canManage = hasPermission("customer:manage");
const canCreate = computed(() => props.allowCreate && canManage);
const { confirm } = useConfirm();

const mode = ref("idle"); // idle | search | new
const query = ref("");
const results = ref([]);
const searching = ref(false);
const searched = ref(false);
const error = ref("");
const searchInput = ref(null);
const nameInput = ref(null);
const newName = ref("");
const newPhone = ref("");
const saving = ref(false);
const noMatch = computed(() => searched.value && !searching.value && !results.value.length && query.value.trim().length >= 2);

// A slow reply to an earlier keystroke must not replace the answer to the latest one.
let searchCount = 0;
// Each save takes a number and reset() moves it on, so a save that was cancelled, or overtaken by
// the sale being completed or cleared, cannot attach its customer to whatever is on screen later.
let createCount = 0;
let timer = null;
async function search() {
  const text = query.value.trim();
  const mine = ++searchCount;
  if (text.length < 2) {
    results.value = [];
    searched.value = false;
    searching.value = false;
    return;
  }
  searching.value = true;
  try {
    // Deactivated customers are listed too, marked, so the cashier learns it before saving a sale.
    const r = await fetch(`/customers/?q=${encodeURIComponent(text)}&inactive=1`);
    if (!r.ok) throw new Error((await readError(r, "Could not search customers.")).message);
    const body = await r.json();
    if (mine !== searchCount) return;
    // Active customers first, so deactivated ones never crowd them out of the short list.
    const found = body.data || [];
    results.value = [...found.filter((c) => c.active !== false), ...found.filter((c) => c.active === false)];
    searched.value = true;
  } catch (e) {
    if (mine === searchCount) error.value = e.message || "Could not search customers.";
  } finally {
    if (mine === searchCount) searching.value = false;
  }
}

// The sale moved on (completed, cleared, a held sale resumed) or is being saved: an open search
// would be left over for the next customer, or change who the receipt names.
// The token is also moved on when the picker is idle (the attached customer's own Reactivate), where
// close() does nothing: a reactivation started for one sale must not attach to the next.
watch(() => props.modelValue?.id, () => {
  createCount++;
  close();
});
watch(() => props.disabled, (disabled) => {
  if (!disabled) return;
  createCount++;
  close();
});

watch(query, () => {
  // Until the search for the new text comes back, the last answer says nothing about it.
  searched.value = false;
  clearTimeout(timer);
  timer = setTimeout(search, 250);
});
onBeforeUnmount(() => clearTimeout(timer));

function startSearch() {
  if (props.disabled) return;
  error.value = "";
  mode.value = "search";
  nextTick(() => searchInput.value?.focus());
}

function reset() {
  clearTimeout(timer);
  searchCount++;
  createCount++;
  saving.value = false;
  query.value = "";
  results.value = [];
  searched.value = false;
  searching.value = false;
  error.value = "";
  newName.value = "";
  newPhone.value = "";
}

function close() {
  if (mode.value === "idle") return;
  reset();
  mode.value = "idle";
}

function pick(c) {
  if (props.disabled) return;
  if (c.active === false) return;
  emit("update:modelValue", { id: c.id, name: c.name, phone: c.phone || "", active: true });
  reset();
  mode.value = "idle";
}

// Enter takes the match only when it is the one and only match for what is in the box now: with
// several, or with results from an earlier keystroke, it could attach the wrong person.
function pickFirst() {
  const only = results.value[0];
  if (searched.value && !searching.value && results.value.length === 1 && only.active !== false) pick(only);
}

// Reactivating takes the customer's current version from the server, so a row that was loaded a
// while ago still works. The customer joins the sale only if the picker is still on the same sale:
// closing it, or the sale completing, in the meantime reactivates without attaching.
async function reactivate(c) {
  if (props.disabled) return;
  const mine = createCount;
  await confirm({
    title: "Reactivate customer?",
    message: props.savedSale
      ? "They will be reactivated now and chosen for this sale. The sale itself changes only when you save."
      : "They will be reactivated and added to this sale.",
    detail: c.name,
    tone: "warning",
    confirmLabel: props.savedSale ? "Reactivate and choose" : "Reactivate and add",
    busyLabel: "Saving…",
    run: async () => {
      let r;
      try {
        r = await fetch(`/customers/${c.id}`);
        if (!r.ok) throw new Error((await readError(r, "Could not read the customer.")).message);
        const current = await r.json();
        if (current.active === false) {
          r = await csrfFetch(`/customers/${c.id}/reactivate`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ version: current.version }),
          });
          if (!r.ok) throw new Error((await readError(r, "Could not reactivate the customer.")).message);
        }
        r = await fetch(`/customers/${c.id}`);
        if (!r.ok) throw new Error((await readError(r, "Could not read the customer.")).message);
      } catch (e) {
        if (e instanceof TypeError) throw new Error("Could not reach the server. Check the connection and try again.");
        throw e;
      }
      const saved = await r.json();
      if (mine === createCount) pick({ ...saved, active: true });
    },
  });
}

function remove() {
  emit("update:modelValue", null);
}

function startNew() {
  const text = query.value.trim();
  // Letters mean a name was typed; otherwise it was the start of a phone number.
  const isName = /\p{L}/u.test(text);
  newName.value = isName ? text : "";
  newPhone.value = isName ? "" : text;
  error.value = "";
  mode.value = "new";
  nextTick(() => nameInput.value?.focus());
}

async function create() {
  if (saving.value) return;
  const name = newName.value.trim();
  if (!name) {
    error.value = "Enter the customer's name.";
    return;
  }
  const mine = ++createCount;
  const stale = () => mine !== createCount;
  saving.value = true;
  error.value = "";
  try {
    const r = await csrfFetch("/customers/", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ version: 0, name, phone: newPhone.value, email: "", notes: "", contact_consent: false }),
    });
    if (r.ok) {
      const { id } = await r.json();
      // Read it back so the phone is shown the way the server stored it.
      const saved = await fetch(`/customers/${id}`).then((x) => (x.ok ? x.json() : null)).catch(() => null);
      if (stale()) return;
      pick(saved || { id, name, phone: "" });
      return;
    }
    const reply = await readError(r, "Could not save the customer.");
    if (stale()) return;
    if (reply.code === "phone_taken") {
      // Someone already has this number: show them so they can be picked instead (or, when
      // deactivated, reactivated).
      const phone = newPhone.value.trim();
      mode.value = "search";
      query.value = phone;
      error.value = reply.message.charAt(0).toUpperCase() + reply.message.slice(1) + ".";
      nextTick(() => searchInput.value?.focus());
      return;
    }
    error.value = reply.message;
  } catch {
    if (!stale()) error.value = "Could not reach the server. Check the connection and try again.";
  } finally {
    if (!stale()) saving.value = false;
  }
}
</script>
