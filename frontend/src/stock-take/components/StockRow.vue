<template>
    <tr class="hover:bg-gray-50 dark:hover:bg-gray-700">
        <td
            class="px-4 py-3 font-medium sticky left-0 bg-white dark:bg-gray-800 z-10"
        >
            {{ item.name }}
            <div class="text-sm text-gray-600 dark:text-gray-400">
                {{ item.manufacturer }}
            </div>
            <div class="text-xs text-gray-500">
                Last edited by {{ item.lastEditedBy }} ·
                {{ item.lastEditedAgo }}
            </div>
        </td>

        <td v-if="showQuantityAndVariance" class="px-4 py-3 text-center">
            {{ item.stock }}
        </td>

        <td class="px-4 py-3 text-center">
            <input
                type="number"
                v-model.number="item.dispCount"
                @change="save"
                class="w-20 px-2 py-2 border rounded text-center dark:bg-gray-700 dark:border-gray-600"
            />
        </td>

        <td class="px-4 py-3 text-center">
            <input
                type="number"
                v-model.number="item.storeCount"
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
                <option value="">Select date</option>
                <option v-for="d in expiryOptions" :key="d" :value="d">
                    {{ d }}
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
export default {
    props: {
        item: Object,
        showQuantityAndVariance: Boolean,
    },
    data() {
        return {
            dispCount: this.item.dispCount,
            storeCount: this.item.storeCount,
            expiry: this.item.expiry || "",
            notes: this.item.notes || "",
            expiryOptions: this.item.expiryOptions || [],
        };
    },
    computed: {
        variance() {
            return (
                (this.item.dispCount ?? 0) +
                (this.item.storeCount ?? 0) -
                (this.item.stock ?? 0)
            );
        },
    },
    methods: {
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
