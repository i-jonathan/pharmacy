<template>
    <div class="p-6 space-y-6">
        <!-- Header -->
        <div
            class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between"
        >
            <div>
                <h1 class="text-2xl font-semibold">Stock Taking: {{ name }}</h1>
                <p class="text-sm text-gray-500 dark:text-gray-400">
                    Started: {{ formatDate(startDate) }} · Created by
                    {{ createdBy }} ·
                    <span class="text-primary font-medium">{{ status }}</span>
                </p>
            </div>

            <div class="flex gap-3 flex-col sm:flex-row">
                <button
                    v-if="showQuantityAndVariance"
                    class="px-4 py-2 rounded-lg border border-gray-300 dark:border-gray-600 hover:bg-gray-100 dark:hover:bg-gray-700"
                    @click="filterVariances"
                >
                    Filter Variances
                </button>

                <button
                    v-if="completeStockPermission"
                    class="bg-primary text-white px-4 py-2 rounded-lg hover:bg-emerald-700"
                    @click="completeStockTaking"
                >
                    Complete Stock Take
                </button>
            </div>
        </div>

        <!-- Stock Table -->
        <StockTable
            :items="items"
            :show-quantity-and-variance="showQuantityAndVariance"
            @update-item="updateItem"
        />

        <!-- Summary -->
        <div
            v-if="showQuantityAndVariance"
            class="flex justify-between text-sm text-gray-600 dark:text-gray-400"
        >
            <div>
                Total Variance:
                <span
                    class="font-medium"
                    :class="{
                        'text-red-600': totalVariance < 0,
                        'text-green-600': totalVariance > 0,
                    }"
                    >{{ totalVariance }}</span
                >
            </div>

            <div>
                Issues noted: <span class="font-medium">{{ totalIssues }}</span>
            </div>
        </div>
    </div>
</template>

<script>
import StockTable from "./components/StockTable.vue";
import { formatDate } from "@/utils/formatters";

export default {
    components: { StockTable },
    data() {
        return {
            name: "",
            startDate: "",
            createdBy: "",
            status: "",
            items: [],
            showQuantityAndVariance: false,
            completeStockPermission: false,
        };
    },
    computed: {
        totalVariance() {
            return this.items.reduce(
                (acc, i) =>
                    acc +
                    ((i.dispCount ?? 0) + (i.storeCount ?? 0) - (i.stock ?? 0)),
                0,
            );
        },
        totalIssues() {
            return this.items.filter(
                (i) => i.dispCount !== i.stock || i.storeCount !== i.stock,
            ).length;
        },
    },
    async mounted() {
        const el = document.getElementById("stock-taking-app");
        const stockTakingID = el.dataset.stockTakingId;

        const res = await fetch(`/stock-taking/api/${stockTakingID}`);
        const data = await res.json();

        const stData = data.stock_taking_data;
        this.name = stData.name;
        this.startDate = stData.started_at;
        this.createdBy = stData.created_by;
        this.status = stData.status;

        this.items = data.items;

        const permissions = data.permissions || {};
        this.showQuantityAndVariance = permissions["stock:view"];
        this.completeStockPermission = permissions["stock:complete"];
    },
    methods: {
        formatDate,
        updateItem(updatedItem) {
            //send to backend
            console.log("updating stuff");
        },
        filterVariances() {
            alert("filtering variances unimplemented");
        },
        completeStockTaking() {
            alert("implement complete stock taking");
        },
    },
};
</script>
