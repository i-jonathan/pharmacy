<template>
    <div class="space-y-6">
        <!-- Header -->
        <div class="flex items-end justify-between">
            <h1 class="text-2xl font-semibold">
                Received Items History
            </h1>
            <div class="text-sm text-gray-500 dark:text-gray-400">
                Use ↑/↓ to browse batches; Enter to open/close details
            </div>
        </div>

        <!-- Filters -->
        <div
            class="bg-white dark:bg-gray-800 p-3 rounded-lg shadow flex flex-wrap items-center justify-between gap-3"
        >
            <div class="flex flex-wrap items-center gap-3">
                <div class="flex flex-col">
                    <label
                        class="text-xs font-medium text-gray-600 dark:text-gray-300 mb-1"
                    >
                        Start Date
                    </label>
                    <input
                        type="date"
                        v-model="startDate"
                        @change="handleDateChange"
                        class="border border-gray-300 dark:border-gray-600 rounded px-2 py-1.5 bg-white dark:bg-gray-700 text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500"
                    />
                </div>
                <div class="flex flex-col">
                    <label
                        class="text-xs font-medium text-gray-600 dark:text-gray-300 mb-1"
                    >
                        End Date
                    </label>
                    <input
                        type="date"
                        v-model="endDate"
                        @change="handleDateChange"
                        class="border border-gray-300 dark:border-gray-600 rounded px-2 py-1.5 bg-white dark:bg-gray-700 text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500"
                    />
                </div>
                <div class="flex flex-col">
                    <label
                        class="text-xs font-medium text-gray-600 dark:text-gray-300 mb-1"
                    >
                        Quick Range
                    </label>
                    <select
                        v-model="quickRange"
                        @change="handleQuickRange"
                        class="border border-gray-300 dark:border-gray-600 rounded px-2 py-1.5 bg-white dark:bg-gray-700 text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500"
                    >
                        <option value="today">Today</option>
                        <option value="yesterday">Yesterday</option>
                        <option value="this-week">This Week</option>
                        <option value="last-week">Last Week</option>
                        <option value="this-month" selected>This Month</option>
                        <option value="last-month">Last Month</option>
                        <option value="all">All</option>
                        <option value="custom">Custom</option>
                    </select>
                </div>
            </div>
        </div>

        <!-- Loading State -->
        <div
            v-if="loading"
            class="text-center py-16 text-gray-500 dark:text-gray-400"
        >
            Loading...
        </div>

        <!-- Error State -->
        <div
            v-else-if="fetchError"
            class="text-center py-16 text-red-500 dark:text-red-400"
        >
            {{ fetchError }}
        </div>

        <!-- Empty State -->
        <div
            v-else-if="batches.length === 0"
            class="text-center py-16 text-gray-500 dark:text-gray-400"
        >
            <p class="text-lg mb-2">No received items yet.</p>
        </div>

        <!-- Table -->
        <div
            v-else
            class="bg-white dark:bg-gray-800 rounded-xl shadow overflow-hidden border border-gray-200 dark:border-gray-700"
        >
            <div class="overflow-x-auto">
                <table class="w-full text-sm text-left">
                    <thead
                        class="text-xs uppercase tracking-wider text-gray-500 dark:text-gray-400 bg-gray-50 dark:bg-gray-700/50"
                    >
                        <tr>
                            <th class="px-6 py-3">Date</th>
                            <th class="px-6 py-3">Supplier</th>
                            <th class="px-6 py-3">Received By</th>
                            <th class="px-6 py-3">Items Count</th>
                            <th class="px-6 py-3 text-right">Total Cost</th>
                        </tr>
                    </thead>
                    <tbody
                        class="divide-y divide-gray-100 dark:divide-gray-700"
                    >
                        <tr
                            v-for="(batch, idx) in batches"
                            :key="batch.id"
                            @click="openPanel(idx)"
                            :class="[
                                'cursor-pointer transition-colors',
                                selectedIndex === idx
                                    ? 'bg-emerald-50 dark:bg-emerald-900/20 font-semibold ring-2 ring-emerald-400'
                                    : 'hover:bg-emerald-50 dark:hover:bg-emerald-900/20',
                            ]"
                        >
                            <td
                                class="px-6 py-4 text-gray-600 dark:text-gray-300 whitespace-nowrap"
                            >
                                {{ formatDate(batch.created_at) }}
                            </td>
                            <td
                                class="px-6 py-4 font-medium text-gray-900 dark:text-gray-100"
                            >
                                {{ batch.supplier_name }}
                            </td>
                            <td
                                class="px-6 py-4 text-gray-600 dark:text-gray-300"
                            >
                                {{ batch.received_by }}
                            </td>
                            <td
                                class="px-6 py-4 text-gray-600 dark:text-gray-300"
                            >
                                {{ totalItems(batch) }}
                            </td>
                            <td
                                class="px-6 py-4 text-gray-900 dark:text-gray-100 text-right font-medium"
                            >
                                {{ formatPrice(totalCost(batch)) }}
                            </td>
                        </tr>
                    </tbody>
                </table>
            </div>
        </div>

        <!-- Side Panel -->
        <div
            v-if="panelOpen"
            class="fixed inset-0 z-40 bg-black/20"
            @click="closePanel"
        ></div>
        <aside
            :class="[
                'panel fixed right-0 top-0 h-full w-full md:w-1/2 bg-white dark:bg-gray-800 border-l border-gray-200 dark:border-gray-700 shadow-xl z-50 overflow-y-auto',
                panelOpen ? 'open' : '',
            ]"
        >
            <div class="p-6 relative">
                <button
                    @click="closePanel"
                    class="absolute top-4 right-4 text-gray-500 hover:text-gray-800 dark:hover:text-gray-200 text-xl leading-none"
                    aria-label="Close details"
                >
                    ✕
                </button>

                <h2
                    class="text-xl font-bold text-emerald-600 dark:text-emerald-400 border-b border-emerald-200 dark:border-emerald-700 pb-2"
                >
                    {{ selectedBatch ? selectedBatch.supplier_name : "" }}
                </h2>

                <!-- Batch Metadata -->
                <p
                    v-if="selectedBatch"
                    class="text-sm text-gray-500 dark:text-gray-400 mb-4 mt-2"
                >
                    {{ formatDate(selectedBatch.created_at) }} • Received by: {{ selectedBatch.received_by }}<template v-if="selectedBatch.note"> • {{ selectedBatch.note }}</template>
                </p>

                <!-- Items Sub-Table -->
                <section class="mt-6">
                    <h3
                        class="text-lg font-semibold text-gray-700 dark:text-gray-200 mb-4"
                    >
                        Items in Batch
                    </h3>
                    <div class="overflow-x-auto">
                        <table
                            class="min-w-full text-sm border border-gray-200 dark:border-gray-700"
                        >
                            <thead class="bg-gray-50 dark:bg-gray-700">
                                <tr>
                                    <th
                                        class="px-4 py-3 text-left font-semibold text-gray-700 dark:text-gray-200 uppercase tracking-wide"
                                    >
                                        Product
                                    </th>
                                    <th
                                        class="px-4 py-3 text-right font-semibold text-gray-700 dark:text-gray-200 uppercase tracking-wide"
                                    >
                                        Qty
                                    </th>
                                    <th
                                        class="px-4 py-3 text-left font-semibold text-gray-700 dark:text-gray-200 uppercase tracking-wide"
                                    >
                                        Expiry
                                    </th>
                                    <th
                                        class="px-4 py-3 text-right font-semibold text-gray-700 dark:text-gray-200 uppercase tracking-wide"
                                    >
                                        Unit Cost Price
                                    </th>
                                    <th
                                        class="px-4 py-3 text-right font-semibold text-gray-700 dark:text-gray-200 uppercase tracking-wide"
                                    >
                                        Total Cost Price
                                    </th>
                                </tr>
                            </thead>
                            <tbody
                                v-if="
                                    selectedBatch &&
                                    selectedBatch.items.length > 0
                                "
                                class="divide-y divide-gray-200 dark:divide-gray-700 bg-white dark:bg-gray-800"
                            >
                                <tr
                                    v-for="(item, idx) in selectedBatch.items"
                                    :key="idx"
                                >
                                    <td
                                        class="px-4 py-3 text-gray-900 dark:text-gray-100"
                                    >
                                        <div class="font-medium">{{ item.product_name }}</div>
                                        <div class="text-xs text-gray-500">{{ item.manufacturer || "—" }}</div>
                                    </td>
                                    <td
                                        class="px-4 py-3 text-right text-gray-600 dark:text-gray-300"
                                    >
                                        {{ item.quantity }}
                                    </td>
                                    <td
                                        class="px-4 py-3 text-gray-600 dark:text-gray-300"
                                    >
                                        {{ formatExpiry(item.expiry_date) }}
                                    </td>
                                    <td
                                        class="px-4 py-3 text-right text-gray-600 dark:text-gray-300"
                                    >
                                        {{ formatPrice(item.cost_price) }}
                                    </td>
                                    <td
                                        class="px-4 py-3 text-right text-gray-600 dark:text-gray-300"
                                    >
                                        {{ formatPrice(item.cost_price * item.quantity) }}
                                    </td>
                                </tr>
                            </tbody>
                            <tfoot
                                v-if="selectedBatch"
                                class="bg-emerald-50 dark:bg-emerald-900/40 font-semibold"
                            >
                                <tr>
                                    <td
                                        class="px-4 py-3 text-gray-700 dark:text-gray-200"
                                        colspan="4"
                                    >
                                        Item Total
                                    </td>
                                    <td
                                        class="px-4 py-3 text-right text-emerald-700 dark:text-emerald-400"
                                    >
                                        {{
                                            formatPrice(
                                                totalCost(selectedBatch),
                                            )
                                        }}
                                    </td>
                                </tr>
                            </tfoot>
                        </table>
                    </div>
                </section>
            </div>
        </aside>
    </div>
</template>

<script>
export default {
    data() {
        const now = new Date();
        const firstOfMonth = new Date(now.getFullYear(), now.getMonth(), 1);

        return {
            batches: [],
            loading: true,
            fetchError: null,
            selectedIndex: -1,
            panelOpen: false,
            quickRange: "this-month",
            startDate: firstOfMonth.toISOString().split("T")[0],
            endDate: now.toISOString().split("T")[0],
        };
    },

    computed: {
        selectedBatch() {
            if (this.selectedIndex < 0 || this.selectedIndex >= this.batches.length) return null;
            return this.batches[this.selectedIndex];
        },
    },

    mounted() {
        this.fetchBatches();
        document.addEventListener("keydown", this.handleKeydown);
    },

    beforeUnmount() {
        document.removeEventListener("keydown", this.handleKeydown);
    },

    methods: {
        async fetchBatches() {
            this.loading = true;
            this.fetchError = null;
            try {
                const params = new URLSearchParams();
                if (this.startDate) params.set("start", this.startDate);
                if (this.endDate) params.set("end", this.endDate);

                const url = `/inventory/received-items-history/api?${params.toString()}`;
                const res = await fetch(url);
                if (!res.ok) {
                    const err = await res.json();
                    throw new Error(
                        err.message || "Failed to load received items history",
                    );
                }
                const data = await res.json();
                this.batches = data.batches || [];
            } catch (err) {
                this.fetchError = err.message;
            } finally {
                this.loading = false;
            }
        },

        applyFilter() {
            this.selectedIndex = -1;
            this.panelOpen = false;
            this.fetchBatches();
        },

        handleDateChange() {
            this.quickRange = "custom";
            this.applyFilter();
        },

        handleQuickRange() {
            const today = new Date();
            let start, end;

            switch (this.quickRange) {
                case "today":
                    start = new Date(today);
                    end = new Date(today);
                    break;
                case "yesterday":
                    start = new Date(today);
                    start.setDate(start.getDate() - 1);
                    end = new Date(start);
                    break;
                case "this-week":
                    start = new Date(today);
                    start.setDate(today.getDate() - today.getDay());
                    end = new Date(today);
                    break;
                case "last-week":
                    start = new Date(today);
                    start.setDate(today.getDate() - today.getDay() - 7);
                    end = new Date(start);
                    end.setDate(start.getDate() + 6);
                    break;
                case "this-month":
                    start = new Date(today.getFullYear(), today.getMonth(), 1);
                    end = new Date(today);
                    break;
                case "last-month":
                    start = new Date(today.getFullYear(), today.getMonth() - 1, 1);
                    end = new Date(today.getFullYear(), today.getMonth(), 0);
                    break;
                case "all":
                    this.startDate = "";
                    this.endDate = "";
                    this.applyFilter();
                    return;
                case "custom":
                    return;
            }

            this.startDate = this.formatDateValue(start);
            this.endDate = this.formatDateValue(end);
            this.applyFilter();
        },

        formatDateValue(date) {
            return date.toISOString().split("T")[0];
        },

        selectBatch(idx) {
            this.selectedIndex = idx;
        },

        openPanel(idx) {
            this.selectedIndex = idx;
            this.panelOpen = true;
        },

        closePanel() {
            this.panelOpen = false;
        },

        handleKeydown(e) {
            if (e.key === "Escape" && this.panelOpen) {
                this.closePanel();
            }
            if (e.key === "ArrowDown") {
                e.preventDefault();
                const next = Math.min(
                    this.selectedIndex === -1 ? 0 : this.selectedIndex + 1,
                    this.batches.length - 1,
                );
                this.selectBatch(next);
            }
            if (e.key === "ArrowUp") {
                e.preventDefault();
                const prev = Math.max(
                    this.selectedIndex === -1 ? 0 : this.selectedIndex - 1,
                    0,
                );
                this.selectBatch(prev);
            }
            if (e.key === "Enter") {
                if (this.selectedIndex >= 0) {
                    if (this.panelOpen) {
                        this.closePanel();
                    } else {
                        this.openPanel(this.selectedIndex);
                    }
                }
            }
        },

        formatDate(dateStr) {
            if (!dateStr) return "—";
            const d = new Date(dateStr);
            return d.toLocaleDateString("en-GB", {
                day: "numeric",
                month: "short",
                year: "numeric",
                hour: "2-digit",
                minute: "2-digit",
            });
        },

        formatPrice(costKobo) {
            if (costKobo == null) return "₦0.00";
            const value = Number(costKobo) / 100;
            return `₦${value.toLocaleString("en-US", {
                minimumFractionDigits: 2,
                maximumFractionDigits: 2,
            })}`;
        },

        formatExpiry(dateStr) {
            if (!dateStr) return "—";
            const d = new Date(dateStr);
            return d.toLocaleDateString("en-GB", {
                month: "short",
                year: "numeric",
            });
        },

        totalItems(batch) {
            if (!batch.items || !batch.items.length) return 0;
            return batch.items.reduce(
                (sum, item) => sum + Number(item.quantity || 0),
                0,
            );
        },

        totalCost(batch) {
            if (!batch.items || !batch.items.length) return 0;
            return batch.items.reduce(
                (sum, item) =>
                    sum +
                    Number(item.cost_price || 0) * Number(item.quantity || 0),
                0,
            );
        },
    },
};
</script>

<style>
/* Slide-in side panel */
.panel {
    transform: translateX(100%);
    transition: transform 0.3s ease;
}
.panel.open {
    transform: translateX(0);
}
</style>
