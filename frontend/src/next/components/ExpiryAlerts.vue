<template>
  <Card>
    <CardHeader>
      <div class="flex items-center justify-between">
        <CardTitle class="text-sm font-semibold uppercase tracking-wider">
          Expiry Alerts
        </CardTitle>
        <Badge variant="warning" class="text-xs">
          {{ totalItems }} items expiring soon
        </Badge>
      </div>
    </CardHeader>
    <CardContent>
      <div v-if="items.length === 0" class="text-center py-8 text-muted-foreground text-sm">
        No items expiring within 90 days
      </div>

      <div v-else class="space-y-3">
        <div
          v-for="item in items"
          :key="item.category"
          class="flex items-center justify-between p-4 rounded-xl bg-amber-50/60 dark:bg-amber-950/20 border border-amber-100 dark:border-amber-900/30"
        >
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 rounded-xl bg-amber-100 dark:bg-amber-900/30 flex items-center justify-center">
              <AlertTriangle :size="18" class="text-amber-600 dark:text-amber-400" />
            </div>
            <div>
              <div class="text-sm font-semibold">
                {{ item.category }}
              </div>
              <div class="text-xs text-muted-foreground">
                {{ item.count }} {{ item.count === 1 ? "item" : "items" }} expiring soon
              </div>
            </div>
          </div>
          <div class="text-right">
            <div class="text-sm font-bold text-amber-700 dark:text-amber-300">
              {{ formatNaira(item.total_cost_kobo) }}
            </div>
            <div class="text-xs text-muted-foreground">at risk</div>
          </div>
        </div>
      </div>
    </CardContent>
  </Card>
</template>

<script setup>
import { computed } from "vue";
import { AlertTriangle } from "lucide-vue-next";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";

const props = defineProps({
  data: { type: Array, default: () => [] },
});

const items = computed(() => props.data);
const totalItems = computed(() => props.data.reduce((sum, item) => sum + item.count, 0));

function formatNaira(kobo) {
  const naira = kobo / 100;
  return `₦${naira.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
}
</script>
