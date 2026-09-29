<template>
  <main class="flex min-h-[calc(100dvh-3.5rem)] items-center justify-center px-4 py-10 sm:px-6">
    <section
      aria-labelledby="not-found-title"
      class="w-full max-w-2xl overflow-hidden rounded-3xl border border-border bg-card shadow-[0_24px_80px_-40px_rgba(15,23,42,0.22)]"
    >
      <div class="p-6 sm:p-12">
        <div class="mb-8 flex items-center gap-4">
          <div
            class="flex h-14 w-14 shrink-0 items-center justify-center rounded-2xl bg-amber-50 text-amber-700 dark:bg-amber-950/60 dark:text-amber-300"
            aria-hidden="true"
          >
            <Search :size="27" :stroke-width="1.7" />
          </div>
          <div>
            <p class="text-xs font-bold uppercase tracking-[0.18em] text-primary">
              Page not found <span class="mx-1 text-muted-foreground/50">/</span> 404
            </p>
            <p class="mt-1 text-sm text-muted-foreground">We couldn’t find that destination.</p>
          </div>
        </div>

        <h2
          id="not-found-title"
          class="max-w-xl text-3xl font-semibold tracking-tight text-foreground sm:text-4xl"
        >
          This page isn’t here.
        </h2>
        <p class="mt-4 max-w-lg text-base leading-7 text-muted-foreground">
          The link may be outdated, or the page may have moved. Check the address or head back to your dashboard.
        </p>

        <div class="mt-9 flex flex-col gap-3 sm:flex-row">
          <router-link
            to="/"
            class="inline-flex min-h-12 items-center justify-center gap-2 rounded-xl bg-primary px-5 py-3 text-sm font-semibold text-primary-foreground shadow-sm transition hover:brightness-95 focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary focus-visible:outline-offset-3"
          >
            Go to dashboard
            <ArrowRight :size="16" :stroke-width="1.8" aria-hidden="true" />
          </router-link>
          <button
            type="button"
            class="inline-flex min-h-12 items-center justify-center gap-2 rounded-xl border border-border bg-background px-5 py-3 text-sm font-semibold text-foreground transition hover:bg-muted focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary focus-visible:outline-offset-3"
            @click="goBack"
          >
            <ArrowLeft :size="16" :stroke-width="1.8" aria-hidden="true" />
            Go back
          </button>
        </div>
      </div>

      <div class="border-t border-border bg-muted/30 px-6 py-4 text-xs text-muted-foreground sm:px-12">
        Primocrest Pharmacy
      </div>
    </section>
  </main>
</template>

<script setup>
import { ArrowLeft, ArrowRight, Search } from "lucide-vue-next";

function goBack() {
  if (window.history.length > 1) window.history.back();
  else window.location.assign("/app/");
}
</script>
