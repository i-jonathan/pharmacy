<template>
  <Card
    class="relative overflow-hidden transition-shadow hover:shadow-md"
    :class="{ 'cursor-pointer': clickable }"
    @click="$emit('click')"
  >
    <div class="absolute top-0 right-0 w-24 h-24 -mr-6 -mt-6 rounded-full opacity-10" :class="accentBgClass" />
    <CardHeader class="pb-2">
      <div class="flex items-center justify-between">
        <span class="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
          {{ title }}
        </span>
        <component :is="icon" :class="iconColorClass" :size="20" :stroke-width="1.5" />
      </div>
    </CardHeader>
    <CardContent>
      <div class="text-2xl font-bold mb-1">
        {{ formattedValue }}
      </div>
      <div v-if="trend !== null && trend !== undefined" class="flex items-center gap-1 text-sm">
        <component
          :is="trend >= 0 ? TrendingUp : TrendingDown"
          :size="14"
          :stroke-width="1.5"
          :class="trend >= 0 ? 'text-emerald-500' : 'text-destructive'"
        />
        <span :class="trend >= 0 ? 'text-emerald-600' : 'text-destructive'" class="font-medium">
          {{ Math.abs(trend).toFixed(1) }}%
        </span>
        <span class="text-muted-foreground text-xs">vs yesterday</span>
      </div>
      <div v-if="subtitle" class="text-xs text-muted-foreground mt-1">
        {{ subtitle }}
      </div>
    </CardContent>
  </Card>
</template>

<script setup>
import { computed } from "vue";
import { TrendingUp, TrendingDown } from "lucide-vue-next";
import { Card, CardContent, CardHeader } from "@/components/ui/card";

const props = defineProps({
  title: { type: String, required: true },
  value: { type: [Number, String], default: 0 },
  formattedValue: { type: String, default: "" },
  subtitle: { type: String, default: "" },
  trend: { type: Number, default: null },
  icon: { type: Object, required: true },
  accent: { type: String, default: "blue" },
  clickable: { type: Boolean, default: false },
});

defineEmits(["click"]);

const accentMap = {
  blue: { color: "text-blue-600 dark:text-blue-400", bg: "bg-blue-500" },
  indigo: { color: "text-indigo-600 dark:text-indigo-400", bg: "bg-indigo-500" },
  emerald: { color: "text-emerald-600 dark:text-emerald-400", bg: "bg-emerald-500" },
  amber: { color: "text-amber-600 dark:text-amber-400", bg: "bg-amber-500" },
  rose: { color: "text-rose-600 dark:text-rose-400", bg: "bg-rose-500" },
};

const iconColorClass = computed(() => accentMap[props.accent]?.color ?? accentMap.indigo.color);
const accentBgClass = computed(() => accentMap[props.accent]?.bg ?? accentMap.indigo.bg);
</script>
