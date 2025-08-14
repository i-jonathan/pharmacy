function createToastContainer() {
  let container = document.getElementById("toast-root");
  if (!container) {
    container = document.createElement("div");
    container.id = "toast-root";
    container.className =
      "fixed bottom-4 right-4 flex flex-col items-end space-y-3 z-50 pointer-events-none";
    document.body.appendChild(container);
  }
  return container;
}

function showToast(message, { type = "info", duration = 3000 } = {}) {
  const container = createToastContainer();

  const baseClasses =
    "pointer-events-auto flex flex-col rounded-xl px-6 py-4 shadow-lg " +
    "border text-lg transition transform duration-300 translate-y-4 opacity-0 " +
    "bg-white dark:bg-gray-800 w-[400px] max-w-[90vw]";

  const typeStyles = {
    success: "border-emerald-300 text-emerald-900 dark:text-emerald-200",
    error: "border-red-300 text-red-900 dark:text-red-200",
    info: "border-gray-300 text-gray-900 dark:text-gray-200",
  };

  const typeIcons = {
    success: "✅",
    error: "⚠️",
    info: "ℹ️",
  };

  const toast = document.createElement("div");
  toast.setAttribute("role", "status");
  toast.setAttribute("aria-live", "polite");
  toast.className = `${baseClasses} ${typeStyles[type] || typeStyles.info}`;
  toast.innerHTML = `
    <div class="flex items-start gap-4">
      <div class="text-2xl leading-none mt-0.5">${typeIcons[type] || typeIcons.info}</div>
      <div class="flex-1">${message}</div>
      <button type="button"
        class="ml-2 rounded-md px-3 py-1.5 text-sm bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600">
        ✖
      </button>
    </div>
    <div class="relative h-1.5 bg-gray-200 dark:bg-gray-700 rounded overflow-hidden mt-3">
      <div class="absolute left-0 top-0 h-full bg-emerald-500 progress-bar"></div>
    </div>
  `;

  const closeBtn = toast.querySelector("button");
  const progressBar = toast.querySelector(".progress-bar");

  container.appendChild(toast);
  requestAnimationFrame(() => {
    toast.classList.remove("translate-y-4", "opacity-0");
    toast.classList.add("translate-y-0", "opacity-100");
  });

  let hideTimer = null;
  let startTime = Date.now();
  let remaining = duration;
  let progressInterval = null;

  function updateProgress() {
    const elapsed = Date.now() - startTime;
    const percent = Math.max(0, ((remaining - elapsed) / duration) * 100);
    progressBar.style.width = `${percent}%`;
  }

  function startTimer() {
    if (duration > 0) {
      startTime = Date.now();
      hideTimer = setTimeout(dismiss, remaining);
      progressInterval = setInterval(updateProgress, 16);
    }
  }

  function clearTimer() {
    clearTimeout(hideTimer);
    clearInterval(progressInterval);
    remaining -= Date.now() - startTime;
  }

  function dismiss() {
    clearTimer();
    toast.classList.remove("opacity-100", "translate-y-0");
    toast.classList.add("opacity-0", "translate-y-4");
    setTimeout(() => toast.remove(), 300);
  }

  toast.addEventListener("mouseenter", clearTimer);
  toast.addEventListener("mouseleave", startTimer);
  closeBtn.addEventListener("click", dismiss);

  startTimer();
  return { dismiss };
}
