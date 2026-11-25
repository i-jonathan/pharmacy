// The health check slip, laid out for the same 80mm roll as the receipt and printed the same way.
// The Health checks screen prints it after a reading; the customer page reprints it as a copy.
//
// slip = {
//   name,                          the customer's name, or what a walk-in asked to have printed
//   walkIn,                        true for a walk-in (nothing was saved)
//   takenAt (Date or ISO text), takenBy,
//   systolic?, diastolic?, pulse?, glucose?, glucoseUnit?, glucoseContext?,
//   copy?: { at } | true           marks the slip as a reprint
// }
//
// Nothing here judges a number: no "normal", "high" or colours, on purpose.

import { escapeHtml, printReceiptHtml } from "./receipt.js";

const STORE_NAME = "Primocrest Pharmacy";

// How the glucose context reads on the slip and the screen.
export const GLUCOSE_CONTEXTS = [
  { value: "fasting", label: "Fasting" },
  { value: "after_meal", label: "After a meal" },
  { value: "random", label: "Random" },
];

export function glucoseContextLabel(value) {
  return GLUCOSE_CONTEXTS.find((c) => c.value === value)?.label.toLowerCase() || "";
}

function when(value) {
  const d = value instanceof Date ? value : new Date(value);
  return Number.isNaN(d.getTime())
    ? ""
    : d.toLocaleString("en-GB", { dateStyle: "medium", timeStyle: "short" });
}

function present(v) {
  return v !== null && v !== undefined && v !== "";
}

function slipName(s) {
  const name = String(s.name || "").trim();
  if (!s.walkIn) return name;
  return name ? `Walk-in: ${name}` : "Walk-in";
}

export function buildHealthSlipHtml(s) {
  const bp = present(s.systolic)
    ? `<div class="bold">Blood pressure</div>
       <div class="indent">${escapeHtml(s.systolic)} / ${escapeHtml(s.diastolic)} mmHg</div>
       ${present(s.pulse) ? `<div class="indent">Pulse ${escapeHtml(s.pulse)} bpm</div>` : ""}`
    : "";
  const context = glucoseContextLabel(s.glucoseContext);
  // mmol/L is read to one decimal place, so 5.0 prints as 5.0, not 5.
  const glucoseText =
    present(s.glucose) && s.glucoseUnit === "mmol/L" ? Number(s.glucose).toFixed(1) : s.glucose;
  const glucose = present(s.glucose)
    ? `<div class="bold">Blood glucose</div>
       <div class="indent">${escapeHtml(glucoseText)} ${escapeHtml(s.glucoseUnit)}${
         context ? ` (${escapeHtml(context)})` : ""
       }</div>`
    : "";
  const copy = s.copy
    ? `<div class="center bold">*** COPY${
        s.copy.at ? " " + escapeHtml(when(s.copy.at)) : ""
      } ***</div>`
    : "";

  return `<!doctype html>
<html>
<head>
<meta charset="utf-8"/>
<title>Health check</title>
<style>
  @page { size: 80mm auto; margin: 0; }
  * { box-sizing: border-box; }
  html, body { margin: 0; padding: 0; background: #fff; }
  body { width: 72mm; margin: 0 auto; padding: 4mm 2mm; font-family: monospace; font-size: 12px; line-height: 1.45; color: #000; }
  .center { text-align: center; }
  .bold { font-weight: bold; }
  .big { font-size: 14px; }
  .small { font-size: 11px; }
  .indent { padding-left: 4mm; font-size: 14px; }
  .name { overflow-wrap: anywhere; }
  .block + .block { margin-top: 4px; }
  hr { border: 0; border-top: 1px dashed #000; margin: 6px 0; }
</style>
</head>
<body>
  <div class="center bold big">${STORE_NAME}</div>
  <hr/>
  <div class="bold">Health check</div>
  ${copy}
  <div class="name">Name: ${escapeHtml(slipName(s))}</div>
  <div>${escapeHtml(when(s.takenAt))}</div>
  <hr/>
  ${bp ? `<div class="block">${bp}</div>` : ""}
  ${glucose ? `<div class="block">${glucose}</div>` : ""}
  <hr/>
  <div class="small">Taken by: ${escapeHtml(s.takenBy)}</div>
  <div class="small">For your own records.</div>
</body>
</html>`;
}

// printHealthSlip prints the slip; it rejects when the browser could not print it.
export function printHealthSlip(slip) {
  return printReceiptHtml(buildHealthSlipHtml(slip));
}

// slipFromSaved turns a saved reading from the server into a slip.
export function slipFromSaved(reading, copy) {
  return {
    name: reading.customer_name,
    walkIn: false,
    takenAt: reading.taken_at,
    takenBy: reading.taken_by,
    systolic: reading.systolic,
    diastolic: reading.diastolic,
    pulse: reading.pulse,
    glucose: reading.glucose,
    glucoseUnit: reading.glucose_unit,
    glucoseContext: reading.glucose_context,
    copy,
  };
}
