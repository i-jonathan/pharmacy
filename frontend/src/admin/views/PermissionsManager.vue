<template>
    <div>
        <div class="flex items-center justify-between mb-6">
            <h2 class="text-xl font-bold text-gray-800 dark:text-gray-100">Permission Management</h2>
            <button
                @click="showCreateModal = true"
                class="inline-flex items-center gap-2 bg-emerald-600 hover:bg-emerald-700 text-white font-medium px-4 py-2 rounded-lg shadow transition"
            >
                <span>+</span> Add Permission
            </button>
        </div>

        <div v-if="loading" class="text-center py-12 text-gray-500 dark:text-gray-400">Loading...</div>

        <div v-else-if="error" class="text-center py-12 text-red-500 dark:text-red-400">
            <p class="mb-4">{{ error }}</p>
            <button @click="fetchData" class="bg-emerald-600 hover:bg-emerald-700 text-white px-4 py-2 rounded-lg">Retry</button>
        </div>

        <div v-else-if="permissions.length === 0" class="text-center py-12 text-gray-500 dark:text-gray-400">
            <p class="text-lg mb-2">No permissions defined.</p>
            <p class="text-sm">Create one to get started.</p>
        </div>

        <div v-else class="bg-white dark:bg-gray-800 rounded-xl shadow overflow-hidden border border-gray-200 dark:border-gray-700">
            <table class="w-full text-sm text-left">
                <thead class="text-xs uppercase tracking-wider text-gray-500 dark:text-gray-400 bg-gray-50 dark:bg-gray-700/50">
                    <tr>
                        <th class="px-6 py-3">Resource</th>
                        <th class="px-6 py-3">Action</th>
                        <th class="px-6 py-3">Assigned Roles</th>
                        <th class="px-6 py-3 w-20"></th>
                    </tr>
                </thead>
                <tbody class="divide-y divide-gray-100 dark:divide-gray-700">
                    <tr v-for="perm in permissions" :key="perm.id" class="hover:bg-gray-50 dark:hover:bg-gray-700/40">
                        <td class="px-6 py-4 font-medium text-gray-900 dark:text-gray-100">{{ perm.resource }}</td>
                        <td class="px-6 py-4 text-gray-600 dark:text-gray-300">{{ perm.action }}</td>
                        <td class="px-6 py-4">
                            <div class="flex flex-wrap gap-1 items-center">
                                <span
                                    v-for="role in getAssignedRoles(perm.id)"
                                    :key="role.id"
                                    class="inline-flex items-center gap-1 px-2 py-0.5 text-xs rounded-full bg-emerald-100 dark:bg-emerald-900/40 text-emerald-700 dark:text-emerald-300"
                                >
                                    {{ role.name }}
                                    <button
                                        @click="removeRole(perm.id, role.id)"
                                        class="hover:text-red-500 transition-colors leading-none"
                                        title="Remove from role"
                                    >&times;</button>
                                </span>
                                <button
                                    v-if="unassignedRoles(perm.id).length > 0"
                                    @click="openAssign(perm)"
                                    class="inline-flex items-center px-2 py-0.5 text-xs rounded-full border border-dashed border-gray-300 dark:border-gray-600 text-gray-500 dark:text-gray-400 hover:border-emerald-400 hover:text-emerald-600 transition-colors"
                                >+ Assign</button>
                            </div>
                        </td>
                        <td class="px-6 py-4 text-right">
                            <button
                                @click="confirmDelete(perm)"
                                class="text-red-500 hover:text-red-700 text-sm transition-colors"
                            >Delete</button>
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>

        <!-- Create Permission Modal -->
        <div v-if="showCreateModal" class="fixed inset-0 z-50 flex items-center justify-center">
            <div class="absolute inset-0 bg-black/40" @click="showCreateModal = false"></div>
            <div class="relative bg-white dark:bg-gray-800 rounded-xl shadow-xl p-6 w-full max-w-md mx-4">
                <h3 class="text-lg font-semibold text-gray-800 dark:text-gray-100 mb-4">Add Permission</h3>
                <form @submit.prevent="createPermission">
                    <div class="mb-4">
                        <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Resource</label>
                        <input
                            v-model="newResource"
                            type="text"
                            required
                            placeholder="e.g. Reports"
                            class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-emerald-500 text-gray-900 dark:text-gray-100"
                        />
                    </div>
                    <div class="mb-4">
                        <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Action</label>
                        <input
                            v-model="newAction"
                            type="text"
                            required
                            placeholder="e.g. View"
                            class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-emerald-500 text-gray-900 dark:text-gray-100"
                        />
                    </div>
                    <div v-if="createError" class="mb-4 text-sm text-red-500">{{ createError }}</div>
                    <div class="flex justify-end gap-3">
                        <button
                            type="button"
                            @click="showCreateModal = false"
                            class="px-4 py-2 text-sm text-gray-600 dark:text-gray-300 hover:text-gray-800 transition-colors"
                        >Cancel</button>
                        <button
                            type="submit"
                            :disabled="creating"
                            class="px-4 py-2 text-sm bg-emerald-600 hover:bg-emerald-700 text-white rounded-lg disabled:opacity-50 transition"
                        >{{ creating ? "Creating..." : "Create" }}</button>
                    </div>
                </form>
            </div>
        </div>

        <!-- Assign Role Modal -->
        <div v-if="assigningPerm" class="fixed inset-0 z-50 flex items-center justify-center">
            <div class="absolute inset-0 bg-black/40" @click="assigningPerm = null"></div>
            <div class="relative bg-white dark:bg-gray-800 rounded-xl shadow-xl p-6 w-full max-w-sm mx-4">
                <h3 class="text-lg font-semibold text-gray-800 dark:text-gray-100 mb-4">
                    Assign "{{ assigningPerm.resource }}:{{ assigningPerm.action }}" to Role
                </h3>
                <div class="space-y-2 mb-4">
                    <button
                        v-for="role in unassignedRoles(assigningPerm.id)"
                        :key="role.id"
                        @click="assignRole(assigningPerm.id, role.id)"
                        class="w-full text-left px-4 py-2 rounded-lg hover:bg-emerald-50 dark:hover:bg-emerald-900/30 text-gray-700 dark:text-gray-200 hover:text-emerald-700 dark:hover:text-emerald-300 transition-colors"
                    >{{ role.name }}</button>
                </div>
                <div class="text-right">
                    <button
                        @click="assigningPerm = null"
                        class="px-4 py-2 text-sm text-gray-500 hover:text-gray-700 transition-colors"
                    >Cancel</button>
                </div>
            </div>
        </div>

        <!-- Delete Confirmation Modal -->
        <div v-if="deletingPerm" class="fixed inset-0 z-50 flex items-center justify-center">
            <div class="absolute inset-0 bg-black/40" @click="deletingPerm = null"></div>
            <div class="relative bg-white dark:bg-gray-800 rounded-xl shadow-xl p-6 w-full max-w-sm mx-4">
                <h3 class="text-lg font-semibold text-gray-800 dark:text-gray-100 mb-2">Delete Permission</h3>
                <p class="text-sm text-gray-500 dark:text-gray-400 mb-4">
                    Are you sure you want to delete "{{ deletingPerm.resource }}:{{ deletingPerm.action }}"?
                    This will remove it from all assigned roles.
                </p>
                <div class="flex justify-end gap-3">
                    <button
                        @click="deletingPerm = null"
                        class="px-4 py-2 text-sm text-gray-600 dark:text-gray-300 hover:text-gray-800 transition-colors"
                    >Cancel</button>
                    <button
                        @click="deletePermission(deletingPerm.id)"
                        :disabled="deleting"
                        class="px-4 py-2 text-sm bg-red-600 hover:bg-red-700 text-white rounded-lg disabled:opacity-50 transition"
                    >{{ deleting ? "Deleting..." : "Delete" }}</button>
                </div>
            </div>
        </div>
    </div>
</template>

<script>
export default {
    name: "PermissionsManager",
    data() {
        return {
            permissions: [],
            roles: [],
            roleAssignments: {},
            loading: true,
            error: null,
            showCreateModal: false,
            newResource: "",
            newAction: "",
            creating: false,
            createError: null,
            assigningPerm: null,
            deletingPerm: null,
            deleting: false,
        };
    },
    mounted() {
        this.fetchData();
    },
    methods: {
        async fetchData() {
            this.loading = true;
            this.error = null;
            try {
                const [permRes, roleRes] = await Promise.all([
                    fetch("/admin/api/permissions"),
                    fetch("/admin/api/roles"),
                ]);
                if (!permRes.ok || !roleRes.ok) throw new Error("Failed to load data");
                this.permissions = await permRes.json();
                this.roles = await roleRes.json();
                this.buildAssignments();
            } catch (e) {
                this.error = e.message || "Failed to load permissions";
            } finally {
                this.loading = false;
            }
        },
        buildAssignments() {
            const map = {};
            for (const perm of this.permissions) {
                map[perm.id] = [];
            }
            for (const role of this.roles) {
                for (const p of role.permissions || []) {
                    if (map[p.id]) {
                        map[p.id].push({ id: role.id, name: role.name });
                    }
                }
            }
            this.roleAssignments = map;
        },
        getAssignedRoles(permId) {
            return this.roleAssignments[permId] || [];
        },
        unassignedRoles(permId) {
            const assigned = this.getAssignedRoles(permId).map((r) => r.id);
            return this.roles.filter((r) => !assigned.includes(r.id));
        },
        async createPermission() {
            this.creating = true;
            this.createError = null;
            try {
                const res = await fetch("/admin/api/permissions", {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({ resource: this.newResource, action: this.newAction }),
                });
                if (!res.ok) {
                    const data = await res.json();
                    throw new Error(data.error || "Failed to create");
                }
                this.showCreateModal = false;
                this.newResource = "";
                this.newAction = "";
                await this.fetchData();
            } catch (e) {
                this.createError = e.message;
            } finally {
                this.creating = false;
            }
        },
        openAssign(perm) {
            this.assigningPerm = perm;
        },
        async assignRole(permId, roleId) {
            try {
                const res = await fetch(`/admin/api/permissions/${permId}/role/${roleId}`, { method: "POST" });
                if (!res.ok) throw new Error("Failed to assign");
                this.assigningPerm = null;
                await this.fetchData();
            } catch (e) {
                alert(e.message);
            }
        },
        async removeRole(permId, roleId) {
            try {
                const res = await fetch(`/admin/api/permissions/${permId}/role/${roleId}`, { method: "DELETE" });
                if (!res.ok) throw new Error("Failed to remove");
                await this.fetchData();
            } catch (e) {
                alert(e.message);
            }
        },
        confirmDelete(perm) {
            this.deletingPerm = perm;
        },
        async deletePermission(permId) {
            this.deleting = true;
            try {
                const res = await fetch(`/admin/api/permissions/${permId}`, { method: "DELETE" });
                if (!res.ok) throw new Error("Failed to delete");
                this.deletingPerm = null;
                await this.fetchData();
            } catch (e) {
                alert(e.message);
            } finally {
                this.deleting = false;
            }
        },
    },
};
</script>
