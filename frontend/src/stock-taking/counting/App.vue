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

        <!-- Mobile Search Bar (full width above table) -->
        <div class="md:hidden">
            <div class="relative">
                <input
                    type="text"
                    v-model="searchQuery"
                    placeholder="Search items..."
                    class="w-full pl-10 pr-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary dark:bg-gray-700 dark:text-white"
                />
                <svg
                    class="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-gray-400"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                >
                    <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
                    />
                </svg>
            </div>
        </div>

        <!-- Desktop Search Bar (in header) -->
        <div class="hidden md:flex md:justify-start">
            <div class="relative w-full max-w-md">
                <input
                    type="text"
                    v-model="searchQuery"
                    placeholder="Search items..."
                    class="w-full pl-10 pr-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary dark:bg-gray-700 dark:text-white"
                />
                <svg
                    class="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-gray-400"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                >
                    <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
                    />
                </svg>
            </div>
        </div>

        <!-- Filter Results Indicator -->
        <div
            v-if="isFiltering"
            class="text-sm text-gray-600 dark:text-gray-400 mb-4"
        >
            Showing {{ filteredItemsCount }} of {{ items.length }} items
            <span v-if="searchQueryDebounced.trim()">
                for "{{ searchQueryDebounced }}"
            </span>
            <span v-if="filterVariancesOnly">
                with variance
            </span>
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
            searchQuery: "",
            searchQueryDebounced: "",
            searchIndex: new Map(), // Maps search terms to item indices
            groupByCategory: true,
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
            let items = this.items;
            
            // Apply search filter first (using indexed search for performance)
            if (this.searchQueryDebounced.trim()) {
                items = this.searchWithIndex(this.searchQueryDebounced);
            }
            
            // Apply variance filter to search results
            if (this.filterVariancesOnly) {
                items = items.filter(
                    (i) => {
                        const variance = (i.dispensary_count ?? 0) + (i.store_count ?? 0) - (i.snapshot_quantity ?? 0);
                        return variance !== 0;
                    }
                );
            }
            
            // Add category headers if grouping is enabled
            if (this.groupByCategory) {
                const result = [];
                const groups = {};
                
                items.forEach(item => {
                    const category = item.category || 'Uncategorized';
                    if (!groups[category]) {
                        groups[category] = [];
                    }
                    groups[category].push(item);
                });
                
                Object.entries(groups).forEach(([category, categoryItems]) => {
                    const totalVariance = categoryItems.reduce((sum, i) => 
                        sum + ((i.dispensary_count ?? 0) + (i.store_count ?? 0) - (i.snapshot_quantity ?? 0)), 0
                    );
                    
                    // Add category header
                    result.push({
                        isCategoryHeader: true,
                        category,
                        count: categoryItems.length,
                        totalVariance
                    });
                    
                    // Add items for this category
                    result.push(...categoryItems);
                });
                
                return result;
            }
            
            return items;
        },
        filteredItemsCount() {
            return this.filteredItems.length;
        },
        isFiltering() {
            return this.filterVariancesOnly || this.searchQueryDebounced.trim() !== '';
        },
    },
    watch: {
        searchQuery: {
            handler: debounce(function(newQuery) {
                this.searchQueryDebounced = newQuery;
            }, 300),
            immediate: true,
        },
        items: {
            handler() {
                this.buildSearchIndex();
            },
            deep: true,
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

        // Build search index after items are loaded
        this.buildSearchIndex();

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
        buildSearchIndex() {
            this.searchIndex.clear();
            
            this.items.forEach((item, index) => {
                // Index product name
                const productName = item.product_name.toLowerCase();
                this.indexTerms(productName, index);
                
                // Index manufacturer
                const manufacturer = item.manufacturer.toLowerCase();
                this.indexTerms(manufacturer, index);
            });
        },
        indexTerms(text, itemIndex) {
            // Split text into words and index each word and its prefixes
            const words = text.split(/\s+/);
            
            words.forEach(word => {
                // Index the full word
                if (!this.searchIndex.has(word)) {
                    this.searchIndex.set(word, new Set());
                }
                this.searchIndex.get(word).add(itemIndex);
                
                // Index all prefixes for autocomplete-like functionality
                for (let i = 1; i <= word.length; i++) {
                    const prefix = word.substring(0, i);
                    if (!this.searchIndex.has(prefix)) {
                        this.searchIndex.set(prefix, new Set());
                    }
                    this.searchIndex.get(prefix).add(itemIndex);
                }
            });
        },
        searchWithIndex(query) {
            if (!query.trim()) {
                return this.items;
            }
            
            const searchTerms = query.toLowerCase().trim().split(/\s+/);
            const resultSets = [];
            
            searchTerms.forEach(term => {
                if (this.searchIndex.has(term)) {
                    resultSets.push(this.searchIndex.get(term));
                }
            });
            
            // If no terms found in index, return empty array
            if (resultSets.length === 0) {
                return [];
            }
            
            // Find intersection of all result sets (items that match all terms)
            const intersection = new Set(resultSets[0]);
            for (let i = 1; i < resultSets.length; i++) {
                for (const itemIndex of intersection) {
                    if (!resultSets[i].has(itemIndex)) {
                        intersection.delete(itemIndex);
                    }
                }
            }
            
            // Convert indices back to items
            return Array.from(intersection).map(index => this.items[index]);
        },
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
