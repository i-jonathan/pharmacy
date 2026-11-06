import { ref, computed, unref } from "vue";
import { usePermissions, useUser } from "./usePermissions.js";
import { releases, visibleSlides } from "../lib/whatsNew.js";

// The newest release this person has opened, remembered in this browser under their user id, so
// people sharing a till each get the dot. A blocked or full store only means the dot stays;
// nothing else depends on it.
const SEEN_KEY = "whatsNewSeen";

function readSeen(key) {
  try {
    return localStorage.getItem(key);
  } catch {
    return null;
  }
}

const open = ref(false);
let seen = null; // created on first use; signing in as someone else reloads the page

export function useWhatsNew() {
  const { hasPermission } = usePermissions();
  const key = `${SEEN_KEY}:${unref(useUser())?.id ?? 0}`;
  if (!seen) seen = ref(readSeen(key));
  const release = releases[0];
  const slides = computed(() => visibleSlides(release, hasPermission));
  // Nothing to show (no slides for this person) means no menu item and no dot.
  const available = computed(() => slides.value.length > 0);
  const unseen = computed(() => available.value && seen.value !== release.id);

  function show() {
    open.value = true;
    seen.value = release.id;
    try {
      localStorage.setItem(key, release.id);
    } catch {
      // Remembering is a convenience only.
    }
  }

  function close() {
    open.value = false;
  }

  return { open, release, slides, available, unseen, show, close };
}
