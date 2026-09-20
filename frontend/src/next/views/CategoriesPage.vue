<template>
  <div class="p-6">
    <div class="mb-6">
      <h1 class="text-2xl font-bold text-foreground">Categories</h1>
      <p class="text-sm text-muted-foreground mt-1">Manage product categories</p>
    </div>
    <CategoriesManager />
  </div>
</template>

<script setup>
import CategoriesManager from "../../admin/views/CategoriesManager.vue";
</script>