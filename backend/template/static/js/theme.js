const html = document.documentElement;
const themeToggles = Array.from(
  document.querySelectorAll("#theme-toggle, #theme-toggle-desktop"),
);

function setToggleIcon(button, isDark) {
  const icon = button.querySelector("span") || button;
  icon.textContent = isDark ? "☀️" : "🌙";
}

function applyTheme(isDark) {
  html.classList.toggle("dark", isDark);
  localStorage.theme = isDark ? "dark" : "light";
  themeToggles.forEach((button) => setToggleIcon(button, isDark));
}

applyTheme(localStorage.theme === "dark");

themeToggles.forEach((button) => {
  button.addEventListener("click", () => {
    applyTheme(!html.classList.contains("dark"));
  });
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
