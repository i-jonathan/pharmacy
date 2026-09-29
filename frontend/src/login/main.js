import { createApp } from "vue";
import LoginPage from "./LoginPage.vue";

const root = document.getElementById("login-app");

if (root) {
  createApp(LoginPage, {
    csrfToken: root.dataset.csrfToken || "",
    hasLoginError: root.dataset.loginError === "true",
  }).mount(root);
}
