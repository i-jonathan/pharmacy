import { ref } from "vue";

const API_BASE = "/api";

export function useDashboard() {
  const loading = ref(true);
  const error = ref("");
  const dashboardData = ref({});
  const dateFilter = ref("today");

  function getDateRange(filter) {
    const now = new Date();
    const fmt = (d) => d.toISOString().slice(0, 10);

    switch (filter) {
      case "today":
        return { start_date: fmt(now), end_date: fmt(now) };
      case "week": {
        const start = new Date(now);
        start.setDate(now.getDate() - 7);
        return { start_date: fmt(start), end_date: fmt(now) };
      }
      case "month": {
        const start = new Date(now);
        start.setMonth(now.getMonth() - 1);
        return { start_date: fmt(start), end_date: fmt(now) };
      }
      case "year": {
        const start = new Date(now);
        start.setFullYear(now.getFullYear() - 1);
        return { start_date: fmt(start), end_date: fmt(now) };
      }
      default:
        return { start_date: fmt(now), end_date: fmt(now) };
    }
  }

  async function fetchDashboard(filter = "today") {
    loading.value = true;
    error.value = "";
    dateFilter.value = filter;

    try {
      const range = getDateRange(filter);
      const params = new URLSearchParams(range);
      const response = await fetch(`${API_BASE}/dashboard?${params}`);

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      dashboardData.value = await response.json();
    } catch (err) {
      console.error("Error fetching dashboard data:", err);
      error.value = "Failed to load dashboard data. Please try again.";
    } finally {
      loading.value = false;
    }
  }

  function getKPIValue(key, defaultValue = 0) {
    const kpi = dashboardData.value?.kpi;
    if (kpi && typeof kpi === "object") {
      return kpi[key] ?? defaultValue;
    }
    return defaultValue;
  }

  return {
    loading,
    error,
    dashboardData,
    dateFilter,
    fetchDashboard,
    getKPIValue,
  };
}
