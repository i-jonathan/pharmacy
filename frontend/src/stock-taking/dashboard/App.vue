<template>
    <div class="max-w-6xl mx-auto px-4 py-8">
        <!-- Header -->
        <div class="flex items-center justify-between mb-6">
            <h1 class="text-2xl font-bold text-gray-800 dark:text-gray-100">
                Stock Taking
            </h1>
            <button
                @click="showCreateModal = true"
                class="inline-flex items-center gap-2 bg-emerald-600 hover:bg-emerald-700 text-white font-medium px-5 py-2.5 rounded-lg shadow transition"
            >
                <span class="text-lg leading-none">+</span>
                New Stock Taking
            </button>
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
            v-else-if="stockTakings.length === 0"
            class="text-center py-16 text-gray-500 dark:text-gray-400"
        >
            <p class="text-lg mb-2">No stock taking sessions yet.</p>
            <p class="text-sm">Click "New Stock Taking" to get started.</p>
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
                            <th class="px-6 py-3">Name</th>
                            <th class="px-6 py-3">Status</th>
                            <th class="px-6 py-3">Created By</th>
                            <th class="px-6 py-3">Started</th>
                            <th class="px-6 py-3">Completed</th>
                            <th class="px-6 py-3">Completed By</th>
                        </tr>
                    </thead>
                    <tbody
                        class="divide-y divide-gray-100 dark:divide-gray-700"
                    >
                        <tr
                            v-for="st in stockTakings"
                            :key="st.id"
                            @click="goToStockTaking(st.id)"
                            class="hover:bg-gray-50 dark:hover:bg-gray-700/40 cursor-pointer transition-colors"
                        >
                            <td
                                class="px-6 py-4 font-medium text-gray-900 dark:text-gray-100"
                            >
                                {{ st.name }}
                            </td>
                            <td class="px-6 py-4">
                                <span :class="statusBadgeClass(st.status)">
                                    {{ st.status }}
                                </span>
                            </td>
                            <td
                                class="px-6 py-4 text-gray-600 dark:text-gray-300"
                            >
                                {{ st.created_by }}
                            </td>
                            <td
                                class="px-6 py-4 text-gray-600 dark:text-gray-300"
                            >
                                {{ formatDate(st.started_at) }}
                            </td>
                            <td
                                class="px-6 py-4 text-gray-600 dark:text-gray-300"
                            >
                                {{
                                    st.completed_at
                                        ? formatDate(st.completed_at)
                                        : "—"
                                }}
                            </td>
                            <td
                                class="px-6 py-4 text-gray-600 dark:text-gray-300"
                            >
                                {{ st.completed_by || "—" }}
                            </td>
                        </tr>
                    </tbody>
                </table>
            </div>
        </div>

        <!-- Create Modal -->
        <div
            v-if="showCreateModal"
            class="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
            @click.self="closeModal"
        >
            <div
                class="bg-white dark:bg-gray-800 rounded-xl shadow-xl w-full max-w-md mx-4 p-6"
            >
                <h2
                    class="text-lg font-semibold text-gray-800 dark:text-gray-100 mb-4"
                >
                    New Stock Taking
                </h2>

                <div
                    v-if="createError"
                    class="mb-4 p-3 bg-red-50 dark:bg-red-900/30 border border-red-200 dark:border-red-700 rounded-lg text-red-700 dark:text-red-300 text-sm"
                >
                    {{ createError }}
                </div>

                <label
                    class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
                >
                    Name
                </label>
                <input
                    ref="nameInput"
                    v-model="newName"
                    @keydown.enter="handleCreate"
                    type="text"
                    placeholder="e.g. April 2026 Stock Count"
                    class="w-full px-4 py-2.5 border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 rounded-lg focus:outline-none focus:ring-2 focus:ring-emerald-500"
                />

                <div class="flex justify-end gap-3 mt-6">
                    <button
                        @click="closeModal"
                        class="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 rounded-lg transition"
                    >
                        Cancel
                    </button>
                    <button
                        @click="handleCreate"
                        :disabled="creating"
                        class="px-5 py-2 text-sm font-medium text-white bg-emerald-600 hover:bg-emerald-700 disabled:opacity-50 disabled:cursor-not-allowed rounded-lg shadow transition"
                    >
                        {{ creating ? "Creating..." : "Create" }}
                    </button>
                </div>
            </div>
        </div>
    </div>
</template>

<script>
export default {
    data() {
        return {
            stockTakings: [],
            loading: true,
            fetchError: null,
            showCreateModal: false,
            newName: "",
            creating: false,
            createError: null,
        };
    },

    mounted() {
        this.fetchStockTakings();
    },

    watch: {
        showCreateModal(val) {
            if (val) {
                this.$nextTick(() => {
                    this.$refs.nameInput?.focus();
                });
            }
        },
    },

    methods: {
        async fetchStockTakings() {
            this.loading = true;
            this.fetchError = null;
            try {
                const res = await fetch("/stock-taking/api/list");
                if (!res.ok) {
                    const err = await res.json();
                    throw new Error(
                        err.message || "Failed to load stock takings",
                    );
                }
                const data = await res.json();
                this.stockTakings = data.stock_takings || [];
            } catch (err) {
                this.fetchError = err.message;
            } finally {
                this.loading = false;
            }
        },

        async handleCreate() {
            const name = this.newName.trim();
            if (!name) return;

            this.creating = true;
            this.createError = null;
            try {
                const res = await fetch("/stock-taking/api/create", {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({ name }),
                });

                const data = await res.json();

                if (!res.ok) {
                    throw new Error(
                        data.message || "Failed to create stock taking",
                    );
                }

                window.location.href = `/stock-taking/${data.id}`;
            } catch (err) {
                this.createError = err.message;
            } finally {
                this.creating = false;
            }
        },

        closeModal() {
            this.showCreateModal = false;
            this.newName = "";
            this.createError = null;
        },

        goToStockTaking(id) {
            window.location.href = `/stock-taking/${id}`;
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

        statusBadgeClass(status) {
            const base =
                "inline-block px-2.5 py-0.5 text-xs font-semibold rounded-full";
            switch (status.toLowerCase()) {
                case "in progress":
                    return `${base} bg-amber-100 text-amber-800 dark:bg-amber-900/40 dark:text-amber-300`;
                case "completed":
                    return `${base} bg-emerald-100 text-emerald-800 dark:bg-emerald-900/40 dark:text-emerald-300`;
                case "cancelled":
                    return `${base} bg-gray-100 text-gray-600 dark:bg-gray-700 dark:text-gray-400`;
                default:
                    return `${base} bg-gray-100 text-gray-600 dark:bg-gray-700 dark:text-gray-400`;
            }
        },
    },
};
</script>
