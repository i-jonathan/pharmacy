<template>
  <Card
    class="relative overflow-hidden transition-colors hover:bg-muted/20"
    :class="{ 'cursor-pointer focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring': clickable }"
    :role="clickable ? 'button' : undefined"
    :tabindex="clickable ? 0 : undefined"
    @click="clickable && $emit('click')"
    @keydown.enter.prevent="clickable && $emit('click')"
    @keydown.space.prevent="clickable && $emit('click')"
  >
    <CardHeader class="pb-2">
      <div class="flex items-center justify-between">
        <span class="text-xs font-medium tracking-wide text-muted-foreground">
          {{ title }}
        </span>
        <span class="grid h-9 w-9 place-items-center rounded-md bg-primary/8">
          <component :is="icon" :class="iconColorClass" :size="18" :stroke-width="1.7" />
        </span>
      </div>
    </CardHeader>
    <CardContent>
      <div class="mb-1 text-2xl font-semibold tabular-nums tracking-tight">
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
  blue: { color: "text-primary" },
  indigo: { color: "text-primary" },
  emerald: { color: "text-primary" },
  amber: { color: "text-amber-700 dark:text-amber-300" },
  rose: { color: "text-destructive" },
};

const iconColorClass = computed(() => accentMap[props.accent]?.color ?? accentMap.indigo.color);

</script>
