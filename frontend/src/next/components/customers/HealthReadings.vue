<template>
  <div class="flex min-h-0 flex-1 flex-col">
    <div class="flex items-center justify-between gap-2 px-4 pt-3">
      <p class="text-xs text-muted-foreground">
        Blood pressure and glucose, newest first. Opening this list is recorded.
      </p>
      <Button variant="ghost" size="sm" class="h-7 gap-1.5 text-xs" :disabled="loading || !!editing" @click="load">
        <RotateCw :size="13" :class="{ 'animate-spin': loading }" />
        Refresh
      </Button>
    </div>

    <p
      v-if="notice"
      role="status"
      class="mx-4 mt-3 rounded-md border px-3 py-2 text-sm"
      :class="notice.error ? 'border-destructive/30 bg-destructive/5 text-destructive' : 'border-border bg-muted/40'"
    >
      {{ notice.text }}
    </p>

    <div class="min-h-0 flex-1 space-y-3 overflow-y-auto px-4 py-3">
      <div v-if="loading && !readings.length" class="flex items-center justify-center py-16 text-muted-foreground">
        <RotateCw :size="18" class="mr-2 animate-spin" />
        <span class="text-sm">Loading readings...</span>
      </div>

      <div v-else-if="loadError" class="flex flex-col items-center py-16 text-center">
        <AlertCircle :size="32" class="mb-2 text-destructive/40" />
        <p class="mb-3 text-sm text-muted-foreground">{{ loadError }}</p>
        <Button variant="outline" size="sm" @click="load">Retry</Button>
      </div>

      <div v-else-if="!readings.length" class="flex flex-col items-center py-16 text-center">
        <HeartPulse :size="32" class="mb-2 text-muted-foreground/40" />
        <p class="text-sm text-muted-foreground">No readings saved for this customer.</p>
      </div>

      <template v-else>
        <article
          v-for="r in readings"
          :key="r.id"
          class="rounded-lg border border-border bg-background p-3"
          :aria-label="`Reading of ${when(r.taken_at)}`"
        >
          <!-- Correction form, in place of the reading -->
          <form v-if="editing && editing.id === r.id" class="space-y-3" novalidate @submit.prevent="saveCorrection">
            <div class="flex items-center justify-between">
              <h3 class="text-sm font-semibold">Correct the reading of {{ when(r.taken_at) }}</h3>
            </div>

            <fieldset class="space-y-3" :disabled="saving">
              <div class="rounded-md border border-border p-3">
                <label class="flex items-center gap-2 text-sm font-medium">
                  <input v-model="editing.bpOn" type="checkbox" class="h-4 w-4" />
                  Blood pressure
                </label>
                <div v-if="editing.bpOn" class="mt-2 grid grid-cols-3 gap-2">
                  <label class="text-xs">
                    <span class="text-muted-foreground">Top</span>
                    <input v-model="editing.systolic" v-bind="numberInput" />
                  </label>
                  <label class="text-xs">
                    <span class="text-muted-foreground">Bottom</span>
                    <input v-model="editing.diastolic" v-bind="numberInput" />
                  </label>
                  <label class="text-xs">
                    <span class="text-muted-foreground">Pulse</span>
                    <input v-model="editing.pulse" v-bind="numberInput" placeholder="Optional" />
                  </label>
                </div>
              </div>

              <div class="rounded-md border border-border p-3">
                <label class="flex items-center gap-2 text-sm font-medium">
                  <input v-model="editing.glucoseOn" type="checkbox" class="h-4 w-4" />
                  Blood glucose
                </label>
                <div v-if="editing.glucoseOn" class="mt-2 space-y-2">
                  <div class="grid grid-cols-2 gap-2">
                    <label class="text-xs">
                      <span class="text-muted-foreground">Reading</span>
                      <input v-model="editing.glucose" v-bind="numberInput" inputmode="decimal" />
                    </label>
                    <label class="text-xs">
                      <span class="text-muted-foreground">Unit</span>
                      <select v-model="editing.glucoseUnit" class="mt-1 w-full rounded-md border border-border bg-background px-2 py-1.5 text-sm">
                        <option value="mmol/L">mmol/L</option>
                        <option value="mg/dL">mg/dL</option>
                      </select>
                    </label>
                  </div>
                  <fieldset>
                    <legend class="text-xs text-muted-foreground">When taken</legend>
                    <div class="mt-1 flex flex-wrap gap-1.5">
                      <label
                        v-for="c in GLUCOSE_CONTEXTS"
                        :key="c.value"
                        class="flex cursor-pointer items-center gap-1.5 rounded-md border px-2 py-1 text-xs"
                        :class="editing.glucoseContext === c.value ? 'border-foreground bg-foreground/5' : 'border-border'"
                      >
                        <input v-model="editing.glucoseContext" type="radio" :name="`correct-context-${r.id}`" :value="c.value" class="h-3.5 w-3.5" />
                        {{ c.label }}
                      </label>
                    </div>
                  </fieldset>
                </div>
              </div>

              <label class="block text-xs">
                <span class="text-muted-foreground">Note (optional)</span>
                <input v-model="editing.note" maxlength="500" autocomplete="off" class="mt-1 w-full rounded-md border border-border bg-background px-2 py-1.5 text-sm" />
              </label>

              <div>
                <p class="mb-1 text-xs text-muted-foreground">Why is it being corrected?</p>
                <ReasonPicker
                  v-model:preset="editing.reasonPreset"
                  v-model:detail="editing.reasonDetail"
                  :reasons="CORRECTION_REASONS"
                />
              </div>
            </fieldset>

            <p v-if="formError" role="alert" class="rounded-md bg-destructive/10 px-3 py-2 text-sm text-destructive">{{ formError }}</p>

            <div class="flex gap-2">
              <Button type="submit" size="sm" :disabled="saving">{{ saving ? "Saving…" : "Save correction" }}</Button>
              <Button type="button" size="sm" variant="outline" :disabled="saving" @click="cancelCorrection">Cancel</Button>
            </div>
          </form>

          <!-- The reading -->
          <template v-else>
            <div class="flex flex-wrap items-baseline justify-between gap-x-3 gap-y-1">
              <p class="text-sm font-semibold tabular-nums">{{ when(r.taken_at) }}</p>
              <span v-if="r.corrected" class="rounded-full bg-muted px-2 py-0.5 text-[10px] font-semibold text-muted-foreground">
                Corrected {{ day(r.updated_at) }}
              </span>
            </div>
            <dl class="mt-2 space-y-1 text-sm">
              <div v-if="r.systolic !== null" class="flex gap-2">
                <dt class="w-28 shrink-0 text-muted-foreground">Blood pressure</dt>
                <dd class="tabular-nums">
                  {{ r.systolic }} / {{ r.diastolic }} mmHg<span v-if="r.pulse !== null">, pulse {{ r.pulse }} bpm</span>
                </dd>
              </div>
              <div v-if="r.glucose !== null" class="flex gap-2">
                <dt class="w-28 shrink-0 text-muted-foreground">Blood glucose</dt>
                <dd class="tabular-nums">{{ describeGlucose(r) }}</dd>
              </div>
              <div v-if="r.note" class="flex gap-2">
                <dt class="w-28 shrink-0 text-muted-foreground">Note</dt>
                <dd class="min-w-0 break-words">{{ r.note }}</dd>
              </div>
            </dl>
            <div class="mt-2 flex flex-wrap items-center justify-between gap-2">
              <p class="text-xs text-muted-foreground">Taken by {{ r.taken_by }}</p>
              <div class="flex gap-1.5">
                <Button size="sm" variant="outline" class="h-7 gap-1.5 text-xs" :disabled="printingId !== null || saving || loading" @click="reprint(r)">
                  <Printer :size="13" />
                  {{ printingId === r.id ? "Printing…" : "Reprint" }}
                </Button>
                <Button v-if="canCorrect" size="sm" variant="outline" class="h-7 gap-1.5 text-xs" :disabled="saving || loading" @click="startCorrection(r)">
                  <Pencil :size="13" />
                  Correct
                </Button>
              </div>
            </div>
          </template>
        </article>

        <PagerBar :page="page" :total="total" :size="perPage" @update:page="goToPage" />
      </template>
    </div>
  </div>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, ref } from "vue";
import { AlertCircle, HeartPulse, Pencil, Printer, RotateCw } from "lucide-vue-next";
import { Button } from "@/components/ui/button";
import PagerBar from "../PagerBar.vue";
import ReasonPicker from "../ReasonPicker.vue";
import { csrfFetch } from "../../lib/csrf.js";
import { readError } from "../../lib/api.js";
import { GLUCOSE_CONTEXTS, printHealthSlip, slipFromSaved } from "../../lib/healthSlip.js";
import { describeGlucose, glucoseText, readingsFrom } from "../../lib/healthReadings.js";
import { usePermissions } from "../../composables/usePermissions.js";
import { useConfirm } from "../../composables/useConfirm.js";

// The customer's saved readings (healthcheck:view), with reprint and, for healthcheck:correct,
// correction. The parent mounts this only once someone opens the tab, because every load of the
// list is recorded in the audit log.
const props = defineProps({
  customer: { type: Object, required: true },
});

const { hasPermission } = usePermissions();
const canCorrect = hasPermission("healthcheck:correct");
const { confirm } = useConfirm();

const CORRECTION_REASONS = ["Typed wrong", "Wrong unit", "Wrong time of glucose test", "Other"];
const MAX_REASON = 200;

const numberInput = {
  type: "text",
  inputmode: "numeric",
  autocomplete: "off",
  maxlength: 6,
  class: "mt-1 w-full rounded-md border border-border bg-background px-2 py-1.5 text-sm tabular-nums",
};

const readings = ref([]);
const page = ref(1);
const total = ref(0);
const perPage = ref(20);
const loading = ref(false);
const loadError = ref("");
const notice = ref(null);
const printingId = ref(null);
let noticeTimer = null;

// editing is the open correction form, or null. openedAs is what it held when opened.
const editing = ref(null);
const saving = ref(false);
const formError = ref("");
let openedAs = "";
const editable = (e) =>
  JSON.stringify([
    e.bpOn, e.systolic, e.diastolic, e.pulse, e.glucoseOn, e.glucose, e.glucoseUnit,
    e.glucoseContext, e.note, e.reasonPreset, e.reasonDetail,
  ]);
const dirty = computed(() => !!editing.value && editable(editing.value) !== openedAs);

function when(iso) {
  const d = new Date(iso);
  return Number.isNaN(d.getTime()) ? "" : d.toLocaleString("en-GB", { dateStyle: "medium", timeStyle: "short" });
}

function day(iso) {
  const d = new Date(iso);
  return Number.isNaN(d.getTime()) ? "" : d.toLocaleDateString("en-GB", { day: "numeric", month: "short", year: "numeric" });
}

function showNotice(text, error = false) {
  notice.value = { text, error };
  clearTimeout(noticeTimer);
  noticeTimer = setTimeout(() => (notice.value = null), error ? 10000 : 4000);
}

function sentence(text) {
  const t = String(text || "").trim();
  return t ? t.charAt(0).toUpperCase() + t.slice(1) + (/[.!?]$/.test(t) ? "" : ".") : "";
}

// A slow reply for an earlier page must not replace the latest one.
let loadCount = 0;
async function load() {
  const mine = ++loadCount;
  loading.value = true;
  loadError.value = "";
  try {
    const r = await fetch(`/health-checks/customer/${props.customer.id}?page=${page.value}`);
    if (!r.ok) throw new Error(sentence((await readError(r, "Could not load the readings.")).message));
    const body = await r.json();
    if (mine !== loadCount) return;
    readings.value = body.data || [];
    // A correction open on a reading that is no longer listed has nothing to show it.
    if (editing.value && !readings.value.some((r) => r.id === editing.value.id)) editing.value = null;
    total.value = body.total_count || 0;
    perPage.value = body.per_page || 20;
  } catch (e) {
    if (mine === loadCount) {
      loadError.value = e instanceof TypeError ? "Could not reach the server. Check the connection and try again." : e.message;
    }
  } finally {
    if (mine === loadCount) loading.value = false;
  }
}

async function goToPage(next) {
  if (next === page.value) return;
  if (saving.value) {
    showNotice("Wait for the correction to finish saving.", true);
    return;
  }
  if (!(await canLeaveCorrection())) return;
  editing.value = null;
  page.value = next;
  load();
}
onMounted(load);
onBeforeUnmount(() => clearTimeout(noticeTimer));

// Reprint: the server records it and answers with the reading, printed as a marked copy.
async function reprint(reading) {
  if (printingId.value !== null) return;
  printingId.value = reading.id;
  try {
    let r;
    try {
      r = await csrfFetch(`/health-checks/${reading.id}/reprint`, { method: "POST" });
    } catch {
      showNotice("Could not reach the server, so nothing was printed. Try again.", true);
      return;
    }
    if (!r.ok) {
      showNotice(sentence((await readError(r, "Could not reprint the reading.")).message), true);
      return;
    }
    let saved;
    try {
      saved = await r.json();
    } catch {
      showNotice("The reading did not arrive in full, so nothing was printed. Try again.", true);
      return;
    }
    try {
      await printHealthSlip(slipFromSaved(saved, { at: new Date() }));
      showNotice("Copy sent to print.");
    } catch {
      showNotice("The slip could not be printed. Check the printer, then press Reprint again.", true);
    }
  } finally {
    printingId.value = null;
  }
}

// formFrom fills the correction form with every field of the reading as loaded, so what is not
// touched is sent back unchanged (the server clears a field left out).
function formFrom(r) {
  const hasGlucose = r.glucose !== null && r.glucose !== undefined;
  return {
    id: r.id,
    version: r.version,
    bpOn: r.systolic !== null && r.systolic !== undefined,
    // As text, the way the inputs give it back, so retyping a number is not a change.
    systolic: String(r.systolic ?? ""),
    diastolic: String(r.diastolic ?? ""),
    pulse: String(r.pulse ?? ""),
    glucoseOn: hasGlucose,
    glucose: hasGlucose ? glucoseText(r.glucose, r.glucose_unit) : "",
    glucoseUnit: r.glucose_unit || "mmol/L",
    glucoseContext: r.glucose_context || "",
    note: r.note || "",
    reasonPreset: "",
    reasonDetail: "",
  };
}

// What the server holds for a reading, in the shape of a correction body, to tell "nothing
// changed" apart before sending.
function storedReadings(r) {
  return JSON.stringify([
    r.systolic ?? null, r.diastolic ?? null, r.pulse ?? null,
    r.glucose === null || r.glucose === undefined ? null : Math.round(r.glucose * 10),
    r.glucose_unit || "", r.glucose_context || "", (r.note || "").trim(),
  ]);
}
function sentReadings(b) {
  return JSON.stringify([
    b.systolic, b.diastolic, b.pulse,
    b.glucose === null ? null : Math.round(b.glucose * 10),
    b.glucose_unit, b.glucose_context, b.note.trim(),
  ]);
}

// canLeaveCorrection is true when it is fine to close the open correction form. It only asks:
// the caller closes the form once it really goes ahead (the panel may still stay open, for
// example when a details save then fails, and then the correction must still be there).
async function canLeaveCorrection() {
  if (saving.value) return false;
  if (!dirty.value) return true;
  return (
    (await confirm({
      title: "Discard the correction?",
      message: "The correction you started has not been saved.",
      detail: props.customer.name,
      tone: "warning",
      confirmLabel: "Discard",
      cancelLabel: "Keep editing",
    })) === true
  );
}

// closeCorrection closes the form. One whose save may have landed (lost reply) reloads the list,
// so the next correction starts from what is saved and isn't told someone else changed it.
function closeCorrection() {
  const e = editing.value;
  editing.value = null;
  formError.value = "";
  if (e?.mayHaveSaved) load();
}

async function startCorrection(r) {
  if (editing.value?.id === r.id) return;
  if (!(await canLeaveCorrection())) return;
  closeCorrection();
  editing.value = formFrom(r);
  openedAs = editable(editing.value);
}

async function cancelCorrection() {
  if (!(await canLeaveCorrection())) return;
  closeCorrection();
}

function composeReason({ reasonPreset: preset, reasonDetail }) {
  const text = (reasonDetail || "").trim();
  if (!preset) return "";
  if (preset === "Other") return text;
  return text ? `${preset}: ${text}` : preset;
}

async function saveCorrection() {
  const e = editing.value;
  if (!e || saving.value) return;
  const reading = readings.value.find((r) => r.id === e.id);
  let values;
  try {
    values = readingsFrom(e);
  } catch (err) {
    formError.value = err.message;
    return;
  }
  const body = { ...values, note: e.note.trim(), version: e.version };
  if (reading && sentReadings(body) === storedReadings(reading)) {
    formError.value = "Nothing was changed. Change a number, or press Cancel.";
    return;
  }
  if (!e.reasonPreset) {
    formError.value = "Choose why the reading is being corrected.";
    return;
  }
  body.reason = composeReason(e);
  if (!body.reason) {
    formError.value = "Describe why the reading is being corrected.";
    return;
  }
  // Counted in characters, as the server counts them (an emoji is one).
  if ([...body.reason].length > MAX_REASON) {
    formError.value = `Keep the reason to ${MAX_REASON} characters or fewer.`;
    return;
  }

  saving.value = true;
  formError.value = "";
  try {
    // A reply that may hide a save: safe to send again, since if the first one landed the version
    // has moved and the server answers 409 changed, which is then read as this person's own save.
    const lost = () => {
      e.mayHaveSaved = true;
      formError.value = "Could not reach the server, so the correction may or may not be saved. Press Save correction again; it can't be saved twice.";
    };
    let r;
    try {
      r = await csrfFetch(`/health-checks/${e.id}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      });
    } catch {
      lost();
      return;
    }
    // No 5xx says the correction failed: a proxy may have given up while it committed, and the
    // app's own 500 can come from a commit whose acknowledgement was lost.
    if (r.status >= 500) {
      lost();
      return;
    }
    if (r.ok) {
      let saved;
      try {
        saved = await r.json();
      } catch {
        // Saved, but the reply was cut off: show what is saved now.
        editing.value = null;
        await load();
        showNotice("Reading corrected. The list was reloaded to show it.");
        return;
      }
      // A reload still on its way was read before this save: drop it, or it would put the old
      // version back.
      loadCount++;
      loading.value = false;
      readings.value = readings.value.map((x) => (x.id === saved.id ? saved : x));
      editing.value = null;
      showNotice("Reading corrected.");
      return;
    }
    const { message, code } = await readError(r, "Could not correct the reading.");
    if (code === "changed") {
      editing.value = null;
      await load();
      // After a lost reply the version that moved may be this person's own correction.
      if (e.mayHaveSaved) {
        showNotice("Your correction may already have been saved. The list shows what is saved now; check it.", true);
      } else {
        showNotice("Someone corrected this reading first. The list shows what is saved now; check it and correct it again if needed.", true);
      }
      return;
    }
    if (code === "health_consent_needed") {
      formError.value = `${props.customer.name} has not agreed to the pharmacy keeping their health readings, so a measurement this reading didn't have can't be added. Record it as a new health check instead, which asks for their consent.`;
      return;
    }
    if (code === "customer_inactive") {
      formError.value = "This customer is deactivated, so a measurement this reading didn't have can't be added. The numbers already saved can still be changed.";
      return;
    }
    formError.value = sentence(message);
  } finally {
    saving.value = false;
  }
}

defineExpose({ dirty, saving, canLeaveCorrection });
</script>
