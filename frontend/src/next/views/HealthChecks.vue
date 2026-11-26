<template>
  <div v-if="!canRecord" class="p-4 sm:p-6 lg:p-8">
    <h1 class="text-2xl font-semibold tracking-tight">Health checks</h1>
    <p class="mt-2 text-sm text-muted-foreground">
      You can't record health checks. Ask a manager if you need to.
    </p>
  </div>
  <div v-else class="mx-auto max-w-2xl space-y-5 p-4 sm:p-6 lg:p-8">
    <header>
      <h1 class="text-2xl font-semibold tracking-tight">Health checks</h1>
      <p class="mt-1 text-sm text-muted-foreground">
        Record a blood pressure or blood glucose reading and print it for the customer.
      </p>
    </header>

    <p
      v-if="notice"
      class="rounded-md border px-3 py-2 text-sm"
      :class="notice.error ? 'border-destructive/30 bg-destructive/5 text-destructive' : 'border-border bg-muted/40'"
      role="status"
    >
      {{ notice.text }}
      <Button v-if="notice.reprint" variant="outline" size="sm" class="ml-2 h-7" :disabled="printing" @click="printAgain">Print again</Button>
    </p>

    <!-- Who. Outside the form: the picker has buttons and a form of its own (adding a customer),
       which must not submit, or nest inside, the reading's form. -->
    <section class="rounded-lg border border-border bg-card py-2">
      <h2 class="px-4 pt-1 text-xs font-semibold uppercase tracking-wider text-muted-foreground">Who</h2>
      <CustomerPicker ref="picker" v-model="customer" :disabled="locked" for-what="reading" />
      <p v-if="!canLookup" class="px-4 py-1.5 text-sm text-muted-foreground">Walk-in customer</p>
      <div v-if="!customer" class="px-4 pb-2 pt-1">
        <label class="block text-sm">
          <span class="text-muted-foreground">Name on the slip (optional)</span>
          <input
            v-model="walkInName"
            maxlength="60"
            autocomplete="off"
            class="mt-1 w-full rounded-md border border-border bg-background px-3 py-2 text-sm"
            :disabled="locked"
          />
        </label>
        <p class="mt-1 text-xs text-muted-foreground">A walk-in's reading is printed and not kept.</p>
      </div>
    </section>

    <form class="space-y-4" novalidate @submit.prevent="submit">
      <!-- Blood pressure -->
      <section class="space-y-3 rounded-lg border border-border bg-card p-4">
        <label class="flex items-center gap-2 text-sm font-semibold">
          <input v-model="bpOn" type="checkbox" class="h-4 w-4" :disabled="locked" />
          Blood pressure
        </label>
        <div v-if="bpOn" class="grid grid-cols-3 gap-3">
          <label class="text-sm">
            <span class="text-muted-foreground">Top (systolic)</span>
            <input v-model="systolic" v-bind="numberInput" aria-describedby="bp-unit" :disabled="locked" />
          </label>
          <label class="text-sm">
            <span class="text-muted-foreground">Bottom (diastolic)</span>
            <input v-model="diastolic" v-bind="numberInput" aria-describedby="bp-unit" :disabled="locked" />
          </label>
          <label class="text-sm">
            <span class="text-muted-foreground">Pulse (optional)</span>
            <input v-model="pulse" v-bind="numberInput" :disabled="locked" />
          </label>
          <p id="bp-unit" class="col-span-3 text-xs text-muted-foreground">mmHg; pulse in beats per minute.</p>
        </div>
      </section>

      <!-- Blood glucose -->
      <section class="space-y-3 rounded-lg border border-border bg-card p-4">
        <label class="flex items-center gap-2 text-sm font-semibold">
          <input v-model="glucoseOn" type="checkbox" class="h-4 w-4" :disabled="locked" />
          Blood glucose
        </label>
        <div v-if="glucoseOn" class="space-y-3">
          <div class="grid grid-cols-2 gap-3">
            <label class="text-sm">
              <span class="text-muted-foreground">Reading</span>
              <input v-model="glucose" v-bind="numberInput" inputmode="decimal" :disabled="locked" />
            </label>
            <label class="text-sm">
              <span class="text-muted-foreground">Unit</span>
              <select
                v-model="glucoseUnit"
                class="mt-1 w-full rounded-md border border-border bg-background px-3 py-2 text-sm"
                :disabled="locked"
              >
                <option value="mmol/L">mmol/L</option>
                <option value="mg/dL">mg/dL</option>
              </select>
            </label>
          </div>
          <fieldset>
            <legend class="text-sm text-muted-foreground">When taken</legend>
            <div class="mt-1 flex flex-wrap gap-2">
              <label
                v-for="c in GLUCOSE_CONTEXTS"
                :key="c.value"
                class="flex cursor-pointer items-center gap-2 rounded-md border px-3 py-1.5 text-sm"
                :class="glucoseContext === c.value ? 'border-foreground bg-foreground/5' : 'border-border'"
              >
                <input v-model="glucoseContext" type="radio" name="glucose-context" :value="c.value" class="h-4 w-4" :disabled="locked" />
                {{ c.label }}
              </label>
            </div>
          </fieldset>
        </div>
      </section>

      <!-- Note: kept with a customer's reading only -->
      <label v-if="customer" class="block text-sm">
        <span class="text-muted-foreground">Note (optional, kept with the reading, not printed)</span>
        <input
          v-model="note"
          maxlength="500"
          autocomplete="off"
          class="mt-1 w-full rounded-md border border-border bg-background px-3 py-2 text-sm"
          :disabled="locked"
        />
      </label>

      <p v-if="customerDeactivated" role="alert" class="rounded-md bg-destructive/10 px-3 py-2 text-sm text-destructive">
        {{ canManage
          ? "This customer is deactivated. Reactivate them or choose someone else."
          : "This customer is deactivated. Ask a manager to reactivate them, or choose someone else." }}
      </p>
      <p v-if="error" role="alert" class="rounded-md bg-destructive/10 px-3 py-2 text-sm text-destructive">{{ error }}</p>

      <div class="flex flex-wrap gap-2">
        <Button type="submit" :disabled="busy || printing || customerDeactivated">
          <Printer :size="16" class="mr-2" />
          {{ printing ? "Printing…" : busy ? "Saving…" : unsure ? "Try saving again" : customer ? "Save and print" : "Print" }}
        </Button>
        <Button type="button" variant="outline" :disabled="busy" @click="clearPressed">Clear</Button>
      </div>
    </form>
  </div>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { onBeforeRouteLeave } from "vue-router";
import { Printer } from "lucide-vue-next";
import { Button } from "@/components/ui/button";
import CustomerPicker from "../components/pos/CustomerPicker.vue";
import { csrfFetch } from "../lib/csrf.js";
import { newRetryKey, readError } from "../lib/api.js";
import { GLUCOSE_CONTEXTS, printHealthSlip, slipFromSaved } from "../lib/healthSlip.js";
import { readingsFrom } from "../lib/healthReadings.js";
import { usePermissions, useUser } from "../composables/usePermissions.js";
import { useConfirm } from "../composables/useConfirm.js";

const { hasPermission } = usePermissions();
const canRecord = hasPermission("healthcheck:record");
const canLookup = hasPermission("customer:view");
const canManage = hasPermission("customer:manage");
const picker = ref(null);
const printing = ref(false);
const user = useUser();
const { confirm } = useConfirm();

const numberInput = {
  type: "text",
  inputmode: "numeric",
  autocomplete: "off",
  maxlength: 6,
  class: "mt-1 w-full rounded-md border border-border bg-background px-3 py-2 text-lg tabular-nums",
};

const customer = ref(null);
const walkInName = ref("");
const bpOn = ref(true);
const systolic = ref("");
const diastolic = ref("");
const pulse = ref("");
const glucoseOn = ref(false);
const glucose = ref("");
const glucoseUnit = ref("mmol/L");
const glucoseContext = ref("");
const note = ref("");
const busy = ref(false);
const error = ref("");
const notice = ref(null);
// One key per form: pressing Save again after a lost reply can't save the reading twice. A new
// one is made only once this form is done (saved, printed or cleared).
let retryKey = newRetryKey();

// unsure holds a save that may have reached the server (its reply was lost): {values, consentGiven}.
// Until it is answered the form is locked and Save resends exactly that request under the same
// key, so the server either returns the first save or saves it once. Changing anything first
// would be refused as a different request (already_saved) and lose the correction.
const unsure = ref(null);
const locked = computed(() => busy.value || !!unsure.value);
const customerDeactivated = computed(() => customer.value?.active === false);
// The last slip printed (or tried), so it can be printed again: after a printer error, or a print
// dialog cancelled to change printers. A walk-in's slip is kept only in this page.
let lastSlip = null;

// What was shown about the last try no longer holds once the form changes.
watch([customer, walkInName, bpOn, systolic, diastolic, pulse, glucoseOn, glucose, glucoseUnit, glucoseContext, note], () => {
  error.value = "";
});
// A note belongs to the customer it was typed for.
watch(() => customer.value?.id, () => {
  note.value = "";
});

// sentence makes a server message ("this customer was not found") read as a sentence.
function sentence(text) {
  const t = String(text || "").trim();
  return t ? t.charAt(0).toUpperCase() + t.slice(1) + (/[.!?]$/.test(t) ? "" : ".") : "";
}

function clearForm() {
  customer.value = null;
  walkInName.value = "";
  bpOn.value = true;
  systolic.value = diastolic.value = pulse.value = "";
  glucoseOn.value = false;
  glucose.value = "";
  glucoseUnit.value = "mmol/L";
  glucoseContext.value = "";
  note.value = "";
  error.value = "";
  notice.value = null;
  unsure.value = null;
  retryKey = newRetryKey();
}

// readings returns what to send, or throws an Error with a message for the person.
function readings() {
  return readingsFrom({
    bpOn: bpOn.value, systolic: systolic.value, diastolic: diastolic.value, pulse: pulse.value,
    glucoseOn: glucoseOn.value, glucose: glucose.value, glucoseUnit: glucoseUnit.value,
    glucoseContext: glucoseContext.value,
  });
}

function slipFromForm(values, name, walkIn) {
  return {
    name,
    walkIn,
    takenAt: new Date(),
    takenBy: user.value?.username || "",
    systolic: values.systolic,
    diastolic: values.diastolic,
    pulse: values.pulse,
    glucose: values.glucose,
    glucoseUnit: values.glucose_unit,
    glucoseContext: values.glucose_context,
  };
}

const UNSURE_MESSAGE =
  "Could not reach the server, so the reading may or may not be saved. Press Try saving again; it won't be saved twice. Clear starts over.";

async function submit() {
  if (busy.value) return;
  notice.value = null;
  if (unsure.value) {
    await saveForCustomer(unsure.value.values, unsure.value.consentGiven);
    return;
  }
  let values;
  try {
    values = readings();
  } catch (e) {
    error.value = e.message;
    return;
  }
  if (picker.value?.isOpen) {
    error.value = customer.value
      ? "Finish changing the customer first: pick someone, or close the search."
      : "Pick the customer from the list, or close the search to print for a walk-in.";
    return;
  }
  if (customer.value) await saveForCustomer(values);
  else await printWalkIn(values);
}

// Clearing while a save may have landed would make a new key, and saving the same reading again
// would keep it twice: say so first.
async function clearPressed() {
  if (unsure.value) {
    const ok = await confirm({
      title: "Start over?",
      message:
        "This reading may already be saved (the reply was lost). Press Try saving again instead to find out without saving it twice.",
      tone: "warning",
      confirmLabel: "Start over",
      cancelLabel: "Keep it",
    });
    if (!ok) return;
  }
  clearForm();
}

// A walk-in's reading never leaves this page.
async function printWalkIn(values) {
  busy.value = printing.value = true;
  const slip = slipFromForm(values, walkInName.value, true);
  try {
    await printHealthSlip(slip);
    lastSlip = slip;
    clearForm();
    notice.value = { reprint: true, text: "Slip sent to print. Nothing was kept." };
  } catch {
    error.value = "The slip could not be printed. Check the printer and try again.";
  } finally {
    busy.value = printing.value = false;
  }
}

// post sends the reading and answers {reading} or {refusal: {message, code}}. When the server
// could not be reached it answers {unsure: true} and remembers the request in `unsure`.
// values carries customer_id, so a retry is for the customer it was first sent for.
async function post(values, consentGiven) {
  const lost = () => {
    unsure.value = { values, consentGiven };
    return { unsure: true };
  };
  let r;
  let reading;
  try {
    r = await csrfFetch("/health-checks/", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ ...values, consent_given: consentGiven, idempotency_key: retryKey }),
    });
    // A gateway error, or a 5xx that is not this app's own JSON answer, does not say the save
    // failed: a proxy may have given up while it committed. The app's own 500 is a rollback.
    if ([502, 503, 504].includes(r.status)) return lost();
    if (r.status >= 500 && !(r.headers.get("Content-Type") || "").includes("application/json")) {
      return lost();
    }
    // The body of a save can be lost too, after the save itself landed.
    if (r.ok) reading = await r.json();
  } catch {
    return lost();
  }
  // Signed out (401) or a stale form token (403) says nothing about an earlier lost save, so the
  // form stays locked for the same request; any other answer settles it.
  if (r.status !== 401 && r.status !== 403) unsure.value = null;
  error.value = "";
  if (r.ok) return { reading };
  return { refusal: await readError(r, "Could not save the health check.") };
}

async function saveForCustomer(values, consentGiven = false) {
  busy.value = true;
  const who = customer.value;
  // A retry passes the request it is repeating, whose customer and note win over the screen.
  values = { note: note.value, customer_id: who.id, ...values };
  let result;
  try {
    result = await post(values, consentGiven);
    if (result.refusal?.code === "health_consent_needed") {
      result = await askConsent(values, who);
      if (!result) return; // cancelled: the form stays as it was
    }
  } finally {
    busy.value = false;
  }
  await finish(result, values, who);
}

// askConsent asks, through the shared dialog, whether the pharmacy may keep the customer's
// readings. It answers like post, {printOnly: true} when they said no, or null when cancelled.
async function askConsent(values, who) {
  let result = null;
  const answer = await confirm({
    title: "Keep this customer's health readings?",
    message:
      "Ask the customer whether the pharmacy may keep their blood pressure and glucose readings, so the pharmacist can look back on them. If they say no, print the slip without keeping it.",
    detail: who.name,
    tone: "warning",
    confirmLabel: "They agree: save and print",
    busyLabel: "Saving…",
    alternate: { label: "Print only, don't save" },
    run: async () => {
      const r = await post(values, true);
      // Refusals the page handles itself close the dialog, and so does a lost reply: "Print only"
      // must not be offered once the save may have reached the server. Anything else is shown
      // inside the dialog.
      if (r.refusal && !["customer_inactive", "already_saved"].includes(r.refusal.code)) {
        throw new Error(sentence(r.refusal.message));
      }
      result = r;
    },
  });
  if (answer === "alternate") return { printOnly: true };
  return answer === true ? result : null;
}

async function finish(result, values, who) {
  if (result.unsure) {
    error.value = UNSURE_MESSAGE;
    return;
  }
  if (result.printOnly) {
    busy.value = printing.value = true;
    const slip = slipFromForm(values, who.name, false);
    try {
      await printHealthSlip(slip);
      lastSlip = slip;
      clearForm();
      notice.value = { reprint: true, text: `Slip sent to print for ${who.name}. The reading was not kept.` };
    } catch {
      error.value = "The slip could not be printed. Check the printer and try again.";
    } finally {
      busy.value = printing.value = false;
    }
    return;
  }
  const { reading, refusal } = result;
  if (refusal?.code === "customer_inactive") {
    customer.value = { ...who, active: false };
    return;
  }
  if (refusal?.code === "already_saved") {
    clearForm();
    notice.value = {
      error: true,
      text: "This form was already saved, so it was not saved again. Check with a manager before recording it once more.",
    };
    return;
  }
  if (refusal) {
    // A 401/403 leaves the lock on: say why the form is still locked, not only the refusal.
    error.value = unsure.value ? `${sentence(refusal.message)} ${UNSURE_MESSAGE}` : sentence(refusal.message);
    return;
  }
  // Saved: the form is done whatever the printer does, so it can't be saved twice.
  clearForm();
  lastSlip = slipFromSaved(reading);
  await printSlip(`Saved for ${reading.customer_name}`);
}

// printSlip prints lastSlip. It uses only `printing`, not `busy`: the form is already clear and
// may be in use for the next person, whose customer picker must not be closed by a reprint.
// `done` is what happened before printing ("Saved for Ada"), or empty for a reprint.
async function printSlip(done = "") {
  printing.value = true;
  try {
    await printHealthSlip(lastSlip);
    notice.value = { reprint: true, text: done ? `${done} and sent to print.` : "Slip sent to print again." };
  } catch {
    notice.value = {
      error: true,
      reprint: true,
      text: done
        ? `${done}, but the slip could not be printed. Check the printer.`
        : "The slip could not be printed. Check the printer.",
    };
  } finally {
    printing.value = false;
  }
}

function printAgain() {
  if (lastSlip) printSlip();
}

// While a save may have landed the page holds the only way to learn its outcome without saving it
// twice (the same request under the same key): leaving or reloading would lose it.
onBeforeRouteLeave(async () => {
  if (busy.value) {
    notice.value = { error: true, text: "Wait for the save to finish." };
    return false;
  }
  if (!unsure.value) return true;
  return await confirm({
    title: "Leave this page?",
    message:
      "This reading may already be saved (the reply was lost). Press Try saving again first to find out without saving it twice.",
    tone: "warning",
    confirmLabel: "Leave",
    cancelLabel: "Stay",
  });
});
function warnBeforeUnload(e) {
  if (unsure.value || busy.value) {
    e.preventDefault();
    e.returnValue = "";
  }
}
onMounted(() => window.addEventListener("beforeunload", warnBeforeUnload));
onBeforeUnmount(() => window.removeEventListener("beforeunload", warnBeforeUnload));
</script>
