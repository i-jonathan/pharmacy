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
                v-if="!isCompleted"
                type="number"
                min="0"
                v-model.number="item.dispensary_count"
                @focus="item.isEditing = true"
                @blur="item.isEditing = false"
                @change="save"
                @input="
                    item.dispensary_count = Math.max(
                        0,
                        $event.target.valueAsNumber || 0,
                    )
                "
                class="w-20 px-2 py-2 border rounded text-center dark:bg-gray-700 dark:border-gray-600"
            />

            <span v-else>
                {{ item.dispensary_count ?? 0 }}
            </span>
        </td>

        <td class="px-4 py-3 text-center">
            <input
                v-if="!isCompleted"
                type="number"
                min="0"
                v-model.number="item.store_count"
                @focus="item.isEditing = true"
                @blur="item.isEditing = false"
                @change="save"
                @input="
                    item.store_count = Math.max(
                        0,
                        $event.target.valueAsNumber || 0,
                    )
                "
                :disabled="!dispensaryEntered"
                :title="
                    !dispensaryEntered ? 'Enter dispensary count first' : ''
                "
                class="w-20 px-2 py-2 border rounded text-center dark:bg-gray-700 dark:border-gray-600 disabled:opacity-50 disabled:cursor-not-allowed"
            />

            <span v-else>
                {{ item.store_count ?? 0 }}
            </span>
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
                v-if="!isCompleted"
                v-model="expiry"
                @focus="item.isEditing = true"
                @blur="item.isEditing = false"
                @change="save"
                :disabled="!dispensaryEntered"
                :title="
                    !dispensaryEntered ? 'Enter dispensary count first' : ''
                "
                class="w-40 px-2 py-2 border rounded text-center disabled:opacity-50 disabled:cursor-not-allowed dark:bg-gray-700 dark:border-gray-600 dark:text-white"
            >
                <option value="" disabled>Select date</option>
                <option v-for="d in expiryOptions" :key="d" :value="d">
                    {{ formatMonthYear(d) }}
                </option>
            </select>

            <span v-else>
                {{ formatMonthYear(expiry) }}
            </span>
        </td>

        <td class="px-4 py-3 text-center">
            <input
                v-if="!isCompleted"
                type="text"
                v-model="notes"
                @focus="item.isEditing = true"
                @blur="item.isEditing = false"
                @change="save"
                placeholder="Optional"
                :disabled="!dispensaryEntered"
                :title="
                    !dispensaryEntered ? 'Enter dispensary count first' : ''
                "
                class="w-full px-2 py-1 border rounded dark:bg-gray-700 dark:border-gray-600 disabled:opacity-50 disabled:cursor-not-allowed"
            />

            <span v-else>
                {{ notes }}
            </span>
        </td>
    </tr>
</template>

<script>
import { formatMonthYear, timeAgo } from "@/utils/formatters";

export default {
    props: {
        item: Object,
        showQuantityAndVariance: Boolean,
        isCompleted: Boolean,
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
        dispensaryEntered() {
            return (
                this.item.dispensary_count !== null &&
                this.item.dispensary_count !== undefined
            );
        },
    },
    methods: {
        formatMonthYear,
        timeAgo,
        save() {
            const updated = {
                ...this.item,
                expiry: this.expiry,
                notes: this.notes,
            };
            this.$emit("update", updated);
        },
    },
};
</script>
