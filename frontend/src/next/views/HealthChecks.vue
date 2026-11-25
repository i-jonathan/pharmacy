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
      <Button v-if="notice.reprint" variant="outline" size="sm" class="ml-2 h-7" :disabled="busy" @click="printAgain">Print again</Button>
    </p>

    <!-- Who. Outside the form: the picker has buttons and a form of its own (adding a customer),
       which must not submit, or nest inside, the reading's form. -->
    <section class="rounded-lg border border-border bg-card py-2">
      <h2 class="px-4 pt-1 text-xs font-semibold uppercase tracking-wider text-muted-foreground">Who</h2>
      <CustomerPicker v-model="customer" :disabled="locked" for-what="reading" />
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
        This customer is deactivated. Reactivate them or choose someone else.
      </p>
      <p v-if="error" role="alert" class="rounded-md bg-destructive/10 px-3 py-2 text-sm text-destructive">{{ error }}</p>

      <div class="flex flex-wrap gap-2">
        <Button type="submit" :disabled="busy || customerDeactivated">
          <Printer :size="16" class="mr-2" />
          {{ busy ? "Saving…" : unsure ? "Try saving again" : customer ? "Save and print" : "Print" }}
        </Button>
        <Button type="button" variant="outline" :disabled="busy" @click="clearForm()">Clear</Button>
      </div>
    </form>
  </div>
</template>

<script setup>
import { computed, ref, watch } from "vue";
import { Printer } from "lucide-vue-next";
import { Button } from "@/components/ui/button";
import CustomerPicker from "../components/pos/CustomerPicker.vue";
import { csrfFetch } from "../lib/csrf.js";
import { newRetryKey, readError } from "../lib/api.js";
import { GLUCOSE_CONTEXTS, printHealthSlip, slipFromSaved } from "../lib/healthSlip.js";
import { usePermissions, useUser } from "../composables/usePermissions.js";
import { useConfirm } from "../composables/useConfirm.js";

const { hasPermission } = usePermissions();
const canRecord = hasPermission("healthcheck:record");
const canLookup = hasPermission("customer:view");
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
// The last saved reading, kept so a slip that failed to print can be printed again.
let lastSaved = null;

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

// The same ranges and rules as the server, so a typo is caught before anything is sent. They are
// typo guards, not clinical limits; no message here says a number is good or bad.
const RANGES = {
  systolic: [50, 300],
  diastolic: [30, 200],
  pulse: [20, 250],
  "mmol/L": [1, 50],
  "mg/dL": [20, 900],
};

function whole(text) {
  const t = String(text).trim();
  return /^\d+$/.test(t) ? Number(t) : null;
}

function inRange(value, [min, max]) {
  return value !== null && value >= min && value <= max;
}

// readings returns what to send, or throws an Error with a message for the person.
function readings() {
  const out = {
    systolic: null, diastolic: null, pulse: null,
    glucose: null, glucose_unit: "", glucose_context: "",
  };
  if (!bpOn.value && !glucoseOn.value) throw new Error("Tick blood pressure, blood glucose, or both.");
  if (bpOn.value) {
    const sys = whole(systolic.value);
    const dia = whole(diastolic.value);
    if (!String(systolic.value).trim() || !String(diastolic.value).trim()) {
      throw new Error("Enter both blood pressure numbers.");
    }
    if (!inRange(sys, RANGES.systolic)) throw new Error("The top number must be between 50 and 300. Check it.");
    if (!inRange(dia, RANGES.diastolic)) throw new Error("The bottom number must be between 30 and 200. Check it.");
    if (sys <= dia) throw new Error("The top number must be higher than the bottom one.");
    out.systolic = sys;
    out.diastolic = dia;
    if (String(pulse.value).trim()) {
      const p = whole(pulse.value);
      if (!inRange(p, RANGES.pulse)) throw new Error("The pulse must be between 20 and 250. Check it.");
      out.pulse = p;
    }
  }
  if (glucoseOn.value) {
    const text = String(glucose.value).trim();
    if (!text) throw new Error("Enter the glucose reading.");
    if (!/^\d+(\.\d)?$/.test(text)) throw new Error("Enter the glucose as a number with at most one decimal place.");
    const value = Number(text);
    const range = RANGES[glucoseUnit.value];
    if (!inRange(value, range)) {
      throw new Error(`Glucose in ${glucoseUnit.value} must be between ${range[0]} and ${range[1]}. Check the reading and the unit.`);
    }
    if (!glucoseContext.value) throw new Error("Choose when the glucose was taken.");
    out.glucose = value;
    out.glucose_unit = glucoseUnit.value;
    out.glucose_context = glucoseContext.value;
  }
  return out;
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
  if (customer.value) await saveForCustomer(values);
  else await printWalkIn(values);
}

// A walk-in's reading never leaves this page.
async function printWalkIn(values) {
  busy.value = true;
  try {
    await printHealthSlip(slipFromForm(values, walkInName.value, true));
    clearForm();
    notice.value = { text: "Slip sent to print. Nothing was kept." };
  } catch {
    error.value = "The slip could not be printed. Check the printer and try again.";
  } finally {
    busy.value = false;
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
    // A 5xx (the server, or a proxy that gave up waiting) does not say the save failed: it may
    // have committed. Only a 4xx is a definite refusal.
    if (r.status >= 500) return lost();
    // The body of a save can be lost too, after the save itself landed.
    if (r.ok) reading = await r.json();
  } catch {
    return lost();
  }
  unsure.value = null;
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
    busy.value = true;
    try {
      await printHealthSlip(slipFromForm(values, who.name, false));
      clearForm();
      notice.value = { text: `Slip sent to print for ${who.name}. The reading was not kept.` };
    } catch {
      error.value = "The slip could not be printed. Check the printer and try again.";
    } finally {
      busy.value = false;
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
    error.value = sentence(refusal.message);
    return;
  }
  // Saved: the form is done whatever the printer does, so it can't be saved twice.
  clearForm();
  lastSaved = reading;
  await printSaved(reading);
}

async function printSaved(reading) {
  busy.value = true;
  try {
    await printHealthSlip(slipFromSaved(reading));
    notice.value = { text: `Saved for ${reading.customer_name} and sent to print.` };
  } catch {
    notice.value = {
      error: true,
      reprint: true,
      text: `Saved for ${reading.customer_name}, but the slip could not be printed. Check the printer.`,
    };
  } finally {
    busy.value = false;
  }
}

function printAgain() {
  if (lastSaved) printSaved(lastSaved);
}
</script>
