<template>
    <div class="p-4 md:p-6">
        <!-- Loading State -->
        <div v-if="loading" class="text-center py-12">
            <div class="text-lg text-gray-600 dark:text-gray-400">
                Loading dashboard data...
            </div>
        </div>

        <!-- Error State -->
        <div v-else-if="error" class="text-center py-12">
            <div class="text-lg text-red-600 dark:text-red-400 mb-4">
                {{ error }}
            </div>
            <button
                @click="fetchDashboardData"
                class="bg-primary hover:bg-primary/90 text-white px-4 py-2 rounded-lg transition-colors"
            >
                Retry
            </button>
        </div>

        <!-- Dashboard Content -->
        <div v-else>
            <!-- KPI Cards -->
            <div
                class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 md:gap-6 mb-6"
            >
                <div
                    class="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-4 md:p-6 cursor-pointer hover:shadow-md transition-shadow"
                    @click="navigateToInventory()"
                >
                    <h3
                        class="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-2"
                    >
                        Total Products
                    </h3>
                    <p class="text-2xl md:text-3xl font-bold text-primary">
                        {{ getKPIValue("total_inventory", 0).toLocaleString() }}
                    </p>
                    <p class="text-sm text-gray-600 dark:text-gray-400">
                        Products in inventory
                    </p>
                </div>

                <div
                    class="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-4 md:p-6"
                >
                    <h3
                        class="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-2"
                    >
                        Today's Sales
                    </h3>
                    <p
                        class="text-2xl md:text-3xl font-bold text-blue-600 dark:text-blue-400"
                    >
                        {{
                            getKPIValue("today_sales", 0)
                                ? formatCurrency(getKPIValue("today_sales", 0))
                                : "🔒 Restricted"
                        }}
                    </p>
                    <p class="text-sm text-gray-600 dark:text-gray-400">
                        {{
                            getKPIValue("today_sales", 0)
                                ? "Total sales today"
                                : "Permission required"
                        }}
                    </p>
                    <div class="mt-2 flex items-center text-sm">
                        <span
                            :class="
                                getTrendColor(getKPIValue('sales_trend', 0))
                            "
                        >
                            {{ getTrendIcon(getKPIValue("sales_trend", 0)) }}
                        </span>
                        <span
                            :class="
                                getTrendColor(getKPIValue('sales_trend', 0))
                            "
                            class="ml-1"
                        >
                            {{
                                getKPIValue("sales_trend", 0)
                                    ? Math.abs(
                                          getKPIValue("sales_trend", 0),
                                      ).toFixed(1) + "%"
                                    : ""
                            }}
                        </span>
                    </div>
                </div>

                <div
                    class="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-4 md:p-6 cursor-pointer hover:shadow-md transition-shadow"
                    @click="navigateToSalesHistory()"
                >
                    <h3
                        class="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-2"
                    >
                        Today's Orders
                    </h3>
                    <div class="flex items-baseline gap-2">
                        <p
                            class="text-2xl md:text-3xl font-bold text-purple-600 dark:text-purple-400"
                        >
                            {{
                                getKPIValue(
                                    "today_transactions",
                                    0,
                                ).toLocaleString()
                            }}
                        </p>
                        <span
                            :class="
                                getTrendColor(
                                    getKPIValue('transaction_trend', 0),
                                )
                            "
                            class="text-sm font-medium"
                        >
                            {{
                                getTrendIcon(
                                    getKPIValue("transaction_trend", 0),
                                )
                            }}
                            {{
                                Math.abs(
                                    getKPIValue("transaction_trend", 0),
                                ).toFixed(1)
                            }}%
                        </span>
                    </div>
                    <p class="text-sm text-gray-600 dark:text-gray-400">
                        Orders processed today
                    </p>
                </div>

                <div
                    class="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-4 md:p-6 cursor-pointer hover:shadow-md transition-shadow"
                    @click="openLowStockModal()"
                >
                    <h3
                        class="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-2"
                    >
                        Low Stock Items
                    </h3>
                    <p
                        class="text-2xl md:text-3xl font-bold text-red-600 dark:text-red-400"
                    >
                        {{ getKPIValue("low_stock_count", 0).toLocaleString() }}
                    </p>
                    <p class="text-sm text-gray-600 dark:text-gray-400">
                        Need reordering
                    </p>
                </div>
            </div>

            <!-- Charts Row -->
            <div class="grid grid-cols-1 lg:grid-cols-2 gap-4 md:gap-6 mb-6">
                <!-- Sales Trend Chart - Only show if user has sales data -->
                <div
                    v-if="getKPIValue('today_sales', 0) !== null"
                    class="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-4 md:p-6"
                >
                    <h3
                        class="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-4"
                    >
                        Sales Trend (7 days)
                    </h3>
                    <div class="h-48 md:h-64 relative">
                        <canvas
                            ref="salesTrendChart"
                            class="w-full h-full"
                        ></canvas>
                    </div>
                </div>

                <!-- Category Sales Pie Chart -->
                <div
                    class="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-4 md:p-6"
                >
                    <h3
                        class="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-4"
                    >
                        Sales by Category
                    </h3>
                    <div class="h-48 md:h-64 relative">
                        <canvas
                            ref="categoryChart"
                            class="w-full h-full"
                        ></canvas>
                    </div>
                </div>
            </div>

            <!-- Expiring Items Alert -->
            <div class="grid grid-cols-1 lg:grid-cols-2 gap-4 md:gap-6 mb-6">
                <!-- Critical Expiry (≤ 30 days) -->
                <div
                    class="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-4 md:p-6"
                >
                    <h3
                        class="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-4 flex items-center justify-between"
                    >
                        <div class="flex items-center">
                            <span
                                class="w-3 h-3 bg-red-500 rounded-full mr-2"
                            ></span>
                            Critical Expiry (≤ 30 days)
                        </div>
                        <button
                            v-if="
                                expiringItems.filter(
                                    (i) => i.daysUntilExpiry <= 30,
                                ).length > 0
                            "
                            @click="
                                copyExpiringItems(
                                    expiringItems.filter(
                                        (i) => i.daysUntilExpiry <= 30,
                                    ),
                                    'Critical',
                                )
                            "
                            class="text-sm bg-red-100 hover:bg-red-200 dark:bg-red-900/30 dark:hover:bg-red-900/50 text-red-700 dark:text-red-300 px-3 py-1 rounded-md font-medium transition-colors"
                        >
                            📋 Copy All
                        </button>
                    </h3>
                    <div class="space-y-3">
                        <div
                            v-for="item in criticalExpiringItems"
                            :key="item.id"
                            class="flex items-center justify-between p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg"
                        >
                            <div class="flex-1">
                                <div
                                    class="text-sm font-medium text-gray-900 dark:text-gray-100"
                                >
                                    {{ item.name }}
                                </div>
                                <div
                                    class="text-xs text-gray-600 dark:text-gray-400"
                                >
                                    Qty: {{ item.quantity }} | Expires:
                                    {{ item.expiryDate }}
                                </div>
                            </div>
                            <div class="text-right">
                                <div
                                    class="text-sm font-bold text-red-600 dark:text-red-400"
                                >
                                    {{ item.daysUntilExpiry }} days
                                </div>
                            </div>
                        </div>
                        <div
                            v-if="
                                expiringItems.filter(
                                    (i) => i.daysUntilExpiry <= 30,
                                ).length === 0
                            "
                            class="text-center py-4 text-gray-500 dark:text-gray-400"
                        >
                            No items expiring within 30 days
                        </div>
                        <div
                            v-if="
                                expiringItems.filter(
                                    (i) => i.daysUntilExpiry <= 30,
                                ).length > 5
                            "
                            class="text-center"
                        >
                            <button
                                @click="openCriticalModal()"
                                class="text-sm text-red-600 dark:text-red-400 hover:text-red-700 dark:hover:text-red-300 font-medium"
                            >
                                View All ({{
                                    expiringItems.filter(
                                        (i) => i.daysUntilExpiry <= 30,
                                    ).length
                                }}
                                items)
                            </button>
                        </div>
                    </div>
                </div>

                <!-- Warning Expiry (31-90 days) -->
                <div
                    class="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-4 md:p-6"
                >
                    <h3
                        class="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-4 flex items-center justify-between"
                    >
                        <div class="flex items-center">
                            <span
                                class="w-3 h-3 bg-amber-500 rounded-full mr-2"
                            ></span>
                            Warning Expiry (31-90 days)
                        </div>
                        <button
                            v-if="
                                expiringItems.filter(
                                    (i) =>
                                        i.daysUntilExpiry > 30 &&
                                        i.daysUntilExpiry <= 90,
                                ).length > 0
                            "
                            @click="
                                copyExpiringItems(
                                    expiringItems.filter(
                                        (i) =>
                                            i.daysUntilExpiry > 30 &&
                                            i.daysUntilExpiry <= 90,
                                    ),
                                    'Warning',
                                )
                            "
                            class="text-sm bg-amber-100 hover:bg-amber-200 dark:bg-amber-900/30 dark:hover:bg-amber-900/50 text-amber-700 dark:text-amber-300 px-3 py-1 rounded-md font-medium transition-colors"
                        >
                            📋 Copy All
                        </button>
                    </h3>
                    <div class="space-y-3">
                        <div
                            v-for="item in warningExpiringItems"
                            :key="item.id"
                            class="flex items-center justify-between p-3 bg-amber-50 dark:bg-amber-900/20 border border-amber-200 dark:border-amber-800 rounded-lg"
                        >
                            <div class="flex-1">
                                <div
                                    class="text-sm font-medium text-gray-900 dark:text-gray-100"
                                >
                                    {{ item.name }}
                                </div>
                                <div
                                    class="text-xs text-gray-600 dark:text-gray-400"
                                >
                                    Qty: {{ item.quantity }} | Expires:
                                    {{ item.expiryDate }}
                                </div>
                            </div>
                            <div class="text-right">
                                <div
                                    class="text-sm font-bold text-amber-600 dark:text-amber-400"
                                >
                                    {{ item.daysUntilExpiry }} days
                                </div>
                            </div>
                        </div>
                        <div
                            v-if="
                                expiringItems.filter(
                                    (i) =>
                                        i.daysUntilExpiry > 30 &&
                                        i.daysUntilExpiry <= 90,
                                ).length === 0
                            "
                            class="text-center py-4 text-gray-500 dark:text-gray-400"
                        >
                            No items expiring within 31-90 days
                        </div>
                        <div
                            v-if="
                                expiringItems.filter(
                                    (i) =>
                                        i.daysUntilExpiry > 30 &&
                                        i.daysUntilExpiry <= 90,
                                ).length > 5
                            "
                            class="text-center"
                        >
                            <button
                                @click="openWarningModal()"
                                class="text-sm text-amber-600 dark:text-amber-400 hover:text-amber-700 dark:hover:text-amber-300 font-medium"
                            >
                                View All ({{
                                    expiringItems.filter(
                                        (i) =>
                                            i.daysUntilExpiry > 30 &&
                                            i.daysUntilExpiry <= 90,
                                    ).length
                                }}
                                items)
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>

    <!-- Toast Notification -->
    <div
        v-if="showToast"
        class="fixed top-4 right-4 bg-green-500 text-white px-4 py-2 rounded-lg shadow-lg z-50 transition-all duration-300 transform"
    >
        {{ toastMessage }}
    </div>

    <!-- Critical Expiry Modal -->
    <div
        v-if="showCriticalModal"
        class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4"
    >
        <div
            class="bg-white dark:bg-gray-800 rounded-lg shadow-xl max-w-2xl w-full max-h-[80vh] overflow-hidden"
        >
            <div class="p-6 border-b border-gray-200 dark:border-gray-700">
                <div class="flex items-center justify-between">
                    <h3
                        class="text-lg font-semibold text-gray-900 dark:text-gray-100 flex items-center"
                    >
                        <span
                            class="w-3 h-3 bg-red-500 rounded-full mr-2"
                        ></span>
                        All Critical Expiry Items (≤ 30 days)
                    </h3>
                    <button
                        @click="showCriticalModal = false"
                        class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
                    >
                        ✕
                    </button>
                </div>
            </div>
            <div class="p-6 overflow-y-auto max-h-[60vh]">
                <div class="space-y-3">
                    <div
                        v-for="item in allCriticalItems"
                        :key="item.id"
                        class="flex items-center justify-between p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg"
                    >
                        <div class="flex-1">
                            <div
                                class="text-sm font-medium text-gray-900 dark:text-gray-100"
                            >
                                {{ item.name }}
                            </div>
                            <div
                                class="text-xs text-gray-600 dark:text-gray-400"
                            >
                                Qty: {{ item.quantity }} | Expires:
                                {{ item.expiryDate }}
                            </div>
                        </div>
                        <div class="text-right">
                            <div
                                class="text-sm font-bold text-red-600 dark:text-red-400"
                            >
                                {{ item.daysUntilExpiry }} days
                            </div>
                        </div>
                    </div>
                </div>
                <div
                    v-if="allCriticalItems.length === 0"
                    class="text-center py-8 text-gray-500 dark:text-gray-400"
                >
                    No items expiring within 30 days
                </div>
            </div>
            <div class="p-6 border-t border-gray-200 dark:border-gray-700">
                <button
                    @click="
                        copyExpiringItems(allCriticalItems, 'Critical');
                        showCriticalModal = false;
                    "
                    class="w-full bg-red-600 hover:bg-red-700 text-white font-medium py-2 px-4 rounded-lg transition-colors"
                >
                    📋 Copy All Critical Items
                </button>
            </div>
        </div>
    </div>

    <!-- Warning Expiry Modal -->
    <div
        v-if="showWarningModal"
        class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4"
    >
        <div
            class="bg-white dark:bg-gray-800 rounded-lg shadow-xl max-w-2xl w-full max-h-[80vh] overflow-hidden"
        >
            <div class="p-6 border-b border-gray-200 dark:border-gray-700">
                <div class="flex items-center justify-between">
                    <h3
                        class="text-lg font-semibold text-gray-900 dark:text-gray-100 flex items-center"
                    >
                        <span
                            class="w-3 h-3 bg-amber-500 rounded-full mr-2"
                        ></span>
                        All Warning Expiry Items (31-90 days)
                    </h3>
                    <button
                        @click="showWarningModal = false"
                        class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
                    >
                        ✕
                    </button>
                </div>
            </div>
            <div class="p-6 overflow-y-auto max-h-[60vh]">
                <div class="space-y-3">
                    <div
                        v-for="item in allWarningItems"
                        :key="item.id"
                        class="flex items-center justify-between p-3 bg-amber-50 dark:bg-amber-900/20 border border-amber-200 dark:border-amber-800 rounded-lg"
                    >
                        <div class="flex-1">
                            <div
                                class="text-sm font-medium text-gray-900 dark:text-gray-100"
                            >
                                {{ item.name }}
                            </div>
                            <div
                                class="text-xs text-gray-600 dark:text-gray-400"
                            >
                                Qty: {{ item.quantity }} | Expires:
                                {{ item.expiryDate }}
                            </div>
                        </div>
                        <div class="text-right">
                            <div
                                class="text-sm font-bold text-amber-600 dark:text-amber-400"
                            >
                                {{ item.daysUntilExpiry }} days
                            </div>
                        </div>
                    </div>
                </div>
                <div
                    v-if="allWarningItems.length === 0"
                    class="text-center py-8 text-gray-500 dark:text-gray-400"
                >
                    No items expiring within 31-90 days
                </div>
            </div>
            <div class="p-6 border-t border-gray-200 dark:border-gray-700">
                <button
                    @click="
                        copyExpiringItems(allWarningItems, 'Warning');
                        showWarningModal = false;
                    "
                    class="w-full bg-amber-600 hover:bg-amber-700 text-white font-medium py-2 px-4 rounded-lg transition-colors"
                >
                    📋 Copy All Warning Items
                </button>
            </div>
        </div>
    </div>
    <!-- Low Stock Modal -->
    <div
        v-if="showLowStockModal"
        class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4"
    >
        <div
            class="bg-white dark:bg-gray-800 rounded-lg shadow-xl max-w-2xl w-full max-h-[80vh] overflow-hidden"
        >
            <div class="p-6 border-b border-gray-200 dark:border-gray-700">
                <div class="flex items-center justify-between">
                    <h3
                        class="text-lg font-semibold text-gray-900 dark:text-gray-100 flex items-center"
                    >
                        <span
                            class="w-3 h-3 bg-red-500 rounded-full mr-2"
                        ></span>
                        All Low Stock Items
                    </h3>
                    <button
                        @click="showLowStockModal = false"
                        class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
                    >
                        ✕
                    </button>
                </div>
            </div>
            <div class="p-6 overflow-y-auto max-h-[60vh]">
                <div class="space-y-3">
                    <div
                        v-for="item in lowStockItems"
                        :key="item.id"
                        class="flex items-center justify-between p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg"
                    >
                        <div class="flex-1">
                            <div
                                class="text-sm font-medium text-gray-900 dark:text-gray-100"
                            >
                                {{ item.name }}
                            </div>
                            <div
                                class="text-xs text-gray-600 dark:text-gray-400"
                            >
                                Current: {{ item.currentStock }} | Reorder at:
                                {{ item.reorderLevel }}
                            </div>
                        </div>
                        <div class="text-right">
                            <div
                                class="text-sm font-bold text-red-600 dark:text-red-400"
                            >
                                {{
                                    item.reorderLevel - item.currentStock
                                }}
                                needed
                            </div>
                        </div>
                    </div>
                </div>
                <div
                    v-if="lowStockItems.length === 0"
                    class="text-center py-8 text-gray-500 dark:text-gray-400"
                >
                    No items are currently low on stock
                </div>
            </div>
            <div class="p-6 border-t border-gray-200 dark:border-gray-700">
                <button
                    @click="
                        copyLowStockItems(lowStockItems);
                        showLowStockModal = false;
                    "
                    class="w-full bg-red-600 hover:bg-red-700 text-white font-medium py-2 px-4 rounded-lg transition-colors"
                >
                    📋 Copy All Low Stock Items
                </button>
            </div>
        </div>
    </div>
</template>

<script setup>
import { ref, onMounted, computed, watch } from "vue";

const dashboardData = ref({});
const loading = ref(true);
const error = ref("");
const categoryChart = ref(null);
const salesTrendChart = ref(null);

// Sample data for charts (fallback if no data)
const salesTrendData = ref([
    { day: "Mon", sales: 0 },
    { day: "Tue", sales: 0 },
    { day: "Wed", sales: 0 },
    { day: "Thu", sales: 0 },
    { day: "Fri", sales: 0 },
    { day: "Sat", sales: 0 },
    { day: "Sun", sales: 0 },
]);

const categoryData = ref([{ name: "No Data", value: 100 }]);

const lowStockItems = ref([]);
const expiringItems = ref([]);

const maxSales = computed(() => {
    return Math.max(...salesTrendData.value.map((item) => item.sales));
});

const criticalExpiringItems = computed(() => {
    return expiringItems.value
        .filter((i) => i.daysUntilExpiry <= 30)
        .slice(0, 5);
});

const warningExpiringItems = computed(() => {
    return expiringItems.value
        .filter((i) => i.daysUntilExpiry > 30 && i.daysUntilExpiry <= 90)
        .slice(0, 5);
});

const allCriticalItems = computed(() => {
    return expiringItems.value.filter((i) => i.daysUntilExpiry <= 30);
});

const allWarningItems = computed(() => {
    return expiringItems.value.filter(
        (i) => i.daysUntilExpiry > 30 && i.daysUntilExpiry <= 90,
    );
});

onMounted(() => {
    fetchDashboardData();
    
    // Add Escape key listener to close modals
    const handleEscape = (e) => {
        if (e.key === 'Escape') {
            showCriticalModal.value = false;
            showWarningModal.value = false;
            showLowStockModal.value = false;
        }
    };
    
    document.addEventListener('keydown', handleEscape);
    
    // Cleanup on unmount
    return () => {
        document.removeEventListener('keydown', handleEscape);
    };
});

// Register Chart.js plugin for "No Data" message
const noDataPlugin = {
    id: "noDataMessage",
    beforeDraw: function (chart) {
        const { ctx, width, height } = chart;
        const dataset = chart.data.datasets[0];

        // Check if all data is zero or empty
        const hasNoData =
            !dataset || dataset.data.every((value) => value === 0);

        if (hasNoData) {
            ctx.save();
            ctx.font = "14px sans-serif";
            ctx.fillStyle = "rgb(107, 114, 128)";
            ctx.textAlign = "center";
            ctx.textBaseline = "middle";
            ctx.fillText("No Sales Data Available", width / 2, height / 2);
            ctx.restore();

            // Stop drawing the chart
            return false;
        }
        return true;
    },
};

// Register the plugin
Chart.register(noDataPlugin);

// Watch for data changes and initialize charts when data is available
watch(
    [dashboardData, categoryData, salesTrendData],
    () => {
        // Small delay to ensure DOM is ready
        setTimeout(() => {
            // Initialize category chart when data is available
            if (categoryChart.value && categoryData.value.length > 0) {
                console.log(
                    "Initializing category chart with data:",
                    categoryData.value,
                );
                const ctx = categoryChart.value.getContext("2d");

                // Destroy existing chart if it exists
                if (categoryChart.value.chart) {
                    categoryChart.value.chart.destroy();
                }

                // Check if there's real data (not just "No Data")
                const hasRealData = categoryData.value.some(
                    (item) => item.name !== "No Data",
                );

                categoryChart.value.chart = new Chart(ctx, {
                    type: "pie",
                    data: {
                        labels: categoryData.value.map((item) => item.name),
                        datasets: [
                            {
                                data: categoryData.value.map(
                                    (item) => item.value,
                                ),
                                backgroundColor: [
                                    "rgb(37, 99, 235)",
                                    "rgb(34, 197, 94)",
                                    "rgb(168, 85, 247)",
                                    "rgb(251, 146, 60)",
                                    "rgb(107, 114, 128)",
                                ],
                                borderColor: "#ffffff",
                                borderWidth: 2,
                            },
                        ],
                    },
                    options: {
                        responsive: true,
                        maintainAspectRatio: false,
                        plugins: {
                            legend: {
                                display: true,
                                position: "right",
                                labels: {
                                    boxWidth: 12,
                                    padding: 15,
                                    font: {
                                        size: 11,
                                    },
                                },
                            },
                            tooltip: {
                                callbacks: {
                                    label: function (context) {
                                        return (
                                            context.label +
                                            ": " +
                                            context.parsed +
                                            "%"
                                        );
                                    },
                                },
                            },
                        },
                    },
                });
            }

            // Initialize sales trend chart when data is available
            if (
                salesTrendChart.value &&
                salesTrendData.value &&
                salesTrendData.value.length > 0
            ) {
                console.log(
                    "Initializing sales trend chart with data:",
                    salesTrendData.value,
                );
                const ctx = salesTrendChart.value.getContext("2d");

                // Destroy existing chart if it exists
                if (salesTrendChart.value.chart) {
                    salesTrendChart.value.chart.destroy();
                }

                salesTrendChart.value.chart = new Chart(ctx, {
                    type: "line",
                    data: {
                        labels: salesTrendData.value.map((item) => item.day),
                        datasets: [
                            {
                                label: "Sales Trend",
                                data: salesTrendData.value.map(
                                    (item) => item.sales,
                                ),
                                borderColor: "rgb(37, 99, 235)",
                                backgroundColor: "rgba(37, 99, 235, 0.1)",
                                pointBackgroundColor: "rgb(37, 99, 235)",
                                pointBorderColor: "#ffffff",
                                pointBorderWidth: 2,
                                pointRadius: 4,
                                pointHoverRadius: 6,
                                tension: 0.3,
                                fill: true,
                            },
                        ],
                    },
                    options: {
                        responsive: true,
                        maintainAspectRatio: false,
                        plugins: {
                            legend: {
                                display: false,
                            },
                            tooltip: {
                                callbacks: {
                                    label: function (context) {
                                        return (
                                            "Sales: ₦" +
                                            context.parsed.y.toLocaleString()
                                        );
                                    },
                                },
                            },
                        },
                        scales: {
                            y: {
                                beginAtZero: true,
                                ticks: {
                                    callback: function (value) {
                                        return "₦" + value.toLocaleString();
                                    },
                                },
                            },
                        },
                    },
                });
            }
        }, 100); // Small delay to ensure DOM is ready
    },
    { deep: true },
);

const showToast = ref(false);
const toastMessage = ref("");
const showCriticalModal = ref(false);
const showWarningModal = ref(false);
const showLowStockModal = ref(false);

const showCopyToast = (message) => {
    toastMessage.value = message;
    showToast.value = true;
    setTimeout(() => {
        showToast.value = false;
    }, 3000);
};

const openLowStockModal = () => {
    showLowStockModal.value = true;
};

const openCriticalModal = () => {
    showCriticalModal.value = true;
};

const openWarningModal = () => {
    showWarningModal.value = true;
};

const navigateToInventory = () => {
    window.location.href = "/inventory/items";
};

const navigateToSalesHistory = () => {
    window.location.href = "/sales/history";
};

const fetchDashboardData = async () => {
    try {
        loading.value = true;
        error.value = "";

        const response = await fetch("/api/dashboard");
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        const data = await response.json();
        dashboardData.value = data;
        console.log("Dashboard data:", data);

        // Update chart data if available from API
        if (
            data.sales_trend &&
            Array.isArray(data.sales_trend) &&
            data.sales_trend.length > 0
        ) {
            salesTrendData.value = data.sales_trend.map((item) => ({
                day: item.day,
                sales: item.sales / 100, // Convert from kobo to naira
            }));
        }

        if (
            data.category_sales &&
            Array.isArray(data.category_sales) &&
            data.category_sales.length > 0
        ) {
            categoryData.value = data.category_sales.map((item) => ({
                name: item.category,
                value: item.sales,
            }));
        }

        if (data.expiring_items && Array.isArray(data.expiring_items)) {
            expiringItems.value = data.expiring_items.map((item) => ({
                id: item.id,
                name: item.product_name,
                quantity: item.quantity,
                expiryDate: new Date(item.expiry_date).toLocaleDateString(),
                daysUntilExpiry: item.days_until_expiry,
            }));
        }

        if (data.low_stock_items && Array.isArray(data.low_stock_items)) {
            lowStockItems.value = data.low_stock_items.map((item) => ({
                id: item.id,
                name: item.product_name,
                currentStock: item.current_stock,
                reorderLevel: item.reorder_level,
            }));
        }
    } catch (err) {
        console.error("Error fetching dashboard data:", err);
        error.value = "Failed to load dashboard data. Please try again.";
    } finally {
        loading.value = false;
    }
};

const formatCurrency = (amount) => {
    // Convert from kobo to naira
    const naira = amount / 100;
    return new Intl.NumberFormat("en-NG", {
        style: "currency",
        currency: "NGN",
    }).format(naira);
};

const getKPIValue = (key, defaultValue) => {
    if (
        dashboardData.value?.kpi &&
        typeof dashboardData.value.kpi === "object"
    ) {
        return dashboardData.value.kpi[key] || defaultValue;
    }
    return defaultValue;
};

const getCategoryColor = (index) => {
    const colors = [
        "rgb(37, 99, 235)",
        "rgb(34, 197, 94)",
        "rgb(168, 85, 247)",
        "rgb(251, 146, 60)",
        "rgb(107, 114, 128)",
    ];
    return colors[index % colors.length];
};

const copyExpiringItems = (items, sectionName) => {
    const itemsText = items
        .map(
            (item) =>
                `${item.name} - Qty: ${item.quantity}, Expires: ${item.expiryDate} (${item.daysUntilExpiry} days)`,
        )
        .join("\n");

    const header = `${sectionName} Expiring Items:\n${"=".repeat(50)}\n`;
    const footer = `\n${"=".repeat(50)}\nTotal: ${items.length} items`;
    const fullText = header + itemsText + footer;

    navigator.clipboard
        .writeText(fullText)
        .then(() => {
            showCopyToast(
                `Copied ${items.length} ${sectionName.toLowerCase()} expiring items to clipboard`,
            );
        })
        .catch((err) => {
            showCopyToast("Failed to copy to clipboard");
            console.error("Failed to copy text: ", err);
        });
};

const copyLowStockItems = (items) => {
    const itemsText = items
        .map((item) => `${item.name}, Current Stock: ${item.currentStock}`)
        .join("\n");

    const header = `Low Stock Items:\n${"=".repeat(50)}\n`;
    const footer = `\n${"=".repeat(50)}\nTotal: ${items.length} items need reordering`;
    const fullText = header + itemsText + footer;

    navigator.clipboard
        .writeText(fullText)
        .then(() => {
            showCopyToast(
                `Copied ${items.length} low stock items to clipboard`,
            );
        })
        .catch((err) => {
            showCopyToast("Failed to copy to clipboard");
            console.error("Failed to copy text: ", err);
        });
};

const getTrendIcon = (trend) => {
    if (trend > 5) return "📈";
    if (trend < -5) return "📉";
    return "";
};

const getTrendColor = (trend) => {
    if (trend > 5) return "text-green-600";
    if (trend < -5) return "text-red-600";
    return "text-gray-600";
};
</script>
