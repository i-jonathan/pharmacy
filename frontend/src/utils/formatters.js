/**
 * Format an ISO date string to "DD MMM YYYY, HH:MM AM/PM"
 * Example: "24 Jan 2026, 08:58 PM"
 */
export function formatDate(isoString, options = {}) {
  if (!isoString) return "";
  const d = new Date(isoString);
  return d.toLocaleString("en-GB", {
    day: "numeric",
    month: "short",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
    hour12: true,
    ...options,
  });
}

/**
 * Format an ISO date string to "MMM YYYY"
 * Example: "Jan 2026"
 */
export function formatMonthYear(isoString) {
  if (!isoString) return "";
  const d = new Date(isoString);
  return d.toLocaleString("en-GB", { month: "short", year: "numeric" });
}

/**
 * Compute a human-friendly "time ago" string from a date
 * Example: "5m ago", "2h ago", "3d ago"
 */
export function timeAgo(isoString) {
  if (!isoString) return "Never";

  const now = new Date();
  const past = new Date(isoString);
  const diff = now - past; // difference in milliseconds

  const seconds = Math.floor(diff / 1000);
  if (seconds < 60) return `${seconds}s ago`;

  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `${minutes}m ago`;

  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}h ago`;

  const days = Math.floor(hours / 24);
  if (days < 30) return `${days}d ago`;

  // Otherwise show a date like "Jan 25"
  return updatedDate.toLocaleDateString(undefined, {
    month: "short",
    day: "numeric",
  });
}

/**
 * Formats a date to a date string like "2026-10-01"
 * @param {Date} date
 * @returns string
 */
export function formatToDateString(date) {
  const d = new Date(date);
  const year = d.getFullYear();
  const month = String(d.getMonth() + 1).padStart(2, "0");
  const day = String(d.getDate()).padStart(2, "0");

  return `${year}-${month}-${day}`;
}
