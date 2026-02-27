const html = document.documentElement;
const themeToggle = document.getElementById("theme-toggle");

if (localStorage.theme === "dark") {
  html.classList.add("dark");
  themeToggle.textContent = "☀️";
}

themeToggle.addEventListener("click", () => {
  const isDark = html.classList.toggle("dark");
  localStorage.theme = isDark ? "dark" : "light";
  themeToggle.textContent = isDark ? "☀️" : "🌙";
});

document.getElementById("user-menu-button").addEventListener("click", () => {
  const dropdown = document.getElementById("user-menu-dropdown");
  dropdown.classList.toggle("hidden");
});

// Optional: close dropdown if clicked outside
document.addEventListener("click", (e) => {
  const menuButton = document.getElementById("user-menu-button");
  const dropdown = document.getElementById("user-menu-dropdown");
  if (!menuButton.contains(e.target) && !dropdown.contains(e.target)) {
    dropdown.classList.add("hidden");
  }
});
