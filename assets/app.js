function initTheme() {
  const getStoredTheme = () => {
    return localStorage.getItem("theme") || "auto";
  };

  const getEffectiveTheme = (theme) => {
    if (theme === "auto") {
      return window.matchMedia("(prefers-color-scheme: dark)").matches
        ? "dark"
        : "light";
    }
    return theme;
  };

  const applyTheme = (theme) => {
    const effectiveTheme = getEffectiveTheme(theme);
    if (effectiveTheme === "dark") {
      document.documentElement.setAttribute("data-theme", "dark");
    } else {
      document.documentElement.removeAttribute("data-theme");
    }
  };

  // Apply theme on page load
  applyTheme(getStoredTheme());

  // Listen for system color scheme changes
  const colorSchemeQuery = window.matchMedia("(prefers-color-scheme: dark)");
  colorSchemeQuery.addEventListener("change", () => {
    if (getStoredTheme() === "auto") {
      applyTheme("auto");
    }
  });

  // Theme toggle functionality
  const themeToggle = document.getElementById("theme-toggle");
  const themeIcon = themeToggle?.querySelector(".theme-icon");

  if (!themeToggle) return;

  const updateToggleButton = (theme) => {
    if (theme === "dark") {
      themeToggle.innerHTML = "🌙";
      themeToggle.setAttribute("aria-label", "Switch to auto mode");
    } else if (theme === "light") {
      themeToggle.innerHTML = "☀️";
      themeToggle.setAttribute("aria-label", "Switch to dark mode");
    } else {
      themeToggle.innerHTML = "🌗";
      themeToggle.setAttribute("aria-label", "Switch to light mode");
    }
  };

  updateToggleButton(getStoredTheme());

  themeToggle.addEventListener("click", () => {
    const currentTheme = getStoredTheme();
    let newTheme;
    if (currentTheme === "light") {
      newTheme = "dark";
    } else if (currentTheme === "dark") {
      newTheme = "auto";
    } else {
      newTheme = "light";
    }
    localStorage.setItem("theme", newTheme);
    applyTheme(newTheme);
    updateToggleButton(newTheme);
  });
}

function initHamburgerMenu() {
  const hamburger = document.querySelector(".hamburger");
  const mainNav = document.querySelector(".main-nav");

  if (!hamburger || !mainNav) return;

  hamburger.addEventListener("click", () => {
    mainNav.classList.toggle("nav-open");
  });
}

window.addEventListener("DOMContentLoaded", () => {
  initTheme();
  initHamburgerMenu();
});
