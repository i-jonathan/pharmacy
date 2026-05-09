<template>
    <div>
        <div class="flex items-center justify-between mb-6">
            <h2 class="text-xl font-bold text-gray-800 dark:text-gray-100">Categories</h2>
            <button
                @click="showCreateModal = true"
                class="inline-flex items-center gap-2 bg-emerald-600 hover:bg-emerald-700 text-white font-medium px-4 py-2 rounded-lg shadow transition"
            >
                <span>+</span> Add Category
            </button>
        </div>

        <div v-if="loading" class="text-center py-12 text-gray-500 dark:text-gray-400">Loading...</div>

        <div v-else-if="error" class="text-center py-12 text-red-500 dark:text-red-400">
            <p class="mb-4">{{ error }}</p>
            <button @click="fetchCategories" class="bg-emerald-600 hover:bg-emerald-700 text-white px-4 py-2 rounded-lg">Retry</button>
        </div>

        <div v-else-if="categories.length === 0" class="text-center py-12 text-gray-500 dark:text-gray-400">
            <p class="text-lg mb-2">No categories.</p>
            <p class="text-sm">Add one to organize your products.</p>
        </div>

        <div v-else class="bg-white dark:bg-gray-800 rounded-xl shadow overflow-hidden border border-gray-200 dark:border-gray-700">
            <table class="w-full text-sm text-left">
                <thead class="text-xs uppercase tracking-wider text-gray-500 dark:text-gray-400 bg-gray-50 dark:bg-gray-700/50">
                    <tr>
                        <th class="px-6 py-3">Name</th>
                        <th class="px-6 py-3">Created</th>
                        <th class="px-6 py-3 w-32">Actions</th>
                    </tr>
                </thead>
                <tbody class="divide-y divide-gray-100 dark:divide-gray-700">
                    <tr v-for="cat in categories" :key="cat.id" class="hover:bg-gray-50 dark:hover:bg-gray-700/40">
                        <td class="px-6 py-4">
                            <div v-if="editingId === cat.id" class="flex gap-2">
                                <input
                                    v-model="editName"
                                    type="text"
                                    class="px-2 py-1 border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 rounded text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500 text-gray-900 dark:text-gray-100 flex-1"
                                    @keyup.enter="saveEdit(cat.id)"
                                    @keyup.escape="editingId = null"
                                />
                                <button @click="saveEdit(cat.id)" class="text-emerald-600 hover:text-emerald-700 text-sm font-medium">Save</button>
                                <button @click="editingId = null" class="text-gray-400 hover:text-gray-600 text-sm">Cancel</button>
                            </div>
                            <span v-else class="font-medium text-gray-900 dark:text-gray-100">{{ cat.name }}</span>
                        </td>
                        <td class="px-6 py-4 text-gray-500 dark:text-gray-400 text-xs">
                            {{ formatDate(cat.created_at) }}
                        </td>
                        <td class="px-6 py-4">
                            <div class="flex gap-3">
                                <button
                                    @click="startEdit(cat)"
                                    class="text-sm text-emerald-600 dark:text-emerald-400 hover:text-emerald-800 dark:hover:text-emerald-300 transition-colors"
                                >Edit</button>
                                <button
                                    @click="confirmDelete(cat)"
                                    class="text-sm text-red-500 hover:text-red-700 transition-colors"
                                >Delete</button>
                            </div>
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>

        <!-- Create Modal -->
        <div v-if="showCreateModal" class="fixed inset-0 z-50 flex items-center justify-center">
            <div class="absolute inset-0 bg-black/40" @click="showCreateModal = false"></div>
            <div class="relative bg-white dark:bg-gray-800 rounded-xl shadow-xl p-6 w-full max-w-md mx-4">
                <h3 class="text-lg font-semibold text-gray-800 dark:text-gray-100 mb-4">Add Category</h3>
                <form @submit.prevent="createCategory">
                    <div class="mb-4">
                        <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Name</label>
                        <input
                            v-model="newName"
                            type="text"
                            required
                            placeholder="Category name"
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

        <!-- Delete Confirmation Modal -->
        <div v-if="deletingCat" class="fixed inset-0 z-50 flex items-center justify-center">
            <div class="absolute inset-0 bg-black/40" @click="deletingCat = null"></div>
            <div class="relative bg-white dark:bg-gray-800 rounded-xl shadow-xl p-6 w-full max-w-sm mx-4">
                <h3 class="text-lg font-semibold text-gray-800 dark:text-gray-100 mb-2">Delete Category</h3>
                <p class="text-sm text-gray-500 dark:text-gray-400 mb-4">
                    Are you sure you want to delete "{{ deletingCat.name }}"?
                    Categories used by products cannot be deleted.
                </p>
                <div v-if="deleteError" class="mb-4 text-sm text-red-500">{{ deleteError }}</div>
                <div class="flex justify-end gap-3">
                    <button
                        @click="deletingCat = null"
                        class="px-4 py-2 text-sm text-gray-600 dark:text-gray-300 hover:text-gray-800 transition-colors"
                    >Cancel</button>
                    <button
                        @click="deleteCategory(deletingCat.id)"
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
    name: "CategoriesManager",
    data() {
        return {
            categories: [],
            loading: true,
            error: null,
            showCreateModal: false,
            newName: "",
            creating: false,
            createError: null,
            editingId: null,
            editName: "",
            deletingCat: null,
            deleting: false,
            deleteError: null,
        };
    },
    mounted() {
        this.fetchCategories();
    },
    methods: {
        async fetchCategories() {
            this.loading = true;
            this.error = null;
            try {
                const res = await fetch("/admin/api/categories");
                if (!res.ok) throw new Error("Failed to load categories");
                this.categories = await res.json();
            } catch (e) {
                this.error = e.message || "Failed to load categories";
            } finally {
                this.loading = false;
            }
        },
        async createCategory() {
            this.creating = true;
            this.createError = null;
            try {
                const res = await fetch("/admin/api/categories", {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({ name: this.newName }),
                });
                if (!res.ok) {
                    const data = await res.json();
                    throw new Error(data.error || "Failed to create");
                }
                this.showCreateModal = false;
                this.newName = "";
                await this.fetchCategories();
            } catch (e) {
                this.createError = e.message;
            } finally {
                this.creating = false;
            }
        },
        startEdit(cat) {
            this.editingId = cat.id;
            this.editName = cat.name;
        },
        async saveEdit(id) {
            try {
                const res = await fetch(`/admin/api/categories/${id}`, {
                    method: "PUT",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({ name: this.editName }),
                });
                if (!res.ok) {
                    const data = await res.json();
                    throw new Error(data.error || "Failed to update");
                }
                this.editingId = null;
                await this.fetchCategories();
            } catch (e) {
                alert(e.message);
            }
        },
        confirmDelete(cat) {
            this.deletingCat = cat;
            this.deleteError = null;
        },
        async deleteCategory(id) {
            this.deleting = true;
            this.deleteError = null;
            try {
                const res = await fetch(`/admin/api/categories/${id}`, { method: "DELETE" });
                if (!res.ok) {
                    const data = await res.json();
                    throw new Error(data.error || "Failed to delete");
                }
                this.deletingCat = null;
                await this.fetchCategories();
            } catch (e) {
                this.deleteError = e.message;
            } finally {
                this.deleting = false;
            }
        },
        formatDate(d) {
            if (!d) return "—";
            return new Date(d).toLocaleDateString("en-US", { year: "numeric", month: "long", day: "numeric" });
        },
    },
};
</script>
