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
                    :class="{
                        'bg-gray-800 text-white border-gray-800 dark:bg-white dark:text-gray-800 dark:border-white': filterVariancesOnly
                    }"
                    @click="filterVariances"
                >
                    {{ filterVariancesOnly ? 'Show All Items' : 'Filter Variances' }}
                </button>

                <button
                    v-if="completeStockPermission"
                    :disabled="isCompleted"
                    class="px-4 py-2 rounded-lg text-white"
                    :class="
                        isCompleted
                            ? 'bg-gray-400 cursor-not-allowed'
                            : 'bg-primary hover:bg-emerald-700'
                    "
                    @click="completeStockTaking"
                >
                    {{ isCompleted ? "Completed" : "Complete Stock Taking" }}
                </button>
            </div>
        </div>

        <!-- Stock Table -->
        <StockTable
            :items="filteredItems"
            :show-quantity-and-variance="showQuantityAndVariance"
            :stock-taking-id="stockTakingID"
            :is-completed="isCompleted"
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
            websocket: null,
            filterVariancesOnly: false,
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
        isCompleted() {
            return this.status === "Completed";
        },
        filteredItems() {
            if (!this.filterVariancesOnly) {
                return this.items;
            }
            
            return this.items.filter(
                (i) => {
                    const variance = (i.dispensary_count ?? 0) + (i.store_count ?? 0) - (i.snapshot_quantity ?? 0);
                    return variance !== 0;
                }
            );
        },
    },
    async mounted() {
        const el = document.getElementById("stock-taking-app");
        this.stockTakingID = el.dataset.stockTakingId;

        const res = await fetch(`/stock-taking/api/${this.stockTakingID}`);
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

        if (!this.isCompleted) {
            this.initWebSocket();
        }
    },
    beforeUnmount() {
        this.closeWebSocket();
    },
    methods: {
        formatDate,
        updateItem: debounce(async function (updatedItem) {
            try {
                // Prepare the payload
                const data = {
                    dispensary_count: updatedItem.dispensary_count,
                    store_count: updatedItem.store_count,
                    updated_expiry: updatedItem.expiry
                        ? formatToDateString(updatedItem.expiry)
                        : null,
                    notes: updatedItem.notes || "",
                };

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
            this.filterVariancesOnly = !this.filterVariancesOnly;
            
            if (this.filterVariancesOnly) {
                const itemsWithVariance = this.filteredItems;
                if (itemsWithVariance.length === 0) {
                    alert("No items with variance found");
                    this.filterVariancesOnly = false;
                }
            }
        },
        async completeStockTaking() {
            try {
                const res = await fetch(
                    `/stock-taking/api/${this.stockTakingID}`,
                    {
                        method: "POST",
                    },
                );

                if (!res.ok) {
                    const errData = await res.json();
                    console.error("Failed to complete stock taking:", errData);
                    alert(errData.error || "Failed to complete stock taking");
                    return;
                }

                const resp = await res.json();
                console.log("Stock taking completed successfully:", resp);
            } catch (err) {
                console.error("Error completing stock taking:", err);
                alert("Error completing stock taking");
            }
        },
        initWebSocket() {
            const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
            const wsUrl = `${protocol}//${window.location.host}/ws?stockTakingId=${this.stockTakingID}`;
            
            this.websocket = new WebSocket(wsUrl);
            
            this.websocket.onopen = () => {
                console.log('WebSocket connected');
            };
            
            this.websocket.onmessage = (event) => {
                const message = JSON.parse(event.data);
                this.handleWebSocketMessage(message);
            };
            
            this.websocket.onclose = () => {
                console.log('WebSocket disconnected');
                // Attempt to reconnect after 3 seconds
                setTimeout(() => {
                    if (!this.isCompleted) {
                        this.initWebSocket();
                    }
                }, 3000);
            };
            
            this.websocket.onerror = (error) => {
                console.error('WebSocket error:', error);
            };
        },
        
        closeWebSocket() {
            if (this.websocket) {
                this.websocket.close();
                this.websocket = null;
            }
        },
        
        handleWebSocketMessage(message) {
            if (message.type === 'stock_item_update') {
                const serverItem = message.data;
                const localItem = this.items.find(
                    (i) => i.product_id === serverItem.product_id,
                );

                // If user is editing this row, skip it
                if (localItem && localItem.isEditing) return;

                if (localItem) {
                    Object.assign(localItem, serverItem);
                }
            } else if (message.type === 'stock_taking_complete') {
                // Refresh the page data to show completion status
                window.location.reload();
            }
        },
    },
};
</script>
