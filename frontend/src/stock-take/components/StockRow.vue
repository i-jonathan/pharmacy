<template>
    <tr class="hover:bg-gray-50 dark:hover:bg-gray-700">
        <td
            class="px-4 py-3 font-medium sticky left-0 bg-white dark:bg-gray-800 z-10"
        >
            {{ item.product_name }}
            <div class="text-sm text-gray-600 dark:text-gray-400">
                {{ item.manufacturer }}
            </div>
            <div class="text-xs text-gray-500">
                Last edited by {{ item.last_updated_by }} ·
                {{ timeAgo(item.last_updated_at) }}
            </div>
        </td>

        <td v-if="showQuantityAndVariance" class="px-4 py-3 text-center">
            {{ item.snapshot_quantity }}
        </td>

        <td class="px-4 py-3 text-center">
            <input
                type="number"
                v-model.number="item.dispensary_count"
                @change="save"
                class="w-20 px-2 py-2 border rounded text-center dark:bg-gray-700 dark:border-gray-600"
            />
        </td>

        <td class="px-4 py-3 text-center">
            <input
                type="number"
                v-model.number="item.store_count"
                @change="save"
                class="w-20 px-2 py-2 border rounded text-center dark:bg-gray-700 dark:border-gray-600"
            />
        </td>

        <td
            v-if="showQuantityAndVariance"
            class="px-4 py-3 text-center font-semibold"
            :class="{
                'text-red-600': variance < 0,
                'text-green-600': variance > 0,
            }"
        >
            {{ variance }}
        </td>

        <td class="px-4 py-3 text-center">
            <select
                v-model="expiry"
                @change="save"
                class="w-40 px-2 py-2 border rounded text-center dark:bg-gray-700 dark:border-gray-600 dark:text-white"
            >
                <option value="" disabled>Select date</option>
                <option v-for="d in expiryOptions" :key="d" :value="d">
                    {{ formatMonthYear(d) }}
                </option>
            </select>
        </td>

        <td class="px-4 py-3">
            <input
                type="text"
                v-model="notes"
                @change="save"
                placeholder="Optional"
                class="w-full px-2 py-1 border rounded dark:bg-gray-700 dark:border-gray-600"
            />
        </td>
    </tr>
</template>

<script>
import { formatMonthYear, timeAgo } from "@/utils/formatters";

export default {
    props: {
        item: Object,
        showQuantityAndVariance: Boolean,
    },
    data() {
        return {
            dispCount: this.item.dispensary_count,
            storeCount: this.item.store_count,
            expiry: this.item.earliest_expiry || "",
            notes: this.item.notes || "",
            expiryOptions: this.item.expiry_options || [],
        };
    },
    computed: {
        variance() {
            return (
                (this.item.dispensary_count ?? 0) +
                (this.item.store_count ?? 0) -
                (this.item.snapshot_quantity ?? 0)
            );
        },
    },
    methods: {
        formatMonthYear,
        timeAgo,
        save() {
            const updated = {
                ...this.item,
                dispCount: this.dispCount,
                storeCount: this.storeCount,
                expiry: this.expiry,
                notes: this.notes,
            };
            this.$emit("update", updated);
        },
    },
};
</script>
