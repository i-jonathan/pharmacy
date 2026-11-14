// showPhone writes a stored phone number the way people read it: +2348031234567 becomes
// 0803 123 4567. Anything else (a foreign number) is shown as stored.
export function showPhone(phone) {
  const m = /^\+234(\d{3})(\d{3})(\d{4})$/.exec(phone || "");
  if (m) return `0${m[1]} ${m[2]} ${m[3]}`;
  return phone || "";
}
