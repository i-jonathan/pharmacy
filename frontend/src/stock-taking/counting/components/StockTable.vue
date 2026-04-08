<template>
    <div class="relative">
        <!-- Sticky Category Header Overlay -->
        <div
            v-if="stickyCategoryHeader"
            class="sticky top-[60px] z-30 bg-emerald-50 dark:bg-emerald-900/20 pointer-events-none"
        >
            <div class="px-4 py-3 font-semibold text-emerald-800 dark:text-emerald-200">
                <div class="flex justify-between items-center">
                    <span>{{ stickyCategoryHeader.category }}</span>
                    <div class="flex gap-4 text-sm">
                        <span>{{ stickyCategoryHeader.count }} items</span>
                        <span
                            v-if="showQuantityAndVariance"
                            :class="{
                                'text-green-600 dark:text-green-400': stickyCategoryHeader.totalVariance > 0,
                                'text-red-600 dark:text-red-400': stickyCategoryHeader.totalVariance < 0,
                            }"
                        >
                            Variance: {{ stickyCategoryHeader.totalVariance }}
                        </span>
                    </div>
                </div>
            </div>
        </div>

        <div 
            class="overflow-y-auto max-h-[900px] shadow"
            :class="{ 'rounded-lg': !stickyCategoryHeader, 'rounded-t-none': stickyCategoryHeader }"
            @scroll="handleScroll" 
            ref="tableContainer"
        >
            <table
                class="w-full min-w-[900px] bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700"
                :class="{ 'rounded-t-lg': !stickyCategoryHeader, 'rounded-t-none': stickyCategoryHeader }"
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
                <template v-for="(item, index) in items" :key="item.id || `category-${index}`">
                    <!-- Category Header Row -->
                    <tr 
                        v-if="item.isCategoryHeader" 
                        class="bg-emerald-50 dark:bg-emerald-900/20"
                        :data-category-header="true"
                        :data-category="item.category"
                        :data-count="item.count"
                        :data-total-variance="item.totalVariance"
                    >
                        <td 
                            :colspan="showQuantityAndVariance ? 7 : 5"
                            class="px-4 py-3 font-semibold text-emerald-800 dark:text-emerald-200"
                        >
                            <div class="flex justify-between items-center">
                                <span>{{ item.category }}</span>
                                <div class="flex gap-4 text-sm">
                                    <span>{{ item.count }} items</span>
                                    <span
                                        v-if="showQuantityAndVariance"
                                        :class="{
                                            'text-green-600 dark:text-green-400': item.totalVariance > 0,
                                            'text-red-600 dark:text-red-400': item.totalVariance < 0,
                                        }"
                                    >
                                        Variance: {{ item.totalVariance }}
                                    </span>
                                </div>
                            </div>
                        </td>
                    </tr>
                    
                    <!-- Regular Item Row -->
                    <StockRow
                        v-else
                        :item="item"
                        :show-quantity-and-variance="showQuantityAndVariance"
                        :is-completed="isCompleted"
                        @update="(i) => $emit('update-item', i)"
                    />
                </template>
            </tbody>
        </table>
    </div>
    </div>
</template>

<script>
import StockRow from "./StockRow.vue";

export default {
    props: {
        items: Array,
        showQuantityAndVariance: Boolean,
        isCompleted: Boolean,
    },
    components: { StockRow },
    data() {
        return {
            stickyCategoryHeader: null,
            categoryHeaderPositions: [],
        };
    },
    mounted() {
        this.calculateCategoryHeaderPositions();
        this.handleScroll();
    },
    updated() {
        this.calculateCategoryHeaderPositions();
    },
    methods: {
        calculateCategoryHeaderPositions() {
            this.categoryHeaderPositions = [];
            const container = this.$refs.tableContainer;
            if (!container) return;

            const categoryHeaders = container.querySelectorAll('tr[data-category-header]');
            categoryHeaders.forEach((header, index) => {
                const rect = header.getBoundingClientRect();
                const containerRect = container.getBoundingClientRect();
                this.categoryHeaderPositions.push({
                    element: header,
                    category: header.dataset.category,
                    count: parseInt(header.dataset.count),
                    totalVariance: parseInt(header.dataset.totalVariance),
                    top: rect.top - containerRect.top + container.scrollTop,
                    bottom: rect.bottom - containerRect.top + container.scrollTop,
                });
            });
        },
        handleScroll() {
            const container = this.$refs.tableContainer;
            if (!container || this.categoryHeaderPositions.length === 0) return;

            const scrollTop = container.scrollTop;
            let currentHeader = null;

            // Only show sticky header if we've scrolled past the first category header
            if (scrollTop > 60 && this.categoryHeaderPositions.length > 0) {
                // Find the category header that should be sticky
                for (let i = this.categoryHeaderPositions.length - 1; i >= 0; i--) {
                    const header = this.categoryHeaderPositions[i];
                    if (scrollTop >= header.top - 60) { // 60px offset to account for table header height
                        currentHeader = header;
                        break;
                    }
                }
            }

            this.stickyCategoryHeader = currentHeader;
        },
    },
};
</script>
