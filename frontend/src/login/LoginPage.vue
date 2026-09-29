<script setup>
import { computed, onMounted, ref } from "vue";
import { Eye, EyeOff, Moon, Sun } from "lucide-vue-next";

const props = defineProps({
  csrfToken: { type: String, default: "" },
  hasLoginError: { type: Boolean, default: false },
});

const showPassword = ref(false);
const isDark = ref(false);
const isSubmitting = ref(false);
const hasLoginError = ref(props.hasLoginError);
const passwordType = computed(() => (showPassword.value ? "text" : "password"));

async function submitLogin(event) {
  if (isSubmitting.value) return;
  hasLoginError.value = false;
  isSubmitting.value = true;
  const formData = new FormData(event.currentTarget);

  try {
    const response = await fetch("/user/login", {
      method: "POST",
      credentials: "same-origin",
      headers: {
        "Accept": "application/json",
        "Content-Type": "application/json",
        "X-CSRF-Token": props.csrfToken,
      },
      body: JSON.stringify({
        username: formData.get("username"),
        password: formData.get("password"),
      }),
    });
    const data = await response.json();
    if (!response.ok || !data.redirect) throw new Error("Login failed");
    window.location.assign(data.redirect);
  } catch {
    hasLoginError.value = true;
    isSubmitting.value = false;
  }
}

function applyTheme(dark) {
  isDark.value = dark;
  document.documentElement.classList.toggle("dark", dark);
  try {
    localStorage.setItem("theme", dark ? "dark" : "light");
  } catch {
    // Theme preference is optional when storage is unavailable.
  }
}

onMounted(() => {
  let savedTheme = null;
  try {
    savedTheme = localStorage.getItem("theme");
  } catch {
    // Fall back to the system preference.
  }
  applyTheme(savedTheme ? savedTheme === "dark" : window.matchMedia("(prefers-color-scheme: dark)").matches);
});
</script>

<template>
  <main class="grid min-h-screen min-h-[100dvh] bg-neutral-50 text-neutral-900 dark:bg-neutral-950 dark:text-neutral-100 lg:grid-cols-[minmax(320px,0.9fr)_1.1fr]">
    <aside class="relative hidden flex-col justify-between overflow-hidden bg-emerald-950 px-12 py-10 text-white lg:flex xl:px-16 xl:py-12">
      <div aria-hidden="true" class="pointer-events-none absolute inset-0 opacity-[0.12]">
        <div class="absolute -right-28 top-24 h-96 w-96 rounded-full border border-white"></div>
        <div class="absolute -right-12 top-40 h-64 w-64 rounded-full border border-white"></div>
        <div class="absolute -bottom-44 -left-32 h-[28rem] w-[28rem] rounded-full border border-white"></div>
      </div>
      <div class="relative inline-flex w-fit items-center gap-3 text-white">
        <span class="grid h-10 w-10 place-items-center border border-white/35 text-sm font-semibold tracking-wide">P</span>
        <span><span class="block text-base font-semibold tracking-tight">Primocrest</span><span class="mt-0.5 block text-xs text-emerald-100/75">PHARMACY</span></span>
      </div>
      <section class="relative max-w-lg py-16">
        <p class="mb-5 text-xs font-semibold uppercase tracking-[0.18em] text-emerald-200">Pharmacy operations</p>
        <h1 class="max-w-md text-4xl font-semibold leading-[1.12] tracking-tight xl:text-5xl">Good care starts with a well-run pharmacy.</h1>
        <p class="mt-6 max-w-md text-base leading-7 text-emerald-50/75">A focused workspace for the people who keep your pharmacy moving, every day.</p>
      </section>
      <div class="relative flex items-center gap-3 border-t border-white/15 pt-5 text-xs text-emerald-100/70"><span aria-hidden="true" class="h-1.5 w-1.5 rounded-full bg-emerald-300"></span><span>Staff access · Inventory · Point of sale</span></div>
    </aside>

    <section class="flex min-h-screen min-h-[100dvh] flex-col bg-neutral-50 px-5 py-6 dark:bg-neutral-950 sm:px-8 sm:py-8 lg:px-12">
      <header class="flex items-center justify-between">
        <div class="inline-flex items-center gap-3 text-neutral-900 dark:text-neutral-100 lg:invisible">
          <span class="grid h-9 w-9 place-items-center border border-neutral-300 text-sm font-semibold dark:border-neutral-700">P</span>
          <span><span class="block text-sm font-semibold tracking-tight">Primocrest</span><span class="block text-[10px] tracking-wide text-neutral-500 dark:text-neutral-400">PHARMACY</span></span>
        </div>
        <button type="button" :aria-label="isDark ? 'Switch to light mode' : 'Switch to dark mode'" :aria-pressed="isDark" class="inline-flex min-h-10 items-center gap-2 rounded-sm border border-neutral-300 px-3 text-xs font-medium text-neutral-700 transition-colors hover:bg-neutral-100 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-emerald-800 dark:border-neutral-700 dark:text-neutral-200 dark:hover:bg-neutral-900" @click="applyTheme(!isDark)">
          <Sun v-if="isDark" :size="16" aria-hidden="true" /><Moon v-else :size="16" aria-hidden="true" />
          <span>{{ isDark ? 'Light mode' : 'Dark mode' }}</span>
        </button>
      </header>

      <div class="flex flex-1 items-center justify-center py-12 lg:py-8">
        <div class="w-full max-w-[420px]">
          <p class="text-sm font-medium text-emerald-800 dark:text-emerald-400">Staff sign-in</p>
          <h2 class="mt-3 text-3xl font-semibold tracking-tight">Welcome back</h2>
          <p class="mt-2 text-sm leading-6 text-neutral-600 dark:text-neutral-400">Sign in with your pharmacy account to continue.</p>

          <div v-if="hasLoginError" class="mt-6 border border-red-200 bg-red-50 px-4 py-3 text-sm leading-5 text-red-800 dark:border-red-900 dark:bg-red-950/40 dark:text-red-200" role="alert">
            We couldn’t sign you in. Check your username and password, then try again.
          </div>

          <form class="mt-8 space-y-5" @submit.prevent="submitLogin">
            <div>
              <label for="username" class="mb-2 block text-sm font-medium">Username</label>
              <input id="username" name="username" type="text" autocomplete="username" autocapitalize="none" spellcheck="false" required autofocus class="login-field min-h-12 w-full rounded-sm border border-neutral-300 bg-white px-3.5 text-sm text-neutral-900 placeholder:text-neutral-400 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-emerald-800 dark:border-neutral-700 dark:bg-neutral-900 dark:text-neutral-100 dark:placeholder:text-neutral-500" />
            </div>
            <div>
              <label for="password" class="mb-2 block text-sm font-medium">Password</label>
              <div class="relative">
                <input id="password" name="password" :type="passwordType" autocomplete="current-password" required class="login-field min-h-12 w-full rounded-sm border border-neutral-300 bg-white px-3.5 pr-12 text-sm text-neutral-900 placeholder:text-neutral-400 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-emerald-800 dark:border-neutral-700 dark:bg-neutral-900 dark:text-neutral-100 dark:placeholder:text-neutral-500" />
                <button type="button" :aria-label="showPassword ? 'Hide password' : 'Show password'" :aria-pressed="showPassword" class="absolute inset-y-0 right-0 inline-flex w-12 items-center justify-center text-neutral-500 hover:text-neutral-800 focus-visible:outline focus-visible:outline-2 focus-visible:outline-inset focus-visible:outline-emerald-800 dark:text-neutral-400 dark:hover:text-neutral-100" @click="showPassword = !showPassword">
                  <EyeOff v-if="showPassword" :size="18" aria-hidden="true" /><Eye v-else :size="18" aria-hidden="true" />
                </button>
              </div>
            </div>
            <button type="submit" :disabled="isSubmitting" class="min-h-12 w-full rounded-sm bg-emerald-800 px-4 text-sm font-semibold text-white transition-colors hover:bg-emerald-900 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-emerald-800 disabled:cursor-wait disabled:opacity-70 dark:bg-emerald-700 dark:hover:bg-emerald-600">{{ isSubmitting ? 'Signing in…' : 'Sign in' }}</button>
          </form>

          <p class="mt-6 text-xs leading-5 text-neutral-500 dark:text-neutral-400">Need an account or password reset? Contact your pharmacy administrator.</p>
        </div>
      </div>

      <footer class="text-center text-xs text-neutral-500 dark:text-neutral-500 lg:text-left">Primocrest Pharmacy · Authorized staff only</footer>
    </section>
  </main>
</template>
