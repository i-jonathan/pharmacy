<template>
  <div v-if="!canView" class="p-4 sm:p-6 lg:p-8">
    <h1 class="text-2xl font-semibold tracking-tight">Customers</h1>
    <p class="mt-2 text-sm text-muted-foreground">
      You do not have permission to see customers. Ask an administrator for access.
    </p>
  </div>

  <div v-else class="flex min-h-[calc(100dvh-3.5rem)] flex-col lg:h-[calc(100dvh-3.5rem)] lg:min-h-0 lg:flex-row">
    <!-- List -->
    <div class="flex min-w-0 flex-1 flex-col overflow-hidden" :class="{ 'border-border lg:border-r': panelOpen }">
      <div class="flex flex-wrap items-start justify-between gap-3 px-4 pb-0 pt-5 sm:px-6">
        <div>
          <h1 class="text-2xl font-bold text-foreground">Customers</h1>
          <p class="mt-1 text-sm text-muted-foreground">People the pharmacy keeps on file</p>
        </div>
        <Button v-if="canManage" class="gap-2" :disabled="saving" @click="startNew">
          <UserPlus :size="16" />
          Add customer
        </Button>
      </div>

      <div
        v-if="notice"
        role="status"
        class="mx-4 mt-4 rounded-md border px-3 py-2 text-sm sm:mx-6"
        :class="notice.error ? 'border-destructive/30 bg-destructive/5 text-destructive' : 'border-border bg-muted/40'"
      >
        {{ notice.text }}
      </div>

      <div class="flex flex-wrap items-center gap-3 px-4 py-4 sm:px-6">
        <div class="relative w-full sm:w-80">
          <Search :size="15" class="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
          <input
            v-model="search"
            type="search"
            class="w-full rounded-md border border-border bg-background py-2 pl-9 pr-3 text-sm"
            placeholder="Search by name or phone"
            aria-label="Search customers by name or phone"
          />
        </div>
        <label class="flex items-center gap-2 text-sm text-muted-foreground">
          <input v-model="showInactive" type="checkbox" class="h-3.5 w-3.5" />
          Show deactivated
        </label>
        <span v-if="!loading && !loadError" class="text-xs text-muted-foreground">
          {{ total }} customer{{ total !== 1 ? "s" : "" }}
        </span>
      </div>

      <div class="min-h-0 flex-1 overflow-y-auto px-4 pb-4 sm:px-6">
        <div v-if="loading && !customers.length" class="flex items-center justify-center py-24 text-muted-foreground">
          <RotateCw :size="20" class="mr-3 animate-spin" />
          <span class="text-sm">Loading customers...</span>
        </div>

        <div v-else-if="loadError" class="flex flex-col items-center justify-center py-24 text-center">
          <AlertCircle :size="40" class="mb-3 text-destructive/40" />
          <p class="mb-3 text-sm text-muted-foreground">{{ loadError }}</p>
          <Button variant="outline" size="sm" @click="load">Retry</Button>
        </div>

        <div v-else-if="!customers.length" class="flex flex-col items-center justify-center py-24 text-center">
          <Users :size="40" class="mb-3 text-muted-foreground/40" />
          <h3 class="mb-1 text-base font-semibold text-foreground">
            {{ search.trim() ? "No customers match" : "No customers yet" }}
          </h3>
          <p class="max-w-sm text-sm text-muted-foreground">
            {{ search.trim()
              ? "Check the spelling, or try part of the phone number."
              : "Customers you add will be listed here." }}
          </p>
        </div>

        <div v-else class="overflow-x-auto rounded-lg border border-border">
          <table class="w-full text-sm">
            <thead>
              <tr class="border-b border-border bg-muted/30">
                <th scope="col" class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-muted-foreground">Name</th>
                <th scope="col" class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-muted-foreground">Phone</th>
                <th scope="col" class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-muted-foreground" :class="panelOpen ? 'hidden' : 'hidden md:table-cell'">Email</th>
                <th scope="col" class="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-muted-foreground">Contact OK</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="c in customers"
                :key="c.id"
                class="cursor-pointer border-b border-border last:border-0 hover:bg-muted/20"
                :class="{ 'bg-primary/5': form && form.id === c.id }"
              >
                <td class="px-4 py-3">
                  <button type="button" class="text-left font-medium hover:underline focus-visible:underline focus-visible:outline-none disabled:cursor-wait disabled:opacity-60" :disabled="saving" @click="open(c)">
                    {{ c.name }}
                  </button>
                  <span v-if="!c.active" class="ml-2 rounded-full bg-muted px-2 py-0.5 text-[10px] font-semibold text-muted-foreground">Deactivated</span>
                </td>
                <td class="whitespace-nowrap px-4 py-3 tabular-nums text-muted-foreground">{{ showPhone(c.phone) }}</td>
                <td class="px-4 py-3 text-muted-foreground" :class="panelOpen ? 'hidden' : 'hidden md:table-cell'">{{ c.email || "—" }}</td>
                <td class="px-4 py-3 text-muted-foreground">{{ c.contact_consent ? "Yes" : "No" }}</td>
              </tr>
            </tbody>
          </table>
        </div>

        <div class="mt-3">
          <PagerBar v-model:page="page" :total="total" :size="perPage" />
        </div>
      </div>
    </div>

    <!-- Add / edit -->
    <aside
      v-if="panelOpen"
      ref="panel"
      class="flex w-full max-w-full flex-col border-t border-border bg-card lg:h-full lg:w-[40%] lg:min-w-[360px] lg:max-w-[520px] lg:flex-shrink-0 lg:border-t-0"
      aria-label="Customer details"
    >
      <div class="flex items-center justify-between border-b border-border px-4 py-3">
        <div>
          <h2 class="text-sm font-semibold">{{ form.id ? (canManage ? "Edit customer" : "Customer") : "New customer" }}</h2>
          <p v-if="form.id && !form.active" class="text-xs text-muted-foreground">Deactivated</p>
        </div>
        <Button variant="ghost" size="icon" class="h-7 w-7 text-muted-foreground" aria-label="Close" @click="closePanel">
          <X :size="15" />
        </Button>
      </div>

      <form class="flex min-h-0 flex-1 flex-col" @submit.prevent="save">
        <fieldset class="min-h-0 flex-1 space-y-4 overflow-y-auto px-4 py-4" :disabled="!canManage || saving">
          <div>
            <label for="customer-name" class="text-xs font-semibold uppercase tracking-wider text-muted-foreground">Name</label>
            <input
              id="customer-name"
              ref="nameInput"
              v-model="form.name"
              maxlength="120"
              required
              autocomplete="off"
              class="mt-1 w-full rounded-md border border-border bg-background px-3 py-2 text-sm disabled:opacity-70"
            />
          </div>
          <div>
            <label for="customer-phone" class="text-xs font-semibold uppercase tracking-wider text-muted-foreground">Phone</label>
            <input
              id="customer-phone"
              v-model="form.phone"
              type="tel"
              inputmode="tel"
              maxlength="30"
              autocomplete="off"
              placeholder="0803 123 4567"
              aria-describedby="customer-phone-help"
              class="mt-1 w-full rounded-md border border-border bg-background px-3 py-2 text-sm disabled:opacity-70"
            />
            <p id="customer-phone-help" class="mt-1 text-xs text-muted-foreground">
              Optional. Each phone number can belong to one customer.
            </p>
          </div>
          <div>
            <label for="customer-email" class="text-xs font-semibold uppercase tracking-wider text-muted-foreground">Email</label>
            <input
              id="customer-email"
              v-model="form.email"
              type="email"
              maxlength="254"
              autocomplete="off"
              class="mt-1 w-full rounded-md border border-border bg-background px-3 py-2 text-sm disabled:opacity-70"
            />
          </div>
          <div>
            <label for="customer-notes" class="text-xs font-semibold uppercase tracking-wider text-muted-foreground">Notes</label>
            <textarea
              id="customer-notes"
              v-model="form.notes"
              rows="3"
              maxlength="1000"
              class="mt-1 w-full rounded-md border border-border bg-background px-3 py-2 text-sm disabled:opacity-70"
            />
          </div>
          <div class="rounded-md border border-border bg-muted/30 px-3 py-3">
            <label class="flex items-start gap-2 text-sm">
              <input v-model="form.contact_consent" type="checkbox" class="mt-0.5 h-4 w-4" />
              <span>Agreed to be contacted by the pharmacy</span>
            </label>
            <p class="mt-1 pl-6 text-xs text-muted-foreground">
              Tick this only if the customer said so. Untick it if they ask not to be contacted.
              <span v-if="form.id && form.consent_updated_at">
                Last changed {{ showDate(form.consent_updated_at) }}.
              </span>
            </p>
          </div>
        </fieldset>

        <p v-if="formError" role="alert" class="mx-4 mb-3 rounded-md border border-destructive/30 bg-destructive/5 px-3 py-2 text-sm text-destructive">
          {{ formError }}
        </p>

        <div v-if="canManage" class="flex items-center gap-2 border-t border-border px-4 py-3">
          <Button type="submit" class="flex-1" :disabled="saving">
            {{ saving ? "Saving…" : form.id ? "Save changes" : "Add customer" }}
          </Button>
          <Button
            v-if="form.id"
            type="button"
            variant="outline"
            :disabled="saving"
            @click="toggleActive"
          >
            {{ form.active ? "Deactivate" : "Reactivate" }}
          </Button>
        </div>
      </form>
    </aside>
  </div>
</template>

<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { AlertCircle, RotateCw, Search, UserPlus, Users, X } from "lucide-vue-next";
import { Button } from "@/components/ui/button";
import PagerBar from "../components/PagerBar.vue";
import { csrfFetch } from "../lib/csrf.js";
import { readError } from "../lib/api.js";
import { useConfirm } from "../composables/useConfirm.js";
import { usePermissions } from "../composables/usePermissions.js";

const { hasPermission } = usePermissions();
const canView = hasPermission("customer:view");
const canManage = hasPermission("customer:manage");
const { confirm } = useConfirm();

const customers = ref([]);
const total = ref(0);
const perPage = ref(25);
const page = ref(1);
const search = ref("");
const showInactive = ref(false);
const loading = ref(false);
const loadError = ref("");
const notice = ref(null);
let noticeTimer = null;

// form is the customer open in the side panel, or null; form.id is 0 for a new one.
const form = ref(null);
const saving = ref(false);
const formError = ref("");
const nameInput = ref(null);
const panel = ref(null);
const panelOpen = computed(() => !!form.value);

// A slow reply to an earlier search must not replace the answer to the latest one.
let loadCount = 0;
async function load() {
  const mine = ++loadCount;
  loading.value = true;
  loadError.value = "";
  try {
    const params = new URLSearchParams({ page: String(page.value) });
    if (search.value.trim()) params.set("q", search.value.trim());
    if (showInactive.value) params.set("inactive", "1");
    const r = await fetch(`/customers/?${params}`);
    if (!r.ok) throw new Error((await readError(r, `Server error (${r.status})`)).message);
    const body = await r.json();
    if (mine !== loadCount) return;
    customers.value = body.data || [];
    total.value = body.total_count || 0;
    perPage.value = body.per_page || 25;
    // The last customer on the last page went (deactivated): go to the page that is now last.
    if (!customers.value.length && total.value > 0 && page.value > 1) {
      page.value = Math.ceil(total.value / perPage.value);
    }
  } catch (e) {
    if (mine === loadCount) loadError.value = e.message || "Could not load customers.";
  } finally {
    if (mine === loadCount) loading.value = false;
  }
}

let searchTimer = null;
watch([search, showInactive], () => {
  clearTimeout(searchTimer);
  searchTimer = setTimeout(() => {
    if (page.value === 1) load();
    else page.value = 1; // the page watcher loads
  }, 250);
});
watch(page, load);
onMounted(() => {
  if (canView) load();
});
onBeforeUnmount(() => {
  clearTimeout(searchTimer);
  clearTimeout(noticeTimer);
});

function showNotice(text, error = false) {
  notice.value = { text, error };
  clearTimeout(noticeTimer);
  noticeTimer = setTimeout(() => (notice.value = null), error ? 8000 : 4000);
}

// +2348031234567 → 0803 123 4567; anything else (a foreign number) as stored.
function showPhone(phone) {
  const m = /^\+234(\d{3})(\d{3})(\d{4})$/.exec(phone || "");
  if (m) return `0${m[1]} ${m[2]} ${m[3]}`;
  return phone || "—";
}

function showDate(iso) {
  return new Date(iso).toLocaleDateString("en-GB", { day: "numeric", month: "short", year: "numeric" });
}

function formFrom(c) {
  return {
    id: c.id,
    version: c.version,
    name: c.name,
    // Shown the way it is usually written; the server normalizes it again on save.
    phone: c.phone ? showPhone(c.phone) : "",
    email: c.email,
    notes: c.notes,
    contact_consent: c.contact_consent,
    consent_updated_at: c.consent_updated_at,
    active: c.active,
  };
}

// On a phone the panel opens below the list, out of sight.
function showPanel() {
  nextTick(() => {
    if (window.innerWidth < 1024) panel.value?.scrollIntoView({ block: "start", behavior: "smooth" });
  });
}

function open(c) {
  formError.value = "";
  form.value = formFrom(c);
  showPanel();
}

function startNew() {
  formError.value = "";
  form.value = {
    id: 0, version: 0, name: "", phone: "", email: "", notes: "",
    contact_consent: false, consent_updated_at: null, active: true,
  };
  nextTick(() => nameInput.value?.focus());
  showPanel();
}

function closePanel() {
  form.value = null;
  formError.value = "";
}

async function save() {
  if (saving.value || !canManage) return;
  const f = form.value;
  if (!f.name.trim()) {
    formError.value = "Enter the customer's name.";
    return;
  }
  const body = {
    version: f.version,
    name: f.name,
    phone: f.phone,
    email: f.email,
    notes: f.notes,
    contact_consent: f.contact_consent,
  };
  saving.value = true;
  formError.value = "";
  try {
    const r = await csrfFetch(f.id ? `/customers/${f.id}` : "/customers/", {
      method: f.id ? "PUT" : "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    });
    if (r.ok) {
      showNotice(f.id ? `Saved ${f.name.trim()}.` : `Added ${f.name.trim()}.`);
      if (form.value === f) closePanel();
      await load();
      return;
    }
    const reply = await readError(r, "Could not save the customer.");
    formError.value = reply.message;
    // Someone else changed this customer first: show what is saved now, keep the message.
    if (r.status === 409 && reply.code !== "phone_taken" && f.id) await refreshOpenCustomer(reply.message);
  } catch {
    formError.value = "Could not reach the server. Check the connection, then check the list before trying again.";
  } finally {
    saving.value = false;
  }
}

// Someone else saved this customer first. Take their version number (and anything else the page
// shows from the saved record) but keep what the person typed, so a minute of work isn't lost.
async function refreshOpenCustomer(message) {
  const id = form.value?.id;
  if (!id) return;
  try {
    const r = await fetch(`/customers/${id}`);
    if (r.ok && form.value?.id === id) {
      const latest = await r.json();
      form.value = {
        ...form.value,
        version: latest.version,
        consent_updated_at: latest.consent_updated_at,
        active: latest.active,
      };
      formError.value = `${message.replace(/[.\s]+$/, "")}. Your changes are still in the form. Save again to replace theirs, or close this panel to keep theirs.`;
    }
  } catch {
    // The message already says to reload.
  }
  load();
}

function toggleActive() {
  const f = form.value;
  if (!f?.id || saving.value) return;
  const stopping = f.active;
  return confirm({
    title: stopping ? "Deactivate customer?" : "Reactivate customer?",
    message: stopping
      ? "They will no longer show in the list or be found at the till. Their record and phone number are kept."
      : "They will show in the list again.",
    detail: f.name,
    tone: "warning",
    confirmLabel: stopping ? "Deactivate" : "Reactivate",
    busyLabel: "Saving…",
    run: async () => {
      let r;
      try {
        r = await csrfFetch(`/customers/${f.id}/${stopping ? "deactivate" : "reactivate"}`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ version: f.version }),
        });
      } catch {
        throw new Error("Could not reach the server. Check the list before trying again.");
      }
      if (!r.ok) {
        const reply = await readError(r, "Could not save the customer.");
        if (r.status === 409 && form.value === f) await refreshOpenCustomer(reply.message);
        throw new Error(reply.message);
      }
      showNotice(`${f.name} ${stopping ? "deactivated" : "reactivated"}.`);
      if (form.value === f) closePanel();
      await load();
    },
  });
}
</script>
