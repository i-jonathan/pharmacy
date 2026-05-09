<template>
    <div class="max-w-6xl mx-auto px-4 py-8">
        <div v-if="loading" class="text-center py-16 text-gray-500 dark:text-gray-400">
            Loading...
        </div>

        <div v-else-if="error" class="text-center py-16 text-red-500 dark:text-red-400">
            <p class="mb-4">{{ error }}</p>
            <button @click="fetchModules" class="bg-emerald-600 hover:bg-emerald-700 text-white px-4 py-2 rounded-lg">
                Retry
            </button>
        </div>

        <div v-else-if="!currentModule">
            <h1 class="text-2xl font-bold text-gray-800 dark:text-gray-100 mb-6">
                Administration
            </h1>

            <div v-if="modules.length === 0" class="text-center py-16 text-gray-500 dark:text-gray-400">
                No admin modules available.
            </div>

            <div v-else class="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div
                    v-for="mod in modules"
                    :key="mod.path"
                    @click="currentModule = mod.path"
                    class="bg-white dark:bg-gray-800 rounded-xl shadow border border-gray-200 dark:border-gray-700 p-6 cursor-pointer hover:shadow-lg hover:border-emerald-300 dark:hover:border-emerald-600 transition-all group"
                >
                    <div class="flex items-start gap-4">
                        <div class="w-12 h-12 bg-emerald-100 dark:bg-emerald-900/40 rounded-lg flex items-center justify-center text-emerald-600 dark:text-emerald-400 text-xl group-hover:scale-110 transition-transform">
                            <span v-if="mod.icon === 'shield'">&#128737;</span>
                            <span v-else-if="mod.icon === 'users'">&#128101;</span>
                            <span v-else-if="mod.icon === 'user-cog'">&#128295;</span>
                            <span v-else-if="mod.icon === 'tags'">&#127991;</span>
                            <span v-else>&#128218;</span>
                        </div>
                        <div class="flex-1">
                            <h2 class="text-lg font-semibold text-gray-800 dark:text-gray-100 group-hover:text-emerald-600 dark:group-hover:text-emerald-400 transition-colors">
                                {{ mod.name }}
                            </h2>
                            <p class="text-sm text-gray-500 dark:text-gray-400 mt-1">
                                {{ mod.description }}
                            </p>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <div v-else>
            <button
                @click="currentModule = null"
                class="inline-flex items-center gap-1 text-sm text-gray-500 dark:text-gray-400 hover:text-emerald-600 dark:hover:text-emerald-400 mb-6 transition-colors"
            >
                <span>&larr;</span> Back to Admin Dashboard
            </button>

            <PermissionsManager v-if="currentModule === 'permissions'" />
            <RolesViewer v-else-if="currentModule === 'roles'" />
            <UsersManager v-else-if="currentModule === 'users'" />
            <CategoriesManager v-else-if="currentModule === 'categories'" />
        </div>
    </div>
</template>

<script>
import PermissionsManager from "./views/PermissionsManager.vue";
import RolesViewer from "./views/RolesViewer.vue";
import UsersManager from "./views/UsersManager.vue";
import CategoriesManager from "./views/CategoriesManager.vue";

export default {
    name: "AdminApp",
    components: { PermissionsManager, RolesViewer, UsersManager, CategoriesManager },
    data() {
        return {
            modules: [],
            currentModule: null,
            loading: true,
            error: null,
        };
    },
    mounted() {
        this.fetchModules();
    },
    methods: {
        async fetchModules() {
            this.loading = true;
            this.error = null;
            try {
                const res = await fetch("/admin/api/modules");
                if (!res.ok) throw new Error("Failed to load modules");
                this.modules = await res.json();
            } catch (e) {
                this.error = e.message || "Failed to load admin modules";
            } finally {
                this.loading = false;
            }
        },
    },
};
</script>
