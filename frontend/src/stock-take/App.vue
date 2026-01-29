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
import { formatDate, formatToDateString } from "@/utils/formatters";
import debounce from "lodash/debounce";

export default {
    components: { StockTable },
    data() {
        return {
            name: "",
            startDate: "",
            createdBy: "",
            status: "",
            items: [],
            stockTakingID: 0,
            showQuantityAndVariance: false,
            completeStockPermission: false,
        };
    },
    computed: {
        totalVariance() {
            return this.items.reduce(
                (acc, i) =>
                    acc +
                    ((i.dispensary_count ?? 0) +
                        (i.store_count ?? 0) -
                        (i.snapshot_quantity ?? 0)),
                0,
            );
        },
        totalIssues() {
            return this.items.filter(
                (i) =>
                    i.dispensary_count !== i.snapshot_quantity ||
                    i.store_count !== i.snapshot_quantity,
            ).length;
        },
    },
    async mounted() {
        const el = document.getElementById("stock-taking-app");
        this.stockTakingID = el.dataset.stockTakingId;

        const res = await fetch(`/stock-taking/api/${this.stockTakingID}`);
        const data = await res.json();

        console.log(data);
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
        updateItem: debounce(async function (updatedItem) {
            try {
                // Prepare the payload
                const data = {
                    dispensary_count: updatedItem.dispensary_count,
                    store_count: updatedItem.storeCount,
                    updated_expiry: updatedItem.expiry
                        ? formatToDateString(updatedItem.expiry)
                        : null,
                    notes: updatedItem.notes || "",
                };

                console.log(updatedItem);
                console.log(data);

                // Make the request
                const res = await fetch(
                    `/stock-taking/api/${this.stockTakingID}/item/${updatedItem.product_id}`,
                    {
                        method: "POST",
                        headers: {
                            "Content-Type": "application/json",
                        },
                        body: JSON.stringify(data),
                    },
                );

                if (!res.ok) {
                    const errData = await res.json();
                    console.error("Failed to update stock item:", errData);
                    alert(errData.error || "Failed to update item");
                    return;
                }

                const resp = await res.json();
                console.log("Stock item updated successfully:", resp);
            } catch (err) {
                console.error("Error updating stock item:", err);
                alert("Error updating stock item");
            }
        }, 1000),
        filterVariances() {
            alert("filtering variances unimplemented");
        },
        completeStockTaking() {
            alert("implement complete stock taking");
        },
    },
};
</script>
