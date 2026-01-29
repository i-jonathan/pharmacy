<template>
    <div class="overflow-y-auto max-h-[900px] rounded-lg shadow">
        <table
            class="w-full min-w-[900px] bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700 table-fixed"
        >
            <thead class="bg-emerald-600 dark:bg-emerald-800">
                <tr>
                    <th
                        class="px-4 py-3 text-left text-white sticky top-0 left-0 bg-emerald-600 z-20 max-w-[350px]"
                    >
                        Item
                    </th>
                    <th
                        v-if="showQuantityAndVariance"
                        class="px-4 py-3 text-white sticky top-0 bg-emerald-600 dark:bg-emerald-800 z-10"
                    >
                        Quantity On Hand
                    </th>
                    <th
                        class="px-4 py-3 text-white sticky top-0 bg-emerald-600 dark:bg-emerald-800 z-10"
                    >
                        Disp. (Counted)
                    </th>
                    <th
                        class="px-4 py-3 text-white sticky top-0 bg-emerald-600 dark:bg-emerald-800 z-10"
                    >
                        Store (Counted)
                    </th>
                    <th
                        v-if="showQuantityAndVariance"
                        class="px-4 py-3 text-white sticky top-0 bg-emerald-600 dark:bg-emerald-800 z-10"
                    >
                        Variance
                    </th>
                    <th
                        class="px-4 py-3 text-white sticky top-0 bg-emerald-600 dark:bg-emerald-800 z-10"
                    >
                        Earliest Expiry
                    </th>
                    <th
                        class="px-4 py-3 text-white sticky top-0 bg-emerald-600 dark:bg-emerald-800 z-10 max-w-[150px]"
                    >
                        Notes
                    </th>
                </tr>
            </thead>
            <tbody>
                <StockRow
                    v-for="item in items"
                    :key="item.id"
                    :item="item"
                    :show-quantity-and-variance="showQuantityAndVariance"
                    @update="(i) => $emit('update-item', i)"
                />
            </tbody>
        </table>
    </div>
</template>

<script>
import StockRow from "./StockRow.vue";

export default {
    props: {
        items: Array,
        showQuantityAndVariance: Boolean,
    },
    components: { StockRow },
};
</script>
