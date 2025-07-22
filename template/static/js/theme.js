const html = document.documentElement;
const themeToggle = document.getElementById("theme-toggle");

// Init dark mode
if (localStorage.theme === "dark") {
  html.classList.add("dark");
  themeToggle.textContent = "☀️";
}

themeToggle.addEventListener("click", () => {
  const isDark = html.classList.toggle("dark");
  localStorage.theme = isDark ? "dark" : "light";
  themeToggle.textContent = isDark ? "☀️" : "🌙";
});

