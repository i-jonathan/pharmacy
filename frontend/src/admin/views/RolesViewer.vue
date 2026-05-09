<template>
    <div>
        <h2 class="text-xl font-bold text-gray-800 dark:text-gray-100 mb-6">Roles</h2>

        <div v-if="loading" class="text-center py-12 text-gray-500 dark:text-gray-400">Loading...</div>

        <div v-else-if="error" class="text-center py-12 text-red-500 dark:text-red-400">
            <p class="mb-4">{{ error }}</p>
            <button @click="fetchRoles" class="bg-emerald-600 hover:bg-emerald-700 text-white px-4 py-2 rounded-lg">Retry</button>
        </div>

        <div v-else-if="roles.length === 0" class="text-center py-12 text-gray-500 dark:text-gray-400">
            No roles defined.
        </div>

        <div v-else class="space-y-4">
            <div
                v-for="role in roles"
                :key="role.id"
                class="bg-white dark:bg-gray-800 rounded-xl shadow border border-gray-200 dark:border-gray-700 overflow-hidden"
            >
                <div class="px-6 py-4 flex items-center justify-between bg-gray-50 dark:bg-gray-700/50">
                    <div>
                        <h3 class="text-lg font-semibold text-gray-800 dark:text-gray-100">{{ role.name }}</h3>
                        <p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
                            Created {{ formatDate(role.created_at) }}
                        </p>
                    </div>
                    <span class="text-sm text-gray-500 dark:text-gray-400">
                        {{ (role.permissions || []).length }} permission{{ (role.permissions || []).length !== 1 ? "s" : "" }}
                    </span>
                </div>
                <div class="px-6 py-4">
                    <div v-if="!role.permissions || role.permissions.length === 0" class="text-sm text-gray-400 dark:text-gray-500">
                        No permissions assigned.
                    </div>
                    <div v-else class="flex flex-wrap gap-2">
                        <span
                            v-for="perm in role.permissions"
                            :key="perm.id"
                            class="px-3 py-1 text-xs rounded-full bg-emerald-100 dark:bg-emerald-900/40 text-emerald-700 dark:text-emerald-300 font-medium"
                        >{{ perm.resource }}:{{ perm.action }}</span>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<script>
export default {
    name: "RolesViewer",
    data() {
        return {
            roles: [],
            loading: true,
            error: null,
        };
    },
    mounted() {
        this.fetchRoles();
    },
    methods: {
        async fetchRoles() {
            this.loading = true;
            this.error = null;
            try {
                const res = await fetch("/admin/api/roles");
                if (!res.ok) throw new Error("Failed to load roles");
                this.roles = await res.json();
            } catch (e) {
                this.error = e.message || "Failed to load roles";
            } finally {
                this.loading = false;
            }
        },
        formatDate(d) {
            if (!d) return "—";
            return new Date(d).toLocaleDateString("en-US", { year: "numeric", month: "long", day: "numeric" });
        },
    },
};
</script>
