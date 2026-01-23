/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./backend/template/**/*.html",
    "./backend/template/static/js/*.js",
    "./frontend/src/**/*.{vue,js,ts}",
  ],
  darkMode: "class",
  theme: {
    extend: {
      colors: {
        primary: "#10B981",
      },
    },
  },
  plugins: [],
};
