import { reactive } from "vue";

// One confirmation dialog for the whole app, replacing window.confirm and the per-page modals.
// <ConfirmDialog /> is mounted once in App.vue; any code can call `await confirm({ ... })`.
//
// The promise resolves true when the user confirms, "alternate" when they pick the optional third
// action, and false on cancel, Escape or a click on the backdrop.
//
// `alternate` is `{ label, run?, danger? }`; `danger` paints it red for a destructive choice.
//
// `run` (and `alternate.run`) let the dialog own the slow part: it shows a busy state while the
// function runs and closes when it returns. If it throws, the message is shown inside the dialog
// and the dialog stays open so the user can retry or cancel.
const state = reactive({
  open: false,
  title: "",
  message: "",
  detail: "",
  lines: [],
  tone: "warning",
  confirmLabel: "Confirm",
  busyLabel: "",
  cancelLabel: "Cancel",
  alternate: null,
  busy: false,
  error: "",
});
let resolver = null;
let runs = {};
// Each confirm() takes the next number. A `run` that was still going when a newer dialog replaced
// its own must not settle, or write its error or busy state into, the newer one.
let generation = 0;

function settle(answer) {
  state.open = false;
  const resolve = resolver;
  resolver = null;
  runs = {};
  if (resolve) resolve(answer);
}

async function choose(answer) {
  if (state.busy) return;
  const run = runs[answer === true ? "confirm" : "alternate"];
  if (!run) return settle(answer);
  const mine = generation;
  state.busy = true;
  state.error = "";
  try {
    await run();
    if (mine === generation) settle(answer);
  } catch (e) {
    if (mine === generation) state.error = e?.message || "Something went wrong. Please try again.";
  } finally {
    if (mine === generation) state.busy = false;
  }
}

export function useConfirm() {
  function confirm({
    title,
    message = "",
    detail = "",
    lines = [],
    tone = "warning",
    confirmLabel = "Confirm",
    busyLabel = "Working…",
    cancelLabel = "Cancel",
    alternate = null,
    run = null,
  }) {
    // A second request replaces the first, which counts as cancelled.
    if (resolver) settle(false);
    generation++;
    runs = { confirm: run, alternate: alternate?.run || null };
    Object.assign(state, {
      open: true,
      title,
      message,
      detail,
      lines,
      tone,
      confirmLabel,
      busyLabel,
      cancelLabel,
      alternate: alternate ? { label: alternate.label, danger: !!alternate.danger } : null,
      busy: false,
      error: "",
    });
    return new Promise((resolve) => {
      resolver = resolve;
    });
  }

  return {
    state,
    confirm,
    accept: () => choose(true),
    chooseAlternate: () => choose("alternate"),
    cancel: () => {
      if (!state.busy) settle(false);
    },
  };
}
