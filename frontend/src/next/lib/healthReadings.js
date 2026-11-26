// Checks and wording for a blood pressure / glucose reading, shared by the Health checks screen and
// the correction form on the customer page.
//
// The ranges are the server's (service/health_checks.go), so a typo is caught before anything is
// sent. They are typo guards, not clinical limits: no message here says a number is good or bad.

import { glucoseContextLabel } from "./healthSlip.js";

export const RANGES = {
  systolic: [50, 300],
  diastolic: [30, 200],
  pulse: [20, 250],
  "mmol/L": [1, 50],
  "mg/dL": [20, 900],
};

// whole reads a whole number; anything else (120.5, 12O) is "not a whole number", which is said
// as such rather than as a range.
function whole(text) {
  const t = String(text ?? "").trim();
  return /^\d+$/.test(t) ? Number(t) : null;
}

function inRange(value, [min, max]) {
  return value !== null && value >= min && value <= max;
}

function filled(v) {
  return String(v ?? "").trim() !== "";
}

// readingsFrom turns what was typed into the fields the server takes, or throws an Error with a
// message for the person. form = { bpOn, systolic, diastolic, pulse, glucoseOn, glucose,
// glucoseUnit, glucoseContext }. A measurement that is not ticked is sent empty (null / ""), which
// on a correction clears it.
export function readingsFrom(form) {
  const out = {
    systolic: null, diastolic: null, pulse: null,
    glucose: null, glucose_unit: "", glucose_context: "",
  };
  if (!form.bpOn && !form.glucoseOn) throw new Error("Tick blood pressure, blood glucose, or both.");
  if (form.bpOn) {
    if (!filled(form.systolic) || !filled(form.diastolic)) {
      throw new Error("Enter both blood pressure numbers.");
    }
    const sys = whole(form.systolic);
    const dia = whole(form.diastolic);
    if (sys === null || dia === null) throw new Error("Enter the blood pressure as whole numbers.");
    if (!inRange(sys, RANGES.systolic)) throw new Error("The top number must be between 50 and 300. Check it.");
    if (!inRange(dia, RANGES.diastolic)) throw new Error("The bottom number must be between 30 and 200. Check it.");
    if (sys <= dia) throw new Error("The top number must be higher than the bottom one.");
    out.systolic = sys;
    out.diastolic = dia;
    if (filled(form.pulse)) {
      const p = whole(form.pulse);
      if (p === null) throw new Error("Enter the pulse as a whole number.");
      if (!inRange(p, RANGES.pulse)) throw new Error("The pulse must be between 20 and 250. Check it.");
      out.pulse = p;
    }
  }
  if (form.glucoseOn) {
    const text = String(form.glucose ?? "").trim();
    if (!text) throw new Error("Enter the glucose reading.");
    if (!/^\d+(\.\d)?$/.test(text)) throw new Error("Enter the glucose as a number with at most one decimal place.");
    const value = Number(text);
    const range = RANGES[form.glucoseUnit];
    if (!range) throw new Error("Choose the glucose unit.");
    if (!inRange(value, range)) {
      throw new Error(`Glucose in ${form.glucoseUnit} must be between ${range[0]} and ${range[1]}. Check the reading and the unit.`);
    }
    if (!form.glucoseContext) throw new Error("Choose when the glucose was taken.");
    out.glucose = value;
    out.glucose_unit = form.glucoseUnit;
    out.glucose_context = form.glucoseContext;
  }
  return out;
}

// glucoseText shows a glucose value the way the slip does: mmol/L to one decimal place.
export function glucoseText(value, unit) {
  if (value === null || value === undefined || value === "") return "";
  return unit === "mmol/L" ? Number(value).toFixed(1) : String(value);
}

// describeGlucose reads "5.6 mmol/L (fasting)".
export function describeGlucose(r) {
  if (r.glucose === null || r.glucose === undefined) return "";
  const context = glucoseContextLabel(r.glucose_context);
  return `${glucoseText(r.glucose, r.glucose_unit)} ${r.glucose_unit || ""}${context ? ` (${context})` : ""}`.trim();
}
