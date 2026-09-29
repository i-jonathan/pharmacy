const html = document.documentElement;
const themeToggles = Array.from(
  document.querySelectorAll("#theme-toggle, #theme-toggle-desktop"),
);

function getPreferredTheme() {
  if (localStorage.theme) return localStorage.theme;
  if (window.matchMedia("(prefers-color-scheme: dark)").matches) return "dark";
  return "light";
}

function setToggleIcon(button, isDark) {
  const icon = button.querySelector("span") || button;
  icon.textContent = isDark ? "☀️" : "🌙";
}

function applyTheme(isDark) {
  html.classList.toggle("dark", isDark);
  localStorage.theme = isDark ? "dark" : "light";
  themeToggles.forEach((button) => setToggleIcon(button, isDark));
}

// Apply the correct theme on load — check localStorage + system preference
applyTheme(getPreferredTheme() === "dark");

themeToggles.forEach((button) => {
  button.addEventListener("click", () => {
    applyTheme(!html.classList.contains("dark"));
  });
});

// Listen for OS-level theme changes when no explicit preference is saved
window.matchMedia("(prefers-color-scheme: dark)").addEventListener("change", (e) => {
  if (!("theme" in localStorage)) {
    applyTheme(e.matches);
  }
});

const menuButton = document.getElementById("user-menu-button");
const dropdown = document.getElementById("user-menu-dropdown");

if (menuButton && dropdown) {
  menuButton.addEventListener("click", () => {
    dropdown.classList.toggle("hidden");
  });

  document.addEventListener("click", (e) => {
    if (!menuButton.contains(e.target) && !dropdown.contains(e.target)) {
      dropdown.classList.add("hidden");
    }
  });
}