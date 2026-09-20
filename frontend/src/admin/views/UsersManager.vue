<template>
    <div>
        <h2 class="text-xl font-bold text-neutral-800 dark:text-neutral-100 mb-6">Users</h2>

        <div v-if="loading" class="text-center py-12 text-neutral-500 dark:text-neutral-400">Loading...</div>

        <div v-else-if="error" class="text-center py-12 text-red-500 dark:text-red-400">
            <p class="mb-4">{{ error }}</p>
            <button @click="fetchData" class="bg-emerald-600 hover:bg-emerald-700 text-white px-4 py-2 rounded-lg">Retry</button>
        </div>

        <div v-else-if="users.length === 0" class="text-center py-12 text-neutral-500 dark:text-neutral-400">
            No users registered.
        </div>

        <div v-else class="bg-white dark:bg-neutral-800 rounded-xl shadow overflow-hidden border border-neutral-200 dark:border-neutral-700">
            <table class="w-full text-sm text-left">
                <thead class="text-xs uppercase tracking-wider text-neutral-500 dark:text-neutral-400 bg-neutral-50 dark:bg-neutral-700/50">
                    <tr>
                        <th class="px-6 py-3">Username</th>
                        <th class="px-6 py-3">Role</th>
                        <th class="px-6 py-3">Created</th>
                        <th class="px-6 py-3 w-40">Actions</th>
                    </tr>
                </thead>
                <tbody class="divide-y divide-neutral-100 dark:divide-neutral-700">
                    <tr v-for="user in users" :key="user.id" class="hover:bg-neutral-50 dark:hover:bg-neutral-700/40">
                        <td class="px-6 py-4 font-medium text-neutral-900 dark:text-neutral-100">{{ user.username }}</td>
                        <td class="px-6 py-4">
                            <select
                                :value="user.role_id"
                                @change="updateRole(user, $event.target.value)"
                                class="px-2 py-1 border border-neutral-300 dark:border-neutral-600 bg-white dark:bg-neutral-700 rounded text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500 text-neutral-900 dark:text-neutral-100"
                            >
                                <option v-for="role in roles" :key="role.id" :value="role.id">{{ role.name }}</option>
                            </select>
                        </td>
                        <td class="px-6 py-4 text-neutral-500 dark:text-neutral-400 text-xs">
                            {{ formatDate(user.created_at) }}
                        </td>
                        <td class="px-6 py-4">
                            <button
                                @click="openResetPassword(user)"
                                class="text-sm text-emerald-600 dark:text-emerald-400 hover:text-emerald-800 dark:hover:text-emerald-300 transition-colors"
                            >Reset Password</button>
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>

        <!-- Reset Password Modal -->
        <div v-if="resetUser" class="fixed inset-0 z-50 flex items-center justify-center">
            <div class="absolute inset-0 bg-black/40" @click="resetUser = null"></div>
            <div class="relative bg-white dark:bg-neutral-800 rounded-xl shadow-xl p-6 w-full max-w-md mx-4">
                <h3 class="text-lg font-semibold text-neutral-800 dark:text-neutral-100 mb-4">
                    Reset Password for {{ resetUser.username }}
                </h3>
                <form @submit.prevent="resetPassword">
                    <div class="mb-4">
                        <label class="block text-sm font-medium text-neutral-700 dark:text-neutral-300 mb-1">New Password</label>
                        <input
                            v-model="newPassword"
                            type="password"
                            required
                            minlength="8"
                            placeholder="At least 8 characters"
                            class="w-full px-3 py-2 border border-neutral-300 dark:border-neutral-600 bg-white dark:bg-neutral-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-emerald-500 text-neutral-900 dark:text-neutral-100"
                        />
                    </div>
                    <div v-if="resetError" class="mb-4 text-sm text-red-500">{{ resetError }}</div>
                    <div class="flex justify-end gap-3">
                        <button
                            type="button"
                            @click="resetUser = null"
                            class="px-4 py-2 text-sm text-neutral-600 dark:text-neutral-300 hover:text-neutral-800 transition-colors"
                        >Cancel</button>
                        <button
                            type="submit"
                            :disabled="resetting"
                            class="px-4 py-2 text-sm bg-emerald-600 hover:bg-emerald-700 text-white rounded-lg disabled:opacity-50 transition"
                        >{{ resetting ? "Resetting..." : "Reset" }}</button>
                    </div>
                </form>
            </div>
        </div>
    </div>
</template>

<script>
export default {
    name: "UsersManager",
    data() {
        return {
            users: [],
            roles: [],
            loading: true,
            error: null,
            resetUser: null,
            newPassword: "",
            resetting: false,
            resetError: null,
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
                const [userRes, roleRes] = await Promise.all([
                    fetch("/admin/api/users"),
                    fetch("/admin/api/roles"),
                ]);
                if (!userRes.ok || !roleRes.ok) throw new Error("Failed to load data");
                this.users = await userRes.json();
                this.roles = await roleRes.json();
            } catch (e) {
                this.error = e.message || "Failed to load users";
            } finally {
                this.loading = false;
            }
        },
        async updateRole(user, newRoleId) {
            const previousRoleId = user.role_id;
            user.role_id = parseInt(newRoleId);
            try {
                const res = await fetch(`/admin/api/users/${user.id}/role`, {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({ role_id: parseInt(newRoleId) }),
                });
                if (!res.ok) throw new Error("Failed to update role");
            } catch (e) {
                user.role_id = previousRoleId;
                alert("Failed to update role: " + e.message);
            }
        },
        openResetPassword(user) {
            this.resetUser = user;
            this.newPassword = "";
            this.resetError = null;
        },
        async resetPassword() {
            this.resetting = true;
            this.resetError = null;
            try {
                const res = await fetch(`/admin/api/users/${this.resetUser.id}/reset-password`, {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({ new_password: this.newPassword }),
                });
                if (!res.ok) {
                    const data = await res.json();
                    throw new Error(data.error || "Failed to reset");
                }
                this.resetUser = null;
                alert("Password reset successfully.");
            } catch (e) {
                this.resetError = e.message;
            } finally {
                this.resetting = false;
            }
        },
        formatDate(d) {
            if (!d) return "—";
            return new Date(d).toLocaleDateString("en-US", { year: "numeric", month: "long", day: "numeric" });
        },
    },
};
</script>
