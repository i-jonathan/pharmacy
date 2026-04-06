<template>
    <div class="bg-white rounded-lg shadow p-6">
        <div class="flex items-center justify-between mb-4">
            <div class="flex items-center">
                <component :is="icon" :class="iconClasses" />
                <div>
                    <h3 class="text-lg font-semibold text-gray-900">
                        {{ title }}
                    </h3>
                    <p class="text-sm text-gray-500">{{ subtitle }}</p>
                </div>
            </div>
            <div class="text-right">
                <div class="text-2xl font-bold" :class="valueClasses">
                    {{ formattedValue }}
                </div>
                <div class="flex items-center text-sm" :class="trendClasses">
                    <component :is="trendIcon" class="h-4 w-4 mr-1" />
                    {{ trend }}
                </div>
            </div>
        </div>
    </div>
</template>

<script setup>
import { computed } from "vue";
import { TrendingUp, TrendingDown } from "lucide-vue-next";

const props = defineProps({
    title: {
        type: String,
        required: true,
    },
    value: {
        type: Number,
        required: true,
    },
    subtitle: {
        type: String,
        default: "",
    },
    trend: {
        type: String,
        default: "",
    },
    trendUp: {
        type: Boolean,
        default: true,
    },
    icon: {
        type: [String, Object],
        required: true,
    },
    color: {
        type: String,
        default: "blue",
        validator: (value) =>
            ["blue", "green", "purple", "red", "yellow", "orange"].includes(
                value,
            ),
    },
});

const formattedValue = computed(() => {
    return props.value.toLocaleString();
});

const trendIcon = computed(() => {
    return props.trendUp ? TrendingUp : TrendingDown;
});

const iconClasses = computed(() => {
    const colorClasses = {
        blue: "h-8 w-8 text-blue-600 mr-3",
        green: "h-8 w-8 text-green-600 mr-3",
        purple: "h-8 w-8 text-purple-600 mr-3",
        red: "h-8 w-8 text-red-600 mr-3",
        yellow: "h-8 w-8 text-yellow-600 mr-3",
        orange: "h-8 w-8 text-orange-600 mr-3",
    };
    return colorClasses[props.color] || colorClasses.blue;
});

const valueClasses = computed(() => {
    const colorClasses = {
        blue: "text-blue-600",
        green: "text-green-600",
        purple: "text-purple-600",
        red: "text-red-600",
        yellow: "text-yellow-600",
        orange: "text-orange-600",
    };
    return colorClasses[props.color] || colorClasses.blue;
});

const trendClasses = computed(() => {
    return props.trendUp ? "text-green-600" : "text-red-600";
});
</script>
